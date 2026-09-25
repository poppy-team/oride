//! Ações nomeadas (ids estáveis em TOML).

use thiserror::Error;

/// Ação de editor resolvida a partir do keymap.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Action {
    Quit,
    Save,
    SaveAs,
    SaveAll,
    Undo,
    Redo,
    InsertNewline,
    InsertTab,
    Backspace,
    Delete,
    MoveLeft { extend: bool },
    MoveRight { extend: bool },
    MoveUp { extend: bool },
    MoveDown { extend: bool },
    MoveLineStart { extend: bool },
    MoveLineEnd { extend: bool },
    MoveDocStart { extend: bool },
    MoveDocEnd { extend: bool },
    SelectAll,
    PageUp,
    PageDown,
    ToggleTree,
    ToggleTerminal,
    TerminalGrow,
    TerminalShrink,
    FocusTree,
    FocusEditor,
    FocusTerminal,
    NextTab,
    PrevTab,
    CloseTab,
    NewTab,
    CommandPalette,
    OpenFileFuzzy,
    TreeNewFile,
    TreeNewDir,
    TreeRefresh,
    OpenFolder,
    FocusToggleTreeEditor,
    ToggleSoftWrap,
    ToggleComment,
    ToggleMdPreview,
    Help,
    Find,
    FindNext,
    FindPrev,
    ProjectFind,
    ProjectReplace,
    Replace,
    Copy,
    Paste,
    Cut,
    ReloadFile,
    // P3 LSP
    LspComplete,
    LspHover,
    LspGotoDefinition,
    LspFormat,
    ToggleDiagnostics,
    // P9 splits + multi-cursor
    SplitVertical,
    SplitHorizontal,
    FocusNextPane,
    ClosePane,
    ResizePaneGrow,
    ResizePaneShrink,
    TreeGrow,
    TreeShrink,
    AddCursorAbove,
    AddCursorBelow,
    ClearExtraCursors,
    // UX polish (menu / SCM / navigation)
    ToggleScm,
    FocusScm,
    GitPull,
    GitPush,
    BufferPicker,
    JumpBack,
    JumpForward,
    WhichKey,
    Welcome,
    ShowDiff,
    // Tier B + mouse
    Surround,
    MultiPicker,
    UndoTree,
    ToggleMouse,
    SelectTheme,
    SelectLocale,
    HealthCheck,
    ToggleModal,
    RunTasks,
}

/// Toda ação distinta, na ordem em que os ids `extend` aparecem no TOML.
///
/// É a tabela canônica de ações do produto: `id()` e `ALL` juntos tornam o
/// mapeamento id↔variante bidirecional e verificável, em vez de depender de o
/// `parse_action` e o keymap concordarem por disciplina.
pub const ALL: &[Action] = &[
    Action::Quit,
    Action::Save,
    Action::SaveAs,
    Action::SaveAll,
    Action::Undo,
    Action::Redo,
    Action::InsertNewline,
    Action::InsertTab,
    Action::Backspace,
    Action::Delete,
    Action::MoveLeft { extend: false },
    Action::MoveRight { extend: false },
    Action::MoveUp { extend: false },
    Action::MoveDown { extend: false },
    Action::MoveLineStart { extend: false },
    Action::MoveLineEnd { extend: false },
    Action::MoveDocStart { extend: false },
    Action::MoveDocEnd { extend: false },
    Action::MoveLeft { extend: true },
    Action::MoveRight { extend: true },
    Action::MoveUp { extend: true },
    Action::MoveDown { extend: true },
    Action::MoveLineStart { extend: true },
    Action::MoveLineEnd { extend: true },
    Action::MoveDocStart { extend: true },
    Action::MoveDocEnd { extend: true },
    Action::SelectAll,
    Action::PageUp,
    Action::PageDown,
    Action::ToggleTree,
    Action::ToggleTerminal,
    Action::TerminalGrow,
    Action::TerminalShrink,
    Action::FocusTree,
    Action::FocusEditor,
    Action::FocusTerminal,
    Action::NextTab,
    Action::PrevTab,
    Action::CloseTab,
    Action::NewTab,
    Action::CommandPalette,
    Action::OpenFileFuzzy,
    Action::TreeNewFile,
    Action::TreeNewDir,
    Action::TreeRefresh,
    Action::OpenFolder,
    Action::FocusToggleTreeEditor,
    Action::ToggleSoftWrap,
    Action::ToggleComment,
    Action::ToggleMdPreview,
    Action::Help,
    Action::Find,
    Action::FindNext,
    Action::FindPrev,
    Action::ProjectFind,
    Action::ProjectReplace,
    Action::Replace,
    Action::Copy,
    Action::Paste,
    Action::Cut,
    Action::ReloadFile,
    Action::LspComplete,
    Action::LspHover,
    Action::LspGotoDefinition,
    Action::LspFormat,
    Action::ToggleDiagnostics,
    Action::SplitVertical,
    Action::SplitHorizontal,
    Action::FocusNextPane,
    Action::ClosePane,
    Action::ResizePaneGrow,
    Action::ResizePaneShrink,
    Action::TreeGrow,
    Action::TreeShrink,
    Action::AddCursorAbove,
    Action::AddCursorBelow,
    Action::ClearExtraCursors,
    Action::ToggleScm,
    Action::FocusScm,
    Action::GitPull,
    Action::GitPush,
    Action::BufferPicker,
    Action::JumpBack,
    Action::JumpForward,
    Action::WhichKey,
    Action::Welcome,
    Action::ShowDiff,
    Action::Surround,
    Action::MultiPicker,
    Action::UndoTree,
    Action::ToggleMouse,
    Action::SelectTheme,
    Action::SelectLocale,
    Action::HealthCheck,
    Action::ToggleModal,
    Action::RunTasks,
];

impl Action {
    /// Id estável usado em TOML, na palette e no modo conformance.
    ///
    /// `MoveLeft { extend: true }` devolve `move_left_extend`: a variante com
    /// `extend` é uma ação distinta, não uma modalidade da mesma ação.
    #[must_use]
    pub fn id(self) -> &'static str {
        match self {
            Self::Quit => "quit",
            Self::Save => "save",
            Self::SaveAs => "save_as",
            Self::SaveAll => "save_all",
            Self::Undo => "undo",
            Self::Redo => "redo",
            Self::InsertNewline => "insert_newline",
            Self::InsertTab => "insert_tab",
            Self::Backspace => "backspace",
            Self::Delete => "delete",
            Self::MoveLeft { extend: false } => "move_left",
            Self::MoveRight { extend: false } => "move_right",
            Self::MoveUp { extend: false } => "move_up",
            Self::MoveDown { extend: false } => "move_down",
            Self::MoveLineStart { extend: false } => "move_line_start",
            Self::MoveLineEnd { extend: false } => "move_line_end",
            Self::MoveDocStart { extend: false } => "move_doc_start",
            Self::MoveDocEnd { extend: false } => "move_doc_end",
            Self::MoveLeft { extend: true } => "move_left_extend",
            Self::MoveRight { extend: true } => "move_right_extend",
            Self::MoveUp { extend: true } => "move_up_extend",
            Self::MoveDown { extend: true } => "move_down_extend",
            Self::MoveLineStart { extend: true } => "move_line_start_extend",
            Self::MoveLineEnd { extend: true } => "move_line_end_extend",
            Self::MoveDocStart { extend: true } => "move_doc_start_extend",
            Self::MoveDocEnd { extend: true } => "move_doc_end_extend",
            Self::SelectAll => "select_all",
            Self::PageUp => "page_up",
            Self::PageDown => "page_down",
            Self::ToggleTree => "toggle_tree",
            Self::ToggleTerminal => "toggle_terminal",
            Self::TerminalGrow => "terminal_grow",
            Self::TerminalShrink => "terminal_shrink",
            Self::TreeGrow => "tree_grow",
            Self::TreeShrink => "tree_shrink",
            Self::FocusTree => "focus_tree",
            Self::FocusEditor => "focus_editor",
            Self::FocusTerminal => "focus_terminal",
            Self::NextTab => "next_tab",
            Self::PrevTab => "prev_tab",
            Self::CloseTab => "close_tab",
            Self::NewTab => "new_tab",
            Self::CommandPalette => "command_palette",
            Self::OpenFileFuzzy => "open_file_fuzzy",
            Self::TreeNewFile => "tree_new_file",
            Self::TreeNewDir => "tree_new_dir",
            Self::TreeRefresh => "tree_refresh",
            Self::OpenFolder => "open_folder",
            Self::FocusToggleTreeEditor => "focus_toggle_tree_editor",
            Self::ToggleSoftWrap => "toggle_soft_wrap",
            Self::ToggleComment => "toggle_comment",
            Self::ToggleMdPreview => "toggle_md_preview",
            Self::Help => "help",
            Self::Find => "find",
            Self::FindNext => "find_next",
            Self::FindPrev => "find_prev",
            Self::ProjectFind => "project_find",
            Self::ProjectReplace => "project_replace",
            Self::Replace => "replace",
            Self::Copy => "copy",
            Self::Paste => "paste",
            Self::Cut => "cut",
            Self::ReloadFile => "reload_file",
            Self::LspComplete => "lsp_complete",
            Self::LspHover => "lsp_hover",
            Self::LspGotoDefinition => "lsp_goto_definition",
            Self::LspFormat => "lsp_format",
            Self::ToggleDiagnostics => "toggle_diagnostics",
            Self::SplitVertical => "split_vertical",
            Self::SplitHorizontal => "split_horizontal",
            Self::FocusNextPane => "focus_next_pane",
            Self::ClosePane => "close_pane",
            Self::ResizePaneGrow => "resize_pane_grow",
            Self::ResizePaneShrink => "resize_pane_shrink",
            Self::AddCursorAbove => "add_cursor_above",
            Self::AddCursorBelow => "add_cursor_below",
            Self::ClearExtraCursors => "clear_extra_cursors",
            Self::ToggleScm => "toggle_scm",
            Self::FocusScm => "focus_scm",
            Self::GitPull => "git_pull",
            Self::GitPush => "git_push",
            Self::BufferPicker => "buffer_picker",
            Self::JumpBack => "jump_back",
            Self::JumpForward => "jump_forward",
            Self::WhichKey => "which_key",
            Self::Welcome => "welcome",
            Self::ShowDiff => "show_diff",
            Self::Surround => "surround",
            Self::MultiPicker => "multi_picker",
            Self::UndoTree => "undo_tree",
            Self::ToggleMouse => "toggle_mouse",
            Self::SelectTheme => "select_theme",
            Self::SelectLocale => "select_locale",
            Self::HealthCheck => "health_check",
            Self::ToggleModal => "toggle_modal",
            Self::RunTasks => "run_tasks",
        }
    }

    #[must_use]
    pub fn palette_label(self) -> &'static str {
        match self {
            Self::Quit => "Quit",
            Self::Save => "Save",
            Self::SaveAs => "Save as…",
            Self::SaveAll => "Save all",
            Self::Undo => "Undo",
            Self::Redo => "Redo",
            Self::InsertNewline => "Insert newline",
            Self::InsertTab => "Insert tab",
            Self::Backspace => "Backspace",
            Self::Delete => "Delete",
            Self::MoveLeft { .. } => "Move left",
            Self::MoveRight { .. } => "Move right",
            Self::MoveUp { .. } => "Move up",
            Self::MoveDown { .. } => "Move down",
            Self::MoveLineStart { .. } => "Line start",
            Self::MoveLineEnd { .. } => "Line end",
            Self::MoveDocStart { .. } => "Document start",
            Self::MoveDocEnd { .. } => "Document end",
            Self::SelectAll => "Select all",
            Self::PageUp => "Page up",
            Self::PageDown => "Page down",
            Self::ToggleTree => "Toggle project tree",
            Self::ToggleTerminal => "Toggle terminal",
            Self::TerminalGrow => "Terminal taller",
            Self::TerminalShrink => "Terminal shorter",
            Self::FocusTree => "Focus tree",
            Self::FocusEditor => "Focus editor",
            Self::FocusTerminal => "Focus terminal",
            Self::NextTab => "Next tab",
            Self::PrevTab => "Previous tab",
            Self::CloseTab => "Close tab",
            Self::NewTab => "New tab",
            Self::CommandPalette => "Command palette",
            Self::OpenFileFuzzy => "Open file (fuzzy)",
            Self::TreeNewFile => "New file (tree)",
            Self::TreeNewDir => "New folder (tree)",
            Self::TreeRefresh => "Refresh tree",
            Self::OpenFolder => "Open folder…",
            Self::FocusToggleTreeEditor => "Focus: toggle tree / editor",
            Self::ToggleSoftWrap => "Toggle soft wrap",
            Self::ToggleComment => "Toggle comment",
            Self::ToggleMdPreview => "Toggle Markdown preview",
            Self::Help => "Show all keybindings",
            Self::Find => "Find…",
            Self::FindNext => "Find next",
            Self::FindPrev => "Find previous",
            Self::ProjectFind => "Find in project…",
            Self::ProjectReplace => "Replace in project…",
            Self::Replace => "Replace…",
            Self::Copy => "Copy",
            Self::Paste => "Paste",
            Self::Cut => "Cut",
            Self::ReloadFile => "Reload file from disk",
            Self::LspComplete => "LSP: complete",
            Self::LspHover => "LSP: hover",
            Self::LspGotoDefinition => "LSP: go to definition",
            Self::LspFormat => "LSP: format document",
            Self::ToggleDiagnostics => "Toggle diagnostics panel",
            Self::SplitVertical => "Split editor vertical",
            Self::SplitHorizontal => "Split editor horizontal",
            Self::FocusNextPane => "Focus next editor pane",
            Self::ClosePane => "Close editor pane",
            Self::ResizePaneGrow => "Split / Pane wider",
            Self::ResizePaneShrink => "Split / Pane narrower",
            Self::TreeGrow => "File tree wider",
            Self::TreeShrink => "File tree narrower",
            Self::AddCursorAbove => "Add cursor above",
            Self::AddCursorBelow => "Add cursor below",
            Self::ClearExtraCursors => "Clear extra cursors",
            Self::ToggleScm => "Toggle SCM panel",
            Self::FocusScm => "Focus SCM panel",
            Self::GitPull => "Git pull (fetch & merge)",
            Self::GitPush => "Git push",
            Self::BufferPicker => "Buffer picker…",
            Self::JumpBack => "Jump back",
            Self::JumpForward => "Jump forward",
            Self::WhichKey => "Which-key (shortcuts)",
            Self::Welcome => "Essential shortcuts",
            Self::ShowDiff => "Git diff (active file)",
            Self::Surround => "Surround selection…",
            Self::MultiPicker => "Multi picker (files/cmds/buffers)",
            Self::UndoTree => "Undo history…",
            Self::ToggleMouse => "Enable / disable mouse",
            Self::SelectTheme => "Preferences: Color Theme",
            Self::SelectLocale => "Preferences: Display Language",
            Self::HealthCheck => "System & LSP health check (:health)",
            Self::ToggleModal => "Toggle modal mode (Vim / CUA)",
            Self::RunTasks => "Run task…",
        }
    }

    pub fn palette_actions() -> &'static [Action] {
        &[
            Action::Save,
            Action::SaveAs,
            Action::SaveAll,
            Action::Undo,
            Action::Redo,
            Action::Find,
            Action::FindNext,
            Action::FindPrev,
            Action::ProjectFind,
            Action::ProjectReplace,
            Action::Replace,
            Action::SelectAll,
            Action::Copy,
            Action::Paste,
            Action::Cut,
            Action::ToggleComment,
            Action::ToggleSoftWrap,
            Action::ToggleMdPreview,
            Action::NewTab,
            Action::CloseTab,
            Action::NextTab,
            Action::PrevTab,
            Action::OpenFolder,
            Action::OpenFileFuzzy,
            Action::CommandPalette,
            Action::Help,
            Action::ToggleTree,
            Action::ToggleTerminal,
            Action::TerminalGrow,
            Action::TerminalShrink,
            Action::TreeGrow,
            Action::TreeShrink,
            Action::FocusTree,
            Action::FocusEditor,
            Action::FocusToggleTreeEditor,
            Action::FocusTerminal,
            Action::TreeNewFile,
            Action::TreeNewDir,
            Action::TreeRefresh,
            Action::ReloadFile,
            Action::LspComplete,
            Action::LspHover,
            Action::LspGotoDefinition,
            Action::LspFormat,
            Action::ToggleDiagnostics,
            Action::SplitVertical,
            Action::SplitHorizontal,
            Action::FocusNextPane,
            Action::ClosePane,
            Action::ResizePaneGrow,
            Action::ResizePaneShrink,
            Action::AddCursorAbove,
            Action::AddCursorBelow,
            Action::ClearExtraCursors,
            Action::ToggleScm,
            Action::FocusScm,
            Action::GitPull,
            Action::GitPush,
            Action::BufferPicker,
            Action::JumpBack,
            Action::JumpForward,
            Action::WhichKey,
            Action::Welcome,
            Action::ShowDiff,
            Action::Surround,
            Action::MultiPicker,
            Action::UndoTree,
            Action::ToggleMouse,
            Action::SelectTheme,
            Action::SelectLocale,
            Action::HealthCheck,
            Action::ToggleModal,
            Action::RunTasks,
            Action::Quit,
        ]
    }
}

#[derive(Debug, Error, PartialEq, Eq)]
#[error("unknown action id: {0}")]
pub struct ActionParseError(pub String);

pub fn parse_action(id: &str) -> Result<Action, ActionParseError> {
    let action = match id {
        "quit" => Action::Quit,
        "save" => Action::Save,
        "save_as" => Action::SaveAs,
        "save_all" => Action::SaveAll,
        "undo" => Action::Undo,
        "redo" => Action::Redo,
        "insert_newline" => Action::InsertNewline,
        "insert_tab" => Action::InsertTab,
        "backspace" => Action::Backspace,
        "delete" => Action::Delete,
        "move_left" => Action::MoveLeft { extend: false },
        "move_right" => Action::MoveRight { extend: false },
        "move_up" => Action::MoveUp { extend: false },
        "move_down" => Action::MoveDown { extend: false },
        "move_line_start" => Action::MoveLineStart { extend: false },
        "move_line_end" => Action::MoveLineEnd { extend: false },
        "move_left_extend" => Action::MoveLeft { extend: true },
        "move_right_extend" => Action::MoveRight { extend: true },
        "move_up_extend" => Action::MoveUp { extend: true },
        "move_down_extend" => Action::MoveDown { extend: true },
        "move_line_start_extend" => Action::MoveLineStart { extend: true },
        "move_line_end_extend" => Action::MoveLineEnd { extend: true },
        "move_doc_start" => Action::MoveDocStart { extend: false },
        "move_doc_end" => Action::MoveDocEnd { extend: false },
        "move_doc_start_extend" => Action::MoveDocStart { extend: true },
        "move_doc_end_extend" => Action::MoveDocEnd { extend: true },
        "select_all" => Action::SelectAll,
        "page_up" => Action::PageUp,
        "page_down" => Action::PageDown,
        "toggle_tree" => Action::ToggleTree,
        "toggle_terminal" => Action::ToggleTerminal,
        "terminal_grow" => Action::TerminalGrow,
        "terminal_shrink" => Action::TerminalShrink,
        "tree_grow" => Action::TreeGrow,
        "tree_shrink" => Action::TreeShrink,
        "focus_tree" => Action::FocusTree,
        "focus_editor" => Action::FocusEditor,
        "focus_terminal" => Action::FocusTerminal,
        "next_tab" => Action::NextTab,
        "prev_tab" => Action::PrevTab,
        "close_tab" => Action::CloseTab,
        "new_tab" => Action::NewTab,
        "command_palette" => Action::CommandPalette,
        "open_file_fuzzy" => Action::OpenFileFuzzy,
        "tree_new_file" => Action::TreeNewFile,
        "tree_new_dir" => Action::TreeNewDir,
        "tree_refresh" => Action::TreeRefresh,
        "open_folder" => Action::OpenFolder,
        "focus_toggle_tree_editor" => Action::FocusToggleTreeEditor,
        "toggle_soft_wrap" => Action::ToggleSoftWrap,
        "toggle_comment" => Action::ToggleComment,
        "toggle_md_preview" => Action::ToggleMdPreview,
        "help" => Action::Help,
        "find" => Action::Find,
        "find_next" => Action::FindNext,
        "find_prev" => Action::FindPrev,
        "project_find" => Action::ProjectFind,
        "project_replace" => Action::ProjectReplace,
        "replace" => Action::Replace,
        "copy" => Action::Copy,
        "paste" => Action::Paste,
        "cut" => Action::Cut,
        "reload_file" => Action::ReloadFile,
        "lsp_complete" => Action::LspComplete,
        "lsp_hover" => Action::LspHover,
        "lsp_goto_definition" => Action::LspGotoDefinition,
        "lsp_format" => Action::LspFormat,
        "toggle_diagnostics" => Action::ToggleDiagnostics,
        "split_vertical" => Action::SplitVertical,
        "split_horizontal" => Action::SplitHorizontal,
        "focus_next_pane" => Action::FocusNextPane,
        "close_pane" => Action::ClosePane,
        "resize_pane_grow" => Action::ResizePaneGrow,
        "resize_pane_shrink" => Action::ResizePaneShrink,
        "add_cursor_above" => Action::AddCursorAbove,
        "add_cursor_below" => Action::AddCursorBelow,
        "clear_extra_cursors" => Action::ClearExtraCursors,
        "toggle_scm" => Action::ToggleScm,
        "focus_scm" => Action::FocusScm,
        "git_pull" => Action::GitPull,
        "git_push" => Action::GitPush,
        "buffer_picker" => Action::BufferPicker,
        "jump_back" => Action::JumpBack,
        "jump_forward" => Action::JumpForward,
        "which_key" => Action::WhichKey,
        "welcome" => Action::Welcome,
        "show_diff" => Action::ShowDiff,
        "surround" => Action::Surround,
        "multi_picker" => Action::MultiPicker,
        "undo_tree" => Action::UndoTree,
        "toggle_mouse" => Action::ToggleMouse,
        "select_theme" => Action::SelectTheme,
        "select_locale" | "display_language" => Action::SelectLocale,
        "health_check" | "checkhealth" | "health" => Action::HealthCheck,
        "toggle_modal" | "modal_mode" => Action::ToggleModal,
        "run_tasks" | "tasks" => Action::RunTasks,
        other => return Err(ActionParseError(other.to_string())),
    };
    Ok(action)
}

#[cfg(test)]
mod tests {
    use std::collections::HashSet;

    use super::*;

    #[test]
    fn id_roundtrips_through_parse_action() {
        for action in ALL {
            let id = action.id();
            let parsed =
                parse_action(id).unwrap_or_else(|error| panic!("id `{id}` não parseia: {error}"));
            assert_eq!(
                parsed, *action,
                "id `{id}` volta como {parsed:?} em vez de {action:?}"
            );
        }
    }

    #[test]
    fn all_covers_every_variant_exactly_once() {
        let unique: HashSet<&'static str> = ALL.iter().map(|action| action.id()).collect();
        assert_eq!(
            unique.len(),
            ALL.len(),
            "ALL tem id repetido: {} ids para {} ações",
            unique.len(),
            ALL.len()
        );
    }

    #[test]
    fn parse_action_rejects_unknown_ids() {
        assert!(parse_action("does_not_exist").is_err());
        assert!(parse_action("").is_err());
    }

    #[test]
    fn aliases_and_canonical_ids_agree() {
        // Aliases documentados devem resolver para a mesma ação do id canônico.
        for (alias, canonical) in [
            ("display_language", Action::SelectLocale),
            ("modal_mode", Action::ToggleModal),
            ("tasks", Action::RunTasks),
            ("checkhealth", Action::HealthCheck),
        ] {
            assert_eq!(parse_action(alias).unwrap(), canonical);
            assert_eq!(parse_action(alias).unwrap().id(), canonical.id());
        }
    }

    #[test]
    fn every_palette_action_exists_and_is_listed_once() {
        let mut seen = HashSet::new();
        for action in Action::palette_actions() {
            assert!(
                ALL.contains(action),
                "{action:?} está na palette mas fora de ALL"
            );
            assert!(seen.insert(action.id()), "{action:?} duplicada na palette");
        }
    }

    #[test]
    fn every_action_has_a_label() {
        for action in ALL {
            assert!(
                !action.palette_label().is_empty(),
                "{action:?} sem palette_label"
            );
        }
    }

    #[test]
    fn parses_polish_and_lsp_actions() {
        assert_eq!(parse_action("find").unwrap(), Action::Find);
        assert_eq!(parse_action("lsp_format").unwrap(), Action::LspFormat);
        assert_eq!(parse_action("git_pull").unwrap(), Action::GitPull);
        assert_eq!(parse_action("git_push").unwrap(), Action::GitPush);
        assert_eq!(
            parse_action("toggle_diagnostics").unwrap(),
            Action::ToggleDiagnostics
        );
        assert_eq!(
            parse_action("project_replace").unwrap(),
            Action::ProjectReplace
        );
        assert_eq!(
            parse_action("resize_pane_grow").unwrap(),
            Action::ResizePaneGrow
        );
        assert_eq!(parse_action("tree_grow").unwrap(), Action::TreeGrow);
        assert_eq!(parse_action("select_theme").unwrap(), Action::SelectTheme);
        assert_eq!(parse_action("select_locale").unwrap(), Action::SelectLocale);
        assert_eq!(parse_action("health").unwrap(), Action::HealthCheck);
        assert_eq!(parse_action("toggle_modal").unwrap(), Action::ToggleModal);
        assert_eq!(parse_action("tasks").unwrap(), Action::RunTasks);
    }
}
