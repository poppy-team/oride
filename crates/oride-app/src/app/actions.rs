//! Execução de comandos, ações de edição, navegação e gerenciamento de abas/arquivos.

use std::path::{Path, PathBuf};
use std::time::{Duration, Instant};

use crossterm::event::KeyEvent;
use oride_core::{DocumentError, DocumentId, Selection};
use oride_fs::{list_files_recursive, ProjectTree};
use oride_git::{
    ahead_behind, current_branch, diff_file, git_pull, git_push, scm_entries, status_map,
};
use oride_keymap::{parse_action, Action, ResolvedKey};
use oride_plugin::{PluginCtx, PluginHook};
use oride_search::{search_project, SearchHit, SearchQuery};
use oride_syntax::{continue_list_on_enter, detect_language, LanguageId};
use oride_terminal::EmbeddedTerminal;
use oride_ui::UiTheme;

use crate::browser::{BrowseMode, PathBrowser};
use crate::clipboard;
use crate::disk_watch::DiskWatch;
use crate::jump_list::Jump;
use crate::split::SplitOrientation;

use super::state::{App, Focus, KeyCommand, Overlay, PromptKind};

struct HostCtx {
    status: Option<String>,
    workspace: PathBuf,
    path: Option<PathBuf>,
    text: String,
    dirty: bool,
}

impl PluginCtx for HostCtx {
    fn set_status(&mut self, message: &str) {
        self.status = Some(message.to_string());
    }
    fn workspace_root(&self) -> &Path {
        &self.workspace
    }
    fn active_path(&self) -> Option<PathBuf> {
        self.path.clone()
    }
    fn active_buffer_text(&self) -> String {
        self.text.clone()
    }
    fn active_is_dirty(&self) -> bool {
        self.dirty
    }
}

impl App {
    #[must_use]
    pub fn map_key(&self, key: KeyEvent) -> Option<KeyCommand> {
        match self.keymap.resolve_event(key)? {
            ResolvedKey::Action(action) => Some(KeyCommand::Action(action)),
            ResolvedKey::InsertChar(character) => Some(KeyCommand::InsertChar(character)),
        }
    }

    pub fn apply(&mut self, command: KeyCommand) {
        let result = self.apply_inner(command);
        if let Err(error) = result {
            self.set_status(format!("error: {error}"));
        }
    }

    pub(crate) fn apply_inner(&mut self, command: KeyCommand) -> Result<(), DocumentError> {
        let document_changed = matches!(
            command,
            KeyCommand::InsertChar(_)
                | KeyCommand::Action(
                    Action::Undo
                        | Action::Redo
                        | Action::InsertNewline
                        | Action::InsertTab
                        | Action::Backspace
                        | Action::Delete
                        | Action::Paste
                        | Action::Cut
                        | Action::ToggleComment
                )
        );
        let refresh_completion = matches!(
            command,
            KeyCommand::InsertChar(character)
                if character == '_' || character.is_alphanumeric()
        );
        match command {
            KeyCommand::InsertChar(character) => {
                if self.focus != Focus::Editor {
                    return Ok(());
                }
                self.quit_confirm_pending = false;
                self.close_tab_confirm = None;
                self.insert_char_or_tab(character)?;
            }
            KeyCommand::Action(action) => self.apply_action(action)?,
        }
        if document_changed {
            self.lsp_sync_active();
        }
        if refresh_completion {
            self.refresh_auto_completion()?;
        }
        self.ensure_cursor_visible();
        Ok(())
    }

    pub(crate) fn refresh_auto_completion(&mut self) -> Result<(), DocumentError> {
        if !self.config.editor.completion_auto {
            return Ok(());
        }
        let language = self.active_language();
        let (replace_start, prefix) = self.active_completion_prefix()?;
        if prefix.chars().count() < self.config.editor.completion_min_chars as usize {
            return Ok(());
        }
        let items = self.offline_completion_choices(language, &prefix);
        if items.is_empty() {
            return Ok(());
        }
        self.overlay = Overlay::Completion {
            items,
            selected: 0,
            replace_start,
        };
        Ok(())
    }

    pub(crate) fn insert_char_or_tab(&mut self, character: char) -> Result<(), DocumentError> {
        if character == '\t' {
            return self.insert_tab();
        }
        let mut encode_buffer = [0u8; 4];
        let encoded_string = character.encode_utf8(&mut encode_buffer);
        self.store.active_mut()?.insert_text(encoded_string)?;
        Ok(())
    }

    pub(crate) fn insert_tab(&mut self) -> Result<(), DocumentError> {
        let text = if self.config.editor.insert_spaces {
            " ".repeat(self.config.editor.tab_size as usize)
        } else {
            "\t".into()
        };
        self.store.active_mut()?.insert_text(&text)?;
        Ok(())
    }

    pub(crate) fn apply_action(&mut self, action: Action) -> Result<(), DocumentError> {
        match action {
            Action::Quit => {
                if self.overlay != Overlay::None {
                    self.overlay = Overlay::None;
                    return Ok(());
                }
                if self.focus != Focus::Editor {
                    self.focus = Focus::Editor;
                    return Ok(());
                }
                let dirty_count = self.store.dirty_count();
                if dirty_count > 0 && !self.quit_confirm_pending {
                    self.quit_confirm_pending = true;
                    self.set_status(format!(
                        "{dirty_count} tab(s) com mudanças — salve ou saia novamente para descartar"
                    ));
                } else {
                    self.should_quit = true;
                }
            }
            Action::Save => {
                let needs_path = self
                    .store
                    .active()
                    .map(|document| document.path().is_none())
                    .unwrap_or(true);
                if needs_path {
                    self.open_save_as_browser();
                    return Ok(());
                }
                let document = self.store.active_mut()?;
                match document.save_to(None) {
                    Ok(()) => {
                        self.quit_confirm_pending = false;
                        if self.config.editor.format_on_save {
                            let _ = self.lsp_format();
                            if let Ok(doc) = self.store.active_mut() {
                                let _ = doc.save_to(None);
                            }
                        }
                        if let Some(path) = self
                            .store
                            .active()
                            .ok()
                            .and_then(|doc| doc.path().map(Path::to_path_buf))
                        {
                            self.notify_saved_path(&path);
                        }
                        self.set_status("saved");
                        self.refresh_git_and_index();
                        self.fire_plugin_hook(PluginHook::OnSave);
                    }
                    Err(DocumentError::Io(error))
                        if error.kind() == std::io::ErrorKind::InvalidInput =>
                    {
                        self.open_save_as_browser();
                    }
                    Err(error) => return Err(error),
                }
            }
            Action::SaveAs => {
                self.open_save_as_browser();
            }
            Action::SaveAll => {
                let report = self.store.save_all();
                for path in &report.saved_paths {
                    self.notify_saved_path(path);
                }
                self.refresh_git_and_index();
                if let Some(failure) = report.failures.first() {
                    self.set_status(format!(
                        "save all: {} ok · {} sem path · {} falha(s): {}: {}",
                        report.saved,
                        report.skipped_without_path,
                        report.failures.len(),
                        failure.path.display(),
                        failure.error
                    ));
                } else {
                    self.set_status(format!(
                        "save all: {} ok · {} sem path",
                        report.saved, report.skipped_without_path
                    ));
                }
            }
            Action::Help => {
                self.overlay = Overlay::Help {
                    query: String::new(),
                    selected: 0,
                };
                self.set_status(format!(
                    "atalhos: {} binds · digite filtra · ↑↓ · Esc",
                    self.keymap.len()
                ));
            }
            Action::Find => {
                self.find.show_replace = false;
                self.find.focus_replace = false;
                self.overlay = Overlay::Find;
                self.set_status(self.find.status());
            }
            Action::ProjectFind => {
                self.overlay = Overlay::ProjectFind {
                    query: String::new(),
                    selected: 0,
                    case_sensitive: false,
                    use_regex: false,
                    hits: Vec::new(),
                    status:
                        "project find · digite · Alt+G glob · Alt+C case · Alt+R regex · Enter abre"
                            .into(),
                    replace_query: None,
                    file_glob: None,
                    focus_field: 0,
                };
            }
            Action::ProjectReplace => {
                self.overlay = Overlay::ProjectFind {
                    query: String::new(),
                    selected: 0,
                    case_sensitive: false,
                    use_regex: false,
                    hits: Vec::new(),
                    status: "project replace · digite busca, substituição e glob opcional".into(),
                    replace_query: Some(String::new()),
                    file_glob: None,
                    focus_field: 0,
                };
            }
            Action::FindNext => {
                self.jump_find(true);
            }
            Action::FindPrev => {
                self.jump_find(false);
            }
            Action::Replace => {
                self.find.show_replace = true;
                self.find.focus_replace = true;
                self.overlay = Overlay::Find;
                self.set_status(self.find.status());
            }
            Action::SelectAll => {
                self.store.active_mut()?.select_all();
                self.set_status("select all");
            }
            Action::Copy => {
                let text = self
                    .store
                    .active()
                    .map(|document| {
                        let selected_text = document.selected_text();
                        if selected_text.is_empty() {
                            document
                                .caret()
                                .ok()
                                .and_then(|caret| document.buffer().line_text(caret.line).ok())
                                .unwrap_or_default()
                        } else {
                            selected_text
                        }
                    })
                    .unwrap_or_default();
                let _ = clipboard::copy_text(&text);
                self.set_status(format!("copied {} bytes", text.len()));
            }
            Action::Paste => {
                let text = clipboard::paste_text();
                if text.is_empty() {
                    self.set_status("clipboard vazio (Ctrl+C copia · buffer interno se sem X11)");
                } else {
                    self.store.active_mut()?.insert_text(&text)?;
                    self.set_status(format!("pasted {} bytes", text.len()));
                }
            }
            Action::Cut => {
                let text = {
                    let document = self.store.active()?;
                    let selected_text = document.selected_text();
                    if selected_text.is_empty() {
                        if let Ok(caret) = document.caret() {
                            document.buffer().line_text(caret.line).unwrap_or_default()
                        } else {
                            String::new()
                        }
                    } else {
                        selected_text
                    }
                };
                if text.is_empty() {
                    self.set_status("nada para cortar");
                    return Ok(());
                }
                let had_selection = !self.store.active()?.selection().is_empty();
                let _ = clipboard::copy_text(&text);
                if had_selection {
                    self.store.active_mut()?.delete_selection()?;
                } else {
                    let document = self.store.active_mut()?;
                    let caret = document.caret()?;
                    let start_offset = document.buffer().line_to_byte(caret.line)?;
                    let next_line = caret.line + 1;
                    let end_offset = if next_line < document.buffer().line_count() {
                        document.buffer().line_to_byte(next_line)?
                    } else {
                        oride_core::ByteOffset::new(document.buffer().len_bytes())
                    };
                    if start_offset == end_offset {
                        return Ok(());
                    }
                    document.select_byte_range(start_offset, end_offset);
                    document.delete_selection()?;
                }
                self.set_status(format!("cut {} bytes", text.len()));
            }
            Action::Undo => {
                let document = self.store.active_mut()?;
                if !document.undo() {
                    self.set_status("nothing to undo");
                }
            }
            Action::Redo => {
                let document = self.store.active_mut()?;
                if !document.redo() {
                    self.set_status("nothing to redo");
                }
            }
            Action::InsertNewline => {
                self.quit_confirm_pending = false;
                self.insert_newline_smart()?;
            }
            Action::InsertTab => {
                self.quit_confirm_pending = false;
                self.insert_tab()?;
            }
            Action::Backspace => {
                self.quit_confirm_pending = false;
                self.store.active_mut()?.backspace()?;
            }
            Action::Delete => {
                self.quit_confirm_pending = false;
                self.store.active_mut()?.delete_forward()?;
            }
            Action::MoveLeft { extend } => {
                self.store.active_mut()?.move_left(extend)?;
            }
            Action::MoveRight { extend } => {
                self.store.active_mut()?.move_right(extend)?;
            }
            Action::MoveUp { extend } => {
                self.store.active_mut()?.move_up(extend)?;
            }
            Action::MoveDown { extend } => {
                self.store.active_mut()?.move_down(extend)?;
            }
            Action::MoveLineStart { extend } => {
                self.store.active_mut()?.move_line_start(extend)?;
            }
            Action::MoveLineEnd { extend } => {
                self.store.active_mut()?.move_line_end(extend)?;
            }
            Action::MoveDocStart { extend } => {
                self.store.active_mut()?.move_buffer_start(extend)?;
            }
            Action::MoveDocEnd { extend } => {
                self.store.active_mut()?.move_buffer_end(extend)?;
            }
            Action::PageUp => {
                let steps = self.last_editor_height.saturating_sub(1).max(1);
                let document = self.store.active_mut()?;
                for _ in 0..steps {
                    document.move_up(false)?;
                }
            }
            Action::PageDown => {
                let steps = self.last_editor_height.saturating_sub(1).max(1);
                let document = self.store.active_mut()?;
                for _ in 0..steps {
                    document.move_down(false)?;
                }
            }
            Action::ToggleTree => {
                self.show_tree = !self.show_tree;
                if self.show_tree {
                    self.focus = Focus::Tree;
                    self.set_status("árvore visível · foco na árvore (Ctrl+E volta ao editor)");
                } else {
                    if self.focus == Focus::Tree {
                        self.focus = Focus::Editor;
                    }
                    self.set_status("árvore oculta");
                }
            }
            Action::ToggleTerminal => {
                self.ensure_terminal();
                if let Some(terminal) = self.terminal.as_mut() {
                    terminal.toggle_visible();
                    if terminal.visible {
                        self.focus = Focus::Terminal;
                        self.set_status(
                            "terminal: digite · Esc=editor · Ctrl+C no shell · Ctrl+` fecha",
                        );
                    } else if self.focus == Focus::Terminal {
                        self.focus = Focus::Editor;
                        self.set_status("terminal oculto");
                    }
                } else {
                    self.set_status("terminal unavailable (PTY)");
                }
            }
            Action::TerminalGrow => {
                let height = if let Some(terminal) = self.terminal.as_mut() {
                    terminal.grow(2);
                    Some(terminal.height_lines)
                } else {
                    None
                };
                if let Some(h) = height {
                    self.set_status(format!("terminal height {h}"));
                }
            }
            Action::TerminalShrink => {
                let height = if let Some(terminal) = self.terminal.as_mut() {
                    terminal.shrink(2);
                    Some(terminal.height_lines)
                } else {
                    None
                };
                if let Some(h) = height {
                    self.set_status(format!("terminal height {h}"));
                }
            }
            Action::ReloadFile => match self.store.active_mut() {
                Ok(document) => match document.reload_from_disk() {
                    Ok(()) => {
                        self.apply_editorconfig_for_active();
                        self.lsp_sync_active();
                        self.set_status("file reloaded");
                    }
                    Err(error) => self.set_status(format!("reload: {error}")),
                },
                Err(error) => self.set_status(format!("reload: {error}")),
            },
            Action::ToggleDiagnostics => {
                self.show_diagnostics = !self.show_diagnostics;
                if self.show_diagnostics {
                    self.overlay = Overlay::Diagnostics { selected: 0 };
                    self.set_status(format!("{} diagnostics", self.diagnostics.len()));
                } else if matches!(self.overlay, Overlay::Diagnostics { .. }) {
                    self.overlay = Overlay::None;
                }
            }
            Action::LspComplete => self.lsp_complete()?,
            Action::LspHover => self.lsp_hover()?,
            Action::LspGotoDefinition => self.lsp_goto()?,
            Action::LspFormat => self.lsp_format()?,
            Action::SplitVertical => self.split_editor(SplitOrientation::Vertical),
            Action::SplitHorizontal => self.split_editor(SplitOrientation::Horizontal),
            Action::FocusNextPane => self.focus_next_pane(),
            Action::ClosePane => {
                if self.split.close_focused() {
                    let id = self.split.focused_pane().doc_id;
                    let _ = self.store.set_active(id);
                    self.scroll_y = self.split.focused_pane().scroll_y;
                    self.set_status("pane closed");
                } else {
                    self.set_status("já em pane único");
                }
            }
            Action::ResizePaneGrow => {
                if self.split.is_split() {
                    self.split.grow(5);
                    self.set_status(format!("split ratio {}%", self.split.ratio_percent));
                } else {
                    self.set_status("split inativo (use Ctrl+\\ para dividir)");
                }
            }
            Action::ResizePaneShrink => {
                if self.split.is_split() {
                    self.split.shrink(5);
                    self.set_status(format!("split ratio {}%", self.split.ratio_percent));
                } else {
                    self.set_status("split inativo (use Ctrl+\\ para dividir)");
                }
            }
            Action::TreeGrow => {
                self.tree_width = (self.tree_width + 2).min(80);
                self.set_status(format!("tree width {}", self.tree_width));
            }
            Action::TreeShrink => {
                self.tree_width = self.tree_width.saturating_sub(2).max(12);
                self.set_status(format!("tree width {}", self.tree_width));
            }
            Action::AddCursorAbove => {
                self.store.active_mut()?.add_cursor_above()?;
                self.set_status(format!(
                    "cursors: {}",
                    1 + self
                        .store
                        .active()
                        .map(|document| document.extra_carets().len())
                        .unwrap_or(0)
                ));
            }
            Action::AddCursorBelow => {
                self.store.active_mut()?.add_cursor_below()?;
                self.set_status(format!(
                    "cursors: {}",
                    1 + self
                        .store
                        .active()
                        .map(|document| document.extra_carets().len())
                        .unwrap_or(0)
                ));
            }
            Action::ClearExtraCursors => {
                if let Ok(document) = self.store.active_mut() {
                    document.clear_extra_carets();
                }
                self.set_status("multi-cursor limpo");
            }
            Action::FocusTree => {
                self.show_tree = true;
                if self.tree.is_none() {
                    self.tree =
                        ProjectTree::open(&self.workspace, self.config.tree.show_hidden).ok();
                }
                self.focus = Focus::Tree;
                self.set_status(
                    "foco: árvore · ↑↓/jk · Enter abrir · ←→ expandir/colapsar · Ctrl+E editor",
                );
            }
            Action::FocusEditor => {
                self.focus = Focus::Editor;
                self.set_status("foco: editor (Ctrl+B árvore · Ctrl+O abrir pasta)");
            }
            Action::FocusToggleTreeEditor => match self.focus {
                Focus::Tree => {
                    self.focus = Focus::Editor;
                    self.set_status("foco: editor");
                }
                _ => {
                    self.show_tree = true;
                    if self.tree.is_none() {
                        self.tree =
                            ProjectTree::open(&self.workspace, self.config.tree.show_hidden).ok();
                    }
                    self.focus = Focus::Tree;
                    self.set_status("foco: árvore");
                }
            },
            Action::FocusTerminal => {
                self.ensure_terminal();
                if let Some(terminal) = self.terminal.as_mut() {
                    terminal.visible = true;
                    self.focus = Focus::Terminal;
                    self.set_status("foco: terminal · digite · Esc=editor");
                } else {
                    self.set_status("terminal unavailable");
                }
            }
            Action::ToggleScm => {
                self.show_scm = !self.show_scm;
                if self.show_scm {
                    self.refresh_scm_cache();
                    self.focus = Focus::Scm;
                    self.set_status(
                        "SCM · ↑↓ · Enter abre · d diff · r refresh · Esc editor · Ctrl+Shift+G",
                    );
                } else {
                    if self.focus == Focus::Scm {
                        self.focus = Focus::Editor;
                    }
                    self.set_status("SCM oculto");
                }
            }
            Action::FocusScm => {
                self.show_scm = true;
                self.refresh_scm_cache();
                self.focus = Focus::Scm;
                self.set_status("foco: SCM");
            }
            Action::GitPull => match git_pull(&self.workspace) {
                Ok(msg) => {
                    self.set_status(format!("git pull: {msg}"));
                    self.refresh_git_and_index();
                    self.refresh_scm_cache();
                }
                Err(err) => self.set_status(format!("git pull: {err}")),
            },
            Action::GitPush => match git_push(&self.workspace) {
                Ok(msg) => {
                    self.set_status(format!("git push: {msg}"));
                    self.refresh_git_and_index();
                    self.refresh_scm_cache();
                }
                Err(err) => self.set_status(format!("git push: {err}")),
            },
            Action::BufferPicker => {
                self.overlay = Overlay::BufferPicker {
                    query: String::new(),
                    selected: 0,
                };
            }
            Action::JumpBack => {
                self.record_jump();
                if let Some(jump) = self.jump_list.back() {
                    self.goto_jump(&jump);
                    self.set_status("jump back");
                } else {
                    self.set_status("jump list vazio");
                }
            }
            Action::JumpForward => {
                if let Some(jump) = self.jump_list.forward() {
                    self.goto_jump(&jump);
                    self.set_status("jump forward");
                } else {
                    self.set_status("sem jump forward");
                }
            }
            Action::WhichKey => {
                self.show_which_key = true;
            }
            Action::Welcome => {
                self.show_welcome = true;
            }
            Action::ShowDiff => {
                self.open_diff_for_active();
            }
            Action::Surround => {
                if self
                    .store
                    .active()
                    .map(|document| document.selection().is_empty())
                    .unwrap_or(true)
                {
                    self.set_status("surround: selecione texto primeiro");
                } else {
                    self.surround_pending = true;
                    self.overlay = Overlay::SurroundPick;
                    self.set_status(
                        "surround: digite o par ( [ { < ou aspas/backtick · Esc cancela",
                    );
                }
            }
            Action::MultiPicker => {
                self.overlay = Overlay::MultiPicker {
                    query: String::new(),
                    selected: 0,
                };
            }
            Action::UndoTree => {
                self.overlay = Overlay::UndoTree { selected: 0 };
            }
            Action::ToggleMouse => {
                self.mouse_enabled = !self.mouse_enabled;
                self.config.mouse = self.mouse_enabled;
                self.set_status(if self.mouse_enabled {
                    "mouse: ON · clique/drag/scroll · desligar: View → Mouse ou mouse=false"
                } else {
                    "mouse: OFF (default) · ligar: View → Enable mouse ou mouse=true no TOML"
                });
            }
            Action::OpenFolder => {
                let browser = PathBrowser::new(&self.workspace, BrowseMode::Folder);
                self.set_status(browser.hint());
                self.overlay = Overlay::Browse(browser);
            }
            Action::OpenFileFuzzy => {
                let browser = PathBrowser::new(&self.workspace, BrowseMode::File);
                self.set_status(browser.hint());
                self.overlay = Overlay::Browse(browser);
            }
            Action::ToggleSoftWrap => {
                self.soft_wrap = !self.soft_wrap;
                self.set_status(if self.soft_wrap {
                    "soft wrap: on"
                } else {
                    "soft wrap: off"
                });
            }
            Action::ToggleMdPreview => {
                let language = self
                    .store
                    .active()
                    .ok()
                    .map(|document| detect_language(document.path()))
                    .unwrap_or_default();
                if !language.is_markdown_family() {
                    self.set_status("preview só para Markdown/MDX");
                } else {
                    self.show_md_preview = !self.show_md_preview;
                    self.preview_scroll = 0;
                    self.set_status(if self.show_md_preview {
                        "md preview: on · segue scroll do editor · Alt+↑/↓ fine · Alt+P fecha"
                    } else {
                        "md preview: off"
                    });
                }
            }
            Action::ToggleComment => {
                self.toggle_line_comment()?;
            }
            Action::NextTab => {
                self.store.activate_next_tab();
                self.scroll_y = 0;
                if let Some(id) = self.store.active_id() {
                    self.split.set_focused_doc(id);
                }
                self.lsp_open_active();
            }
            Action::PrevTab => {
                self.store.activate_prev_tab();
                self.scroll_y = 0;
                if let Some(id) = self.store.active_id() {
                    self.split.set_focused_doc(id);
                }
                self.lsp_open_active();
            }
            Action::NewTab => {
                let id = self.store.open_empty();
                self.split.set_focused_doc(id);
                self.focus = Focus::Editor;
                self.scroll_y = 0;
            }
            Action::CloseTab => self.close_active_tab()?,
            Action::CommandPalette => {
                self.overlay = Overlay::CommandPalette {
                    query: String::new(),
                    selected: 0,
                };
            }
            Action::TreeNewFile => {
                self.show_tree = true;
                self.focus = Focus::Tree;
                self.overlay = Overlay::Prompt {
                    kind: PromptKind::NewFile,
                    buffer: String::new(),
                };
            }
            Action::TreeNewDir => {
                self.show_tree = true;
                self.focus = Focus::Tree;
                self.overlay = Overlay::Prompt {
                    kind: PromptKind::NewDir,
                    buffer: String::new(),
                };
            }
            Action::TreeRefresh => {
                if let Some(tree) = self.tree.as_mut() {
                    if let Err(error) = tree.refresh() {
                        self.set_status(format!("refresh: {error}"));
                    } else {
                        self.set_status("tree refreshed");
                    }
                }
                self.refresh_git_and_index();
                self.refresh_scm_cache();
            }
            Action::SelectTheme => {
                let registry = oride_config::ThemeRegistry::load_with_paths(Some(&self.workspace));
                let mut themes = registry.list_names();
                themes.sort();
                let current = self.config.theme.clone();
                let selected = themes
                    .iter()
                    .position(|t| {
                        oride_config::normalize_theme_name(t)
                            == oride_config::normalize_theme_name(&current)
                    })
                    .unwrap_or(0);
                self.overlay = Overlay::ThemePicker {
                    query: String::new(),
                    selected,
                    initial_theme: current,
                    themes,
                };
                self.set_status("seletor de tema: ↑↓ navega/preview · Enter aplica · Esc cancela");
            }
            Action::SelectLocale => {
                let all_locales = oride_i18n::available_locales();
                let locales = all_locales
                    .iter()
                    .map(|loc| loc.menu_label_for())
                    .collect::<Vec<_>>();
                let selected = all_locales
                    .iter()
                    .position(|l| l.as_str() == self.locale.as_str())
                    .unwrap_or(0);
                self.overlay = Overlay::LocalePicker {
                    query: String::new(),
                    selected,
                    locales,
                };
                self.set_status("seletor de idioma: ↑↓ seleciona · Enter aplica · Esc cancela");
            }
            Action::HealthCheck => {
                let report = crate::health::check_health(
                    &self.config.theme,
                    self.locale.as_str(),
                    self.vim.is_some(),
                );
                self.overlay = Overlay::HealthCheck { scroll: 0, report };
                self.set_status("diagnóstico do ambiente (:health) — Esc ou Enter para fechar");
            }
            Action::ToggleModal => {
                if self.vim.is_some() {
                    self.vim = None;
                    self.set_status("modo modal desativado (CUA padrão)");
                } else {
                    self.vim = Some(crate::modal::VimState::default());
                    self.set_status("modo modal ativado (Vim Normal)");
                }
            }
            Action::RunTasks => {
                self.execute_task_runner(None);
            }
        }
        Ok(())
    }

    pub(crate) fn active_language(&self) -> LanguageId {
        self.store
            .active()
            .ok()
            .map(|document| self.detect_document_language(document.path()))
            .unwrap_or(LanguageId::Plain)
    }

    pub(crate) fn detect_document_language(&self, path: Option<&Path>) -> LanguageId {
        if let Some(path) = path {
            if let Some(id) = self.plugin_host.detect_language_by_path(path) {
                return id;
            }
        }
        detect_language(path)
    }

    pub(crate) fn insert_newline_smart(&mut self) -> Result<(), DocumentError> {
        let language = self.active_language();
        if language.is_markdown_family() {
            let (line, caret_line) = {
                let document = self.store.active()?;
                let caret = document.caret()?;
                (
                    document.buffer().line_text(caret.line).unwrap_or_default(),
                    caret.line,
                )
            };
            if let Some(continuation) = continue_list_on_enter(&line) {
                self.store
                    .active_mut()?
                    .insert_text(&format!("\n{continuation}"))?;
                return Ok(());
            }
            if let Some(prefix) = oride_syntax::list_prefix(&line) {
                if line[prefix.len()..].trim().is_empty() {
                    let document = self.store.active_mut()?;
                    let start = document.buffer().line_to_byte(caret_line)?;
                    let end = oride_core::ByteOffset::new(start.as_usize() + line.len());
                    document.set_selection(Selection::new(start, end));
                    document.delete_selection()?;
                    return Ok(());
                }
            }
        }
        self.store.active_mut()?.insert_text("\n")?;
        Ok(())
    }

    pub(crate) fn toggle_line_comment(&mut self) -> Result<(), DocumentError> {
        let language = self.active_language();
        let open = match self.plugin_host.comment_open(language) {
            Some(opening) => opening,
            None => {
                self.set_status("comentário não definido para esta linguagem");
                return Ok(());
            }
        };
        let close = self.plugin_host.comment_close(language).unwrap_or("");
        let document = self.store.active_mut()?;
        let caret = document.caret()?;
        let line = document.buffer().line_text(caret.line).unwrap_or_default();
        let indent_len = line.len() - line.trim_start().len();
        let indent = &line[..indent_len];
        let body = line.trim_start();

        let new_line = if close.is_empty() {
            let open_trimmed = open.trim_end();
            if let Some(rest) = body.strip_prefix(open_trimmed) {
                let rest = rest.strip_prefix(' ').unwrap_or(rest);
                format!("{indent}{rest}")
            } else {
                format!("{indent}{open}{body}")
            }
        } else {
            let open_trimmed = open.trim();
            let close_trimmed = close.trim();
            if body.starts_with(open_trimmed) && body.ends_with(close_trimmed) {
                let inner = body
                    .strip_prefix(open_trimmed)
                    .and_then(|s| s.strip_suffix(close_trimmed))
                    .unwrap_or(body)
                    .trim();
                format!("{indent}{inner}")
            } else {
                format!("{indent}{open_trimmed} {body} {close_trimmed}")
            }
        };

        let start = document.buffer().line_to_byte(caret.line)?;
        let end = oride_core::ByteOffset::new(start.as_usize() + line.len());
        document.set_selection(Selection::new(start, end));
        document.delete_selection()?;
        document.insert_text(&new_line)?;
        let head = document.buffer().line_to_byte(caret.line)?;
        document.set_selection(Selection::caret(head));
        Ok(())
    }

    pub(crate) fn apply_language_defaults(&mut self, language: LanguageId) {
        if self.plugin_host.default_soft_wrap(language) {
            self.soft_wrap = true;
        }
        self.fire_plugin_hook(PluginHook::OnOpen);
    }

    pub(crate) fn run_plugin_command(&mut self, command_id: &str) {
        let text = self
            .store
            .active()
            .map(|document| document.buffer().as_string())
            .unwrap_or_default();
        let path = self
            .store
            .active()
            .ok()
            .and_then(|document| document.path().map(Path::to_path_buf));
        let dirty = self
            .store
            .active()
            .map(|document| document.is_dirty())
            .unwrap_or(false);
        let workspace = self.workspace.clone();
        let mut adapter = HostCtx {
            status: None,
            workspace,
            path,
            text,
            dirty,
        };
        match self.plugin_host.run_command(command_id, &mut adapter) {
            Ok(()) => {
                if let Some(status_message) = adapter.status {
                    self.set_status(status_message);
                }
            }
            Err(error) => self.set_status(format!("plugin: {error}")),
        }
    }

    pub(crate) fn split_editor(&mut self, orientation: SplitOrientation) {
        let id = match self.store.active_id() {
            Some(id) => id,
            None => {
                self.set_status("sem documento");
                return;
            }
        };
        self.split.sync_scroll(self.scroll_y);
        self.split.split(orientation, id);
        let _ = self.store.set_active(id);
        self.set_status(match orientation {
            SplitOrientation::Vertical => "split vertical · F6 / Ctrl+Alt+←→ troca pane",
            SplitOrientation::Horizontal => "split horizontal · F6 troca pane",
        });
    }

    pub(crate) fn focus_next_pane(&mut self) {
        if !self.split.is_split() {
            self.set_status("sem split");
            return;
        }
        self.split.sync_scroll(self.scroll_y);
        self.split.focus_next();
        let pane = self.split.focused_pane().clone();
        let _ = self.store.set_active(pane.doc_id);
        self.scroll_y = pane.scroll_y;
        self.set_status(format!(
            "pane {}/{}",
            self.split.focused + 1,
            self.split.panes.len()
        ));
    }

    pub(crate) fn fire_plugin_hook(&mut self, hook: PluginHook) {
        let text = self
            .store
            .active()
            .map(|document| document.buffer().as_string())
            .unwrap_or_default();
        let path = self
            .store
            .active()
            .ok()
            .and_then(|document| document.path().map(Path::to_path_buf));
        let dirty = self
            .store
            .active()
            .map(|document| document.is_dirty())
            .unwrap_or(false);
        let workspace = self.workspace.clone();
        let mut adapter = HostCtx {
            status: None,
            workspace,
            path,
            text,
            dirty,
        };
        self.plugin_host.dispatch_hook(hook, &mut adapter);
        if let Some(status_message) = adapter.status {
            self.set_status(status_message);
        }
    }

    pub(crate) fn open_workspace_folder(&mut self, path: PathBuf) {
        let path = if path.as_os_str().is_empty() {
            self.set_status("caminho vazio");
            return;
        } else {
            path
        };
        let canonical_path = match std::fs::canonicalize(&path) {
            Ok(resolved) => resolved,
            Err(error) => {
                self.set_status(format!("pasta inválida: {error}"));
                return;
            }
        };
        if !canonical_path.is_dir() {
            self.set_status(format!("não é pasta: {}", canonical_path.display()));
            return;
        }
        self.workspace = canonical_path;
        match ProjectTree::open(&self.workspace, self.config.tree.show_hidden) {
            Ok(tree) => self.tree = Some(tree),
            Err(error) => {
                self.tree = None;
                self.set_status(format!("árvore: {error}"));
                return;
            }
        }
        self.show_tree = true;
        self.focus = Focus::Tree;
        self.tree_scroll = 0;
        self.refresh_git_and_index();
        if let Some(old_terminal) = self.terminal.take() {
            drop(old_terminal);
        }
        self.terminal = EmbeddedTerminal::spawn(
            &self.workspace,
            80,
            self.config.terminal.default_height,
            Some(&self.config.terminal.shell),
        )
        .ok();
        self.disk_watch = DiskWatch::start(&self.workspace);
        self.lsp_clients.clear();
        self.lsp_failures.clear();
        self.lsp_open_active();
        self.set_status(format!("projeto: {}", self.workspace.display()));
    }

    pub(crate) fn close_active_tab(&mut self) -> Result<(), DocumentError> {
        let id = self
            .store
            .active_id()
            .ok_or(DocumentError::NoActiveDocument)?;
        let dirty = self
            .store
            .get(id)
            .map(|document| document.is_dirty())
            .unwrap_or(false);
        if dirty && self.close_tab_confirm != Some(id) {
            self.close_tab_confirm = Some(id);
            self.set_status("tab dirty — Ctrl+W again to close without save");
            return Ok(());
        }
        self.close_tab_confirm = None;
        if let Some(path) = self
            .store
            .get(id)
            .and_then(|document| document.path())
            .map(Path::to_path_buf)
        {
            let language = detect_language(Some(&path));
            if let Some(client) = self.lsp_clients.get_mut(&language) {
                let _ = client.did_close(&path);
            }
            self.diagnostics
                .retain(|(diagnostic_path, _)| diagnostic_path != &path);
        }
        let _ = self.store.close(id)?;
        if self.store.tab_ids().is_empty() {
            self.store.open_empty();
        }
        if let Some(active_id) = self.store.active_id() {
            self.split.set_focused_doc(active_id);
            for pane in &mut self.split.panes {
                if pane.doc_id == id {
                    pane.doc_id = active_id;
                    pane.scroll_y = 0;
                }
            }
        }
        self.scroll_y = 0;
        Ok(())
    }

    pub fn open_document_path(&mut self, path: &Path) -> Result<DocumentId, DocumentError> {
        let id = self.store.open_path(path)?;
        self.split.set_focused_doc(id);
        self.scroll_y = 0;
        let language = detect_language(Some(path));
        self.apply_language_defaults(language);
        self.apply_editorconfig_for_active();
        self.lsp_open_active();
        Ok(id)
    }

    pub(crate) fn refresh_git_and_index(&mut self) {
        self.git_status = status_map(&self.workspace);
        self.git_branch = current_branch(&self.workspace);
        self.git_ahead_behind = ahead_behind(&self.workspace);
        self.blame_cache = None;
        self.blame_refresh_at = Some(Instant::now() + Duration::from_millis(150));
        self.refresh_file_index();
    }

    pub(crate) fn refresh_file_index(&mut self) {
        self.file_index =
            list_files_recursive(&self.workspace, self.config.tree.show_hidden).unwrap_or_default();
    }

    pub(crate) fn refresh_scm_cache(&mut self) {
        self.scm_cache = scm_entries(&self.workspace);
        if self.scm_selected >= self.scm_cache.len() && !self.scm_cache.is_empty() {
            self.scm_selected = self.scm_cache.len() - 1;
        }
        if self.scm_cache.is_empty() {
            self.scm_selected = 0;
        }
    }

    pub(crate) fn ensure_terminal(&mut self) {
        let is_dead = self
            .terminal
            .as_ref()
            .and_then(|t| t.last_error.as_deref())
            .map(|err| err.contains("encerrado"))
            .unwrap_or(false);
        if self.terminal.is_none() || is_dead {
            let height = self
                .terminal
                .as_ref()
                .map(|t| t.height_lines)
                .unwrap_or(self.config.terminal.default_height)
                .max(3);
            match EmbeddedTerminal::spawn(
                &self.workspace,
                80,
                height,
                Some(&self.config.terminal.shell),
            ) {
                Ok(mut term) => {
                    term.height_lines = height;
                    term.visible = true;
                    self.terminal = Some(term);
                }
                Err(error) => self.set_status(format!("pty: {error}")),
            }
        }
    }

    pub(crate) fn record_jump(&mut self) {
        let Ok(document) = self.store.active() else {
            return;
        };
        let path = document.path().map(Path::to_path_buf);
        let byte = document.selection().head;
        let line = document.caret().map(|c| c.line).unwrap_or(0);
        self.jump_list.push(Jump { path, byte, line });
    }

    pub(crate) fn goto_jump(&mut self, jump: &Jump) {
        if let Some(path) = &jump.path {
            let _ = self.open_document_path(path);
        }
        if let Ok(document) = self.store.active_mut() {
            document.jump_to_byte(jump.byte);
        }
        self.scroll_y = jump.line.saturating_sub(2);
        self.focus = Focus::Editor;
        self.ensure_cursor_visible();
    }

    pub(crate) fn open_diff_for_active(&mut self) {
        let path = self
            .store
            .active()
            .ok()
            .and_then(|document| document.path().map(Path::to_path_buf));
        let Some(path) = path else {
            self.set_status("diff: sem path");
            return;
        };
        self.open_diff_for_path(&path);
    }

    pub(crate) fn open_diff_for_path(&mut self, path: &Path) {
        match diff_file(&self.workspace, path) {
            Some(text) => {
                let lines: Vec<String> = text.lines().map(|line| line.to_string()).collect();
                self.overlay = Overlay::Diff {
                    path: path.to_path_buf(),
                    lines,
                    scroll: 0,
                };
            }
            None => self.set_status(format!("diff vazio: {}", path.display())),
        }
    }

    pub(crate) fn apply_theme_by_name(&mut self, name: &str) -> bool {
        let registry = oride_config::ThemeRegistry::load_with_paths(Some(&self.workspace));
        if let Some(theme_def) = registry.get(name) {
            if let Ok(ui_theme) = UiTheme::from_theme_definition(theme_def) {
                self.theme = ui_theme;
                self.config.theme = theme_def.name.clone();
                self.config.theme_ui = theme_def.ui.clone();
                self.config.syntax = theme_def.syntax.clone();
                return true;
            }
        }
        false
    }

    pub fn update_locale(&mut self, locale: oride_i18n::Locale) {
        self.locale = locale;
        self.config.locale = self.locale.as_str().to_string();
        self.menus = crate::menus::menus_for_locale(&self.locale);
        let _ = oride_config::save_user_locale(self.locale.as_str());
        let name = self.locale.display_name();
        self.set_status(self.locale.messages().locale_applied(&name));
    }

    pub(crate) fn apply_action_id(&mut self, id: &str) {
        if let Some(cmd) = id.strip_prefix("plugin:") {
            self.run_plugin_command(cmd);
            return;
        }
        match parse_action(id) {
            Ok(action) => {
                let _ = self.apply_action(action);
            }
            Err(_) => self.set_status(format!("ação desconhecida: {id}")),
        }
    }

    pub(crate) fn apply_surround(&mut self, open: char) {
        let close = match open {
            '(' => ')',
            '[' => ']',
            '{' => '}',
            '<' => '>',
            '"' => '"',
            '\'' => '\'',
            '`' => '`',
            _ => open,
        };
        let Ok(document) = self.store.active_mut() else {
            return;
        };
        if document.selection().is_empty() {
            self.set_status("surround: sem seleção");
            return;
        }
        let selected = document.selected_text();
        let wrapped = format!("{open}{selected}{close}");
        let _ = document.delete_selection();
        let _ = document.insert_text(&wrapped);
        self.set_status(format!("surround: {open}…{close}"));
    }

    pub(crate) fn recompute_find(&mut self) {
        let text = self
            .store
            .active()
            .map(|document| document.buffer().as_string())
            .unwrap_or_default();
        self.find.recompute(&text);
    }

    pub(crate) fn recompute_find_and_jump(&mut self) {
        self.recompute_find();
        if let Some(matched) = self.find.current_match() {
            if let Ok(document) = self.store.active_mut() {
                document.select_byte_range(
                    oride_core::ByteOffset::new(matched.start),
                    oride_core::ByteOffset::new(matched.end),
                );
            }
            self.ensure_cursor_visible();
        }
    }

    pub(crate) fn jump_find(&mut self, forward: bool) {
        self.recompute_find();
        let matched = if forward {
            self.find.next()
        } else {
            self.find.prev()
        };
        if let Some(m) = matched {
            if let Ok(document) = self.store.active_mut() {
                document.select_byte_range(
                    oride_core::ByteOffset::new(m.start),
                    oride_core::ByteOffset::new(m.end),
                );
            }
            self.ensure_cursor_visible();
        }
        self.set_status(self.find.status());
    }

    pub(crate) fn replace_current_match(&mut self) {
        self.recompute_find();
        let Some(matched) = self.find.current_match() else {
            self.set_status("replace: nenhuma ocorrência");
            return;
        };
        let replacement = self.find.replace.clone();
        if let Ok(document) = self.store.active_mut() {
            document.select_byte_range(
                oride_core::ByteOffset::new(matched.start),
                oride_core::ByteOffset::new(matched.end),
            );
            let _ = document.delete_selection();
            let _ = document.insert_text(&replacement);
        }
        self.recompute_find();
        if let Some(matched) = self.find.current_match() {
            if let Ok(document) = self.store.active_mut() {
                document.select_byte_range(
                    oride_core::ByteOffset::new(matched.start),
                    oride_core::ByteOffset::new(matched.end),
                );
            }
        }
        self.set_status(format!("replaced 1 · {}", self.find.status()));
    }

    pub(crate) fn replace_all_matches(&mut self) {
        self.recompute_find();
        if self.find.matches.is_empty() {
            self.set_status("replace all: 0 ocorrências");
            return;
        }
        let replacement = self.find.replace.clone();
        let matches = self.find.matches.clone();
        let total_matches = matches.len();
        if let Ok(document) = self.store.active_mut() {
            for matched in matches.into_iter().rev() {
                document.select_byte_range(
                    oride_core::ByteOffset::new(matched.start),
                    oride_core::ByteOffset::new(matched.end),
                );
                let _ = document.delete_selection();
                let _ = document.insert_text(&replacement);
            }
        }
        self.recompute_find();
        self.set_status(format!("replaced {total_matches} ocorrência(s)"));
    }

    pub(crate) fn open_save_as_browser(&mut self) {
        let (start_dir, name) = self
            .store
            .active()
            .ok()
            .and_then(|document| document.path().map(|p| p.to_path_buf()))
            .map(|path| {
                let file_name = path
                    .file_name()
                    .map(|n| n.to_string_lossy().into_owned())
                    .unwrap_or_else(|| "untitled.txt".into());
                let parent_dir = path
                    .parent()
                    .map(Path::to_path_buf)
                    .unwrap_or_else(|| self.workspace.clone());
                (parent_dir, file_name)
            })
            .unwrap_or_else(|| (self.workspace.clone(), "untitled.txt".into()));
        let mut browser = PathBrowser::new(&start_dir, BrowseMode::SaveAs);
        browser.filter = name;
        self.set_status(format!(
            "{} · (atalhos: Ctrl+Shift+S · F12 · Alt+Shift+S)",
            browser.hint()
        ));
        self.overlay = Overlay::Browse(browser);
    }

    pub(crate) fn jump_to_project_hit(&mut self, hit: SearchHit) {
        let path = if hit.path.is_absolute() {
            hit.path.clone()
        } else {
            self.workspace.join(&hit.path)
        };
        if let Err(error) = self.open_document_path(&path) {
            self.set_status(format!("open: {error}"));
            return;
        }
        if let Ok(document) = self.store.active_mut() {
            let line = hit.line.saturating_sub(1);
            let column = hit.column.saturating_sub(1);
            if let Ok(offset) = document
                .buffer()
                .caret_to_byte(oride_core::Caret::new(line, column))
            {
                document.jump_to_byte(offset);
            }
        }
        self.scroll_y = hit.line.saturating_sub(1);
        self.ensure_cursor_visible();
        self.focus = Focus::Editor;
        self.overlay = Overlay::None;
        self.set_status(format!(
            "{}:{}  {}",
            path.file_name()
                .and_then(|name| name.to_str())
                .unwrap_or("?"),
            hit.line,
            hit.line_text.chars().take(40).collect::<String>()
        ));
    }

    pub(crate) fn recompute_project_find(
        &self,
        query: &str,
        file_glob: Option<&str>,
        hits: &mut Vec<SearchHit>,
        status: &mut String,
        case_sensitive: bool,
        use_regex: bool,
    ) {
        if query.trim().is_empty() {
            hits.clear();
            *status = "project find · digite · Alt+G glob · Alt+C case · Alt+R regex".into();
            return;
        }
        let search_query = SearchQuery {
            pattern: query.to_string(),
            case_sensitive,
            use_regex,
            file_glob: file_glob.map(str::to_string),
            max_hits: 500,
        };
        match search_project(&self.workspace, &search_query) {
            Ok(result) => {
                *hits = result.hits;
                let backend = match result.backend {
                    oride_search::SearchBackend::Ripgrep => "rg",
                    oride_search::SearchBackend::RustWalk => "rust",
                };
                let truncated_marker = if result.truncated {
                    " · truncated"
                } else {
                    ""
                };
                let glob_info = if let Some(glob) = file_glob {
                    if !glob.trim().is_empty() {
                        format!(" · glob:{glob}")
                    } else {
                        String::new()
                    }
                } else {
                    String::new()
                };
                *status = format!(
                    "{} hits · {backend}{glob_info}{truncated_marker}",
                    hits.len()
                );
            }
            Err(error) => {
                hits.clear();
                *status = format!("erro: {error}");
            }
        }
    }

    pub(crate) fn execute_project_replace(
        &mut self,
        query: &str,
        replacement: &str,
        file_glob: Option<&str>,
        case_sensitive: bool,
        use_regex: bool,
    ) {
        if query.trim().is_empty() {
            self.set_status("busca vazia: nada a substituir");
            return;
        }
        let search_query = SearchQuery {
            pattern: query.to_string(),
            case_sensitive,
            use_regex,
            file_glob: file_glob.map(str::to_string),
            max_hits: 500,
        };
        match oride_search::replace_in_project(&self.workspace, &search_query, replacement) {
            Ok((summary, modified_paths)) => {
                for path in &modified_paths {
                    self.disk_watch.mark_saved(path);
                    if let Some(doc_id) = self.store.document_id_for_path(path) {
                        if let Some(doc) = self.store.get_mut(doc_id) {
                            let _ = doc.reload_from_disk();
                        }
                    }
                }
                self.set_status(format!(
                    "Substituídas {} ocorrências em {} arquivos",
                    summary.replacements_count, summary.files_modified
                ));
                self.overlay = Overlay::None;
            }
            Err(err) => {
                self.set_status(format!("erro na substituição: {err}"));
            }
        }
    }

    pub fn execute_task_runner(&mut self, task_name: Option<&str>) {
        let tasks = crate::tasks::discover_tasks(Some(&self.workspace));
        if tasks.is_empty() {
            self.set_status("nenhuma tarefa em .oride/tasks.toml ou ~/.config/oride/tasks.toml");
            return;
        }

        if let Some(name) = task_name {
            if let Some(task) = tasks.iter().find(|t| t.name.eq_ignore_ascii_case(name)) {
                let task_clone = task.clone();
                self.run_task(&task_clone);
            } else {
                self.set_status(format!("tarefa não encontrada: {name}"));
            }
        } else if let Some(task) = tasks.first() {
            let task_clone = task.clone();
            self.run_task(&task_clone);
        }
    }

    pub fn run_task(&mut self, task: &crate::tasks::TaskDef) {
        let active_path = self
            .store
            .active()
            .ok()
            .and_then(|d| d.path().map(Path::to_path_buf));
        let cursor = if let Ok(doc) = self.store.active() {
            if let Ok(pos) = doc.caret() {
                (pos.line, pos.column)
            } else {
                (0, 0)
            }
        } else {
            (0, 0)
        };

        let resolved_cmd = crate::tasks::resolve_task_variables(
            &task.command,
            active_path.as_deref(),
            &self.workspace,
            cursor,
        );

        if task.run_in == "background" {
            let ws = self.workspace.clone();
            std::thread::spawn(move || {
                let _ = oride_osutil::shell_command(&resolved_cmd)
                    .current_dir(ws)
                    .output();
            });
            self.set_status(format!("tarefa em background: {}", task.name));
        } else {
            self.ensure_terminal();
            if let Some(term) = &mut self.terminal {
                term.visible = true;
                let _ = term.write_str(&format!("{resolved_cmd}\n"));
            }
            self.focus = Focus::Terminal;
            self.set_status(format!("executando tarefa: {}", task.name));
        }
    }
}
