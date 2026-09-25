//! Tratamento de eventos de mouse, hit testing e seleção por clique/arrasto.

use std::time::Instant;

use crossterm::event::{KeyModifiers, MouseButton, MouseEvent, MouseEventKind};

use crate::mouse::{
    accurate_tab_index_at, is_multi_click, list_row_at, menu_index_at, screen_to_caret,
    tab_index_at, tree_row_at, word_bounds, HitTarget,
};
use crate::split::SplitOrientation;

use super::state::{App, Focus, Overlay, SplitterDrag};

impl App {
    /// Mouse ativo (configuração / menu). Default: false.
    #[must_use]
    pub fn mouse_is_enabled(&self) -> bool {
        self.mouse_enabled
    }

    /// Em arrasto de seleção ou redimensionamento de divisor.
    #[must_use]
    pub fn is_mouse_dragging(&self) -> bool {
        self.mouse_enabled && (self.mouse_drag_anchor.is_some() || self.splitter_drag.is_some())
    }

    /// Eventos de mouse (clique, arrasto, scroll, foco de painéis).
    pub fn handle_mouse(&mut self, event: MouseEvent) {
        if !self.mouse_enabled {
            return;
        }
        if !matches!(self.overlay, Overlay::None) {
            if matches!(event.kind, MouseEventKind::Down(MouseButton::Left))
                && matches!(self.overlay, Overlay::Hover { .. } | Overlay::SurroundPick)
            {
                self.overlay = Overlay::None;
                self.surround_pending = false;
            }
            return;
        }
        if self.show_which_key || self.show_welcome {
            if matches!(event.kind, MouseEventKind::Down(MouseButton::Left)) {
                self.show_which_key = false;
                self.show_welcome = false;
            }
            return;
        }
        if self.menu_open.is_some() {
            self.handle_mouse_menu(event);
            return;
        }

        let mouse_x = event.column;
        let mouse_y = event.row;
        let target = self.hit_regions.at(mouse_x, mouse_y);

        match event.kind {
            MouseEventKind::ScrollUp | MouseEventKind::ScrollDown => {
                let is_scroll_up = matches!(event.kind, MouseEventKind::ScrollUp);
                match target {
                    HitTarget::Editor => {
                        if is_scroll_up {
                            self.scroll_y = self.scroll_y.saturating_sub(3);
                        } else {
                            self.scroll_y = self.scroll_y.saturating_add(3);
                        }
                        self.split.sync_scroll(self.scroll_y);
                    }
                    HitTarget::Tree => {
                        if let Some(tree) = self.tree.as_mut() {
                            if is_scroll_up {
                                tree.move_selection(-3);
                            } else {
                                tree.move_selection(3);
                            }
                            self.ensure_tree_visible();
                        }
                    }
                    HitTarget::Scm => {
                        if is_scroll_up {
                            self.scm_selected = self.scm_selected.saturating_sub(1);
                        } else if !self.scm_cache.is_empty() {
                            self.scm_selected =
                                (self.scm_selected + 1).min(self.scm_cache.len() - 1);
                        }
                    }
                    HitTarget::Terminal => {
                        self.focus = Focus::Terminal;
                    }
                    HitTarget::Preview => {
                        if is_scroll_up {
                            self.preview_scroll = self.preview_scroll.saturating_sub(3);
                        } else {
                            self.preview_scroll = self.preview_scroll.saturating_add(3);
                        }
                    }
                    _ => {}
                }
            }
            MouseEventKind::Down(MouseButton::Left) => {
                let current_time = Instant::now();
                let is_multiple_click = is_multi_click(self.last_click, mouse_x, mouse_y, 450);
                if is_multiple_click {
                    self.click_count = self.click_count.saturating_add(1).min(3);
                } else {
                    self.click_count = 1;
                }
                self.last_click = Some((current_time, mouse_x, mouse_y));
                self.mouse_down_at(target, mouse_x, mouse_y, self.click_count, event.modifiers);
            }
            MouseEventKind::Drag(MouseButton::Left) => {
                if let Some(drag) = self.splitter_drag {
                    match drag {
                        SplitterDrag::Tree {
                            start_x,
                            initial_width,
                        } => {
                            let delta = mouse_x as i32 - start_x as i32;
                            let new_width = (initial_width as i32 + delta).clamp(12, 80) as u16;
                            self.tree_width = new_width;
                        }
                        SplitterDrag::Editor {
                            start_coord,
                            initial_ratio,
                            total_size,
                            orientation,
                        } => {
                            let current_coord = if orientation == SplitOrientation::Vertical {
                                mouse_x
                            } else {
                                mouse_y
                            };
                            let delta_pixels = current_coord as i32 - start_coord as i32;
                            let delta_percent = (delta_pixels * 100) / (total_size.max(1) as i32);
                            let new_ratio =
                                (initial_ratio as i32 + delta_percent).clamp(15, 85) as u16;
                            self.split.ratio_percent = new_ratio;
                        }
                    }
                } else if self.mouse_drag_anchor.is_some() || target == HitTarget::Editor {
                    // Continua o arrasto mesmo se o cursor saiu um pouco da área do editor
                    self.mouse_drag_to(mouse_x, mouse_y);
                }
            }
            MouseEventKind::Up(MouseButton::Left) => {
                // Não limpa seleção ativa — apenas encerra o modo de arrasto
                self.mouse_drag_anchor = None;
                self.splitter_drag = None;
            }
            MouseEventKind::Down(MouseButton::Right) => {
                self.show_which_key = true;
            }
            _ => {}
        }
    }

    pub(crate) fn handle_mouse_menu(&mut self, event: MouseEvent) {
        // 1. Se o evento ocorreu na barra de menus no topo
        if event.row == self.hit_regions.menu.y {
            let titles: Vec<&str> = self.menus.iter().map(|menu| menu.title.as_str()).collect();
            if let Some(menu_index) = menu_index_at(&titles, self.hit_regions.menu, event.column) {
                if matches!(event.kind, MouseEventKind::Down(MouseButton::Left)) {
                    if self.menu_open.map(|(idx, _)| idx) == Some(menu_index) {
                        self.menu_open = None;
                    } else {
                        self.menu_open = Some((menu_index, 0));
                    }
                } else if matches!(
                    event.kind,
                    MouseEventKind::Moved | MouseEventKind::Drag(MouseButton::Left)
                ) && self.menu_open.map(|(idx, _)| idx) != Some(menu_index)
                {
                    self.menu_open = Some((menu_index, 0));
                }
                return;
            }
        }

        // 2. Se o evento ocorreu dentro do dropdown aberto
        if let Some((menu_index, _)) = self.menu_open {
            if let Some(menu) = self.menus.get(menu_index) {
                if let Some(rect) = self.hit_regions.menu_dropdown {
                    let min_y = rect.y.saturating_add(1);
                    let max_y = rect.y.saturating_add(rect.height.saturating_sub(1));
                    if event.column >= rect.x
                        && event.column < rect.x.saturating_add(rect.width)
                        && event.row >= min_y
                        && event.row < max_y
                    {
                        let item_index = (event.row - min_y) as usize;
                        if item_index < menu.items.len() {
                            if matches!(
                                event.kind,
                                MouseEventKind::Down(MouseButton::Left)
                                    | MouseEventKind::Up(MouseButton::Left)
                            ) {
                                let action_id = menu.items[item_index].action_id.clone();
                                self.menu_open = None;
                                self.apply_action_id(&action_id);
                                return;
                            } else if matches!(
                                event.kind,
                                MouseEventKind::Moved | MouseEventKind::Drag(MouseButton::Left)
                            ) {
                                self.menu_open = Some((menu_index, item_index));
                                return;
                            }
                        }
                    }
                }
            }
        }

        // Clicou fora do menu e do dropdown: apenas se for clique (Down) fecha
        if matches!(event.kind, MouseEventKind::Down(MouseButton::Left)) {
            self.menu_open = None;
        }
    }

    pub(crate) fn mouse_down_at(
        &mut self,
        target: HitTarget,
        mouse_x: u16,
        mouse_y: u16,
        click_count: u8,
        modifiers: KeyModifiers,
    ) {
        match target {
            HitTarget::Menu => {
                let titles: Vec<&str> = self.menus.iter().map(|menu| menu.title.as_str()).collect();
                if let Some(menu_index) = menu_index_at(&titles, self.hit_regions.menu, mouse_x) {
                    self.menu_open = Some((menu_index, 0));
                }
            }
            HitTarget::Tree => {
                self.focus = Focus::Tree;
                self.show_tree = true;
                if let (Some(area), Some(tree)) = (self.hit_regions.tree, self.tree.as_mut()) {
                    if let Some(row) = tree_row_at(area, self.tree_scroll, mouse_y) {
                        tree.set_selected(row);
                        if click_count >= 2 {
                            match tree.activate_selected() {
                                Ok(Some(path)) => {
                                    let _ = self.open_document_path(&path);
                                    self.focus = Focus::Editor;
                                    self.scroll_y = 0;
                                }
                                Ok(None) => {}
                                Err(error) => self.set_status(format!("tree: {error}")),
                            }
                        }
                    }
                }
            }
            HitTarget::Tabs => {
                if let Some(area) = self.hit_regions.tabs {
                    let summaries = self.store.tab_summaries();
                    let tab_index = accurate_tab_index_at(area, &summaries, mouse_x)
                        .or_else(|| tab_index_at(area, summaries.len(), mouse_x));
                    if let Some(index) = tab_index {
                        let tab_id = summaries.get(index).map(|tab| tab.id);
                        if let Some(document_id) = tab_id {
                            let _ = self.store.set_active(document_id);
                            self.split.set_focused_doc(document_id);
                            self.scroll_y = 0;
                            self.focus = Focus::Editor;
                        }
                    }
                }
            }
            HitTarget::Editor => {
                self.focus = Focus::Editor;
                self.place_caret_from_mouse(mouse_x, mouse_y, click_count, modifiers);
            }
            HitTarget::Scm => {
                self.show_scm = true;
                self.focus = Focus::Scm;
                if let Some(area) = self.hit_regions.scm {
                    if let Some(row) = list_row_at(area, 0, mouse_y) {
                        if row < self.scm_cache.len() {
                            self.scm_selected = row;
                        }
                        if click_count >= 2 {
                            if let Some((_, relative_path)) =
                                self.scm_cache.get(self.scm_selected).cloned()
                            {
                                let full_path = self.workspace.join(relative_path);
                                let _ = self.open_document_path(&full_path);
                                self.focus = Focus::Editor;
                            }
                        }
                    }
                }
            }
            HitTarget::Terminal => {
                self.ensure_terminal();
                if let Some(terminal) = self.terminal.as_mut() {
                    terminal.visible = true;
                }
                self.focus = Focus::Terminal;
            }
            HitTarget::Preview => {
                self.handle_preview_click(mouse_x, mouse_y);
            }
            HitTarget::TreeSplitter => {
                self.splitter_drag = Some(SplitterDrag::Tree {
                    start_x: mouse_x,
                    initial_width: self.tree_width,
                });
                self.mouse_drag_anchor = None;
            }
            HitTarget::EditorSplitter => {
                let orientation = self.split.orientation;
                let (start_coord, total_size) = if orientation == SplitOrientation::Vertical {
                    let total = self.hit_regions.editor.map(|r| r.width).unwrap_or(80);
                    (mouse_x, total)
                } else {
                    let total = self.hit_regions.editor.map(|r| r.height).unwrap_or(24);
                    (mouse_y, total)
                };
                self.splitter_drag = Some(SplitterDrag::Editor {
                    start_coord,
                    initial_ratio: self.split.ratio_percent,
                    total_size,
                    orientation,
                });
                self.mouse_drag_anchor = None;
            }
            HitTarget::Outside => {}
        }
    }

    pub(crate) fn place_caret_from_mouse(
        &mut self,
        mouse_x: u16,
        mouse_y: u16,
        click_count: u8,
        modifiers: KeyModifiers,
    ) {
        let Ok(document) = self.store.active() else {
            return;
        };
        let Some(caret) = screen_to_caret(document.buffer(), &self.hit_regions, mouse_x, mouse_y)
        else {
            return;
        };
        let Ok(byte_offset) = document.buffer().caret_to_byte(caret) else {
            return;
        };
        let caret_line = caret.line;
        let _ = document;

        if click_count >= 3 {
            // Seleção de linha inteira
            if let Ok(document) = self.store.active_mut() {
                let start_offset = document
                    .buffer()
                    .line_to_byte(caret_line)
                    .unwrap_or(byte_offset);
                let end_offset = if caret_line + 1 < document.buffer().line_count() {
                    document
                        .buffer()
                        .line_to_byte(caret_line + 1)
                        .unwrap_or(oride_core::ByteOffset::new(document.buffer().len_bytes()))
                } else {
                    oride_core::ByteOffset::new(document.buffer().len_bytes())
                };
                document.select_byte_range(start_offset, end_offset);
            }
            self.mouse_drag_anchor = None;
        } else if click_count >= 2 {
            // Seleção de palavra
            if let Ok(document) = self.store.active_mut() {
                let (start_offset, end_offset) = word_bounds(document.buffer(), byte_offset);
                document.select_byte_range(start_offset, end_offset);
            }
            self.mouse_drag_anchor = None;
        } else if modifiers.contains(KeyModifiers::ALT) {
            let cursor_count = if let Ok(document) = self.store.active_mut() {
                document.add_cursor_at(byte_offset);
                Some(document.extra_carets().len() + 1)
            } else {
                None
            };
            if let Some(count) = cursor_count {
                self.set_status(format!("multi-cursor: {count} cursores ativos"));
            }
            self.mouse_drag_anchor = None;
        } else {
            if let Ok(document) = self.store.active_mut() {
                document.jump_to_byte(byte_offset);
            }
            self.mouse_drag_anchor = Some(byte_offset);
        }
        self.ensure_cursor_visible();
    }

    pub(crate) fn mouse_drag_to(&mut self, mouse_x: u16, mouse_y: u16) {
        let Some(anchor_offset) = self.mouse_drag_anchor else {
            return;
        };
        let Ok(document) = self.store.active() else {
            return;
        };
        let Some(caret) = screen_to_caret(document.buffer(), &self.hit_regions, mouse_x, mouse_y)
        else {
            return;
        };
        let Ok(head_offset) = document.buffer().caret_to_byte(caret) else {
            return;
        };
        let _ = document;
        if let Ok(document) = self.store.active_mut() {
            document.set_selection_live(anchor_offset, head_offset);
        }
        self.ensure_cursor_visible();
    }

    pub(crate) fn handle_preview_click(&mut self, mouse_x: u16, mouse_y: u16) {
        let Some(rect) = self.hit_regions.preview else {
            return;
        };
        if mouse_x <= rect.x || mouse_x >= rect.x.saturating_add(rect.width).saturating_sub(1) {
            return;
        }
        if mouse_y <= rect.y || mouse_y >= rect.y.saturating_add(rect.height).saturating_sub(1) {
            return;
        }
        let line_offset = (mouse_y - rect.y - 1) as usize;
        let col = (mouse_x - rect.x - 1) as usize;

        if let Some((_, _, _, ref lines)) = self.cached_preview {
            let max_scroll = lines.len().saturating_sub(1);
            let scroll = self
                .scroll_y
                .min(max_scroll)
                .saturating_add(self.preview_scroll)
                .min(max_scroll);
            let idx = scroll + line_offset;
            if let Some(line) = lines.get(idx) {
                if let Some(link) = line.link_at_col(col).or_else(|| line.first_link()) {
                    let target = link.url.clone();
                    match open_url_or_path(&target) {
                        Ok(()) => self.set_status(format!("Abrindo: {target}")),
                        Err(err) => self.set_status(format!("Falha ao abrir {target}: {err}")),
                    }
                }
            }
        }
    }

    pub(crate) fn open_active_preview_link(&mut self) {
        if let Some((_, _, _, ref lines)) = self.cached_preview {
            let caret_line = self
                .store
                .active()
                .ok()
                .and_then(|doc| doc.caret().ok())
                .map(|c| c.line)
                .unwrap_or(0);
            let max_scroll = lines.len().saturating_sub(1);
            let scroll = self
                .scroll_y
                .min(max_scroll)
                .saturating_add(self.preview_scroll)
                .min(max_scroll);
            let target_line = lines.get(caret_line).or_else(|| lines.get(scroll));
            if let Some(line) = target_line {
                if let Some(link) = line.first_link() {
                    let target = link.url.clone();
                    match open_url_or_path(&target) {
                        Ok(()) => self.set_status(format!("Abrindo: {target}")),
                        Err(err) => self.set_status(format!("Falha ao abrir {target}: {err}")),
                    }
                } else {
                    self.set_status("Nenhum link na linha atual");
                }
            }
        }
    }
}

/// Abre uma URL ou caminho no navegador/aplicativo padrão do sistema operacional de forma desacoplada.
pub fn open_url_or_path(target: &str) -> std::io::Result<()> {
    // A escolha do abridor por plataforma vive em `oride-osutil`, para não haver
    // duas listas de sistemas operacionais para manter em sincronia.
    oride_osutil::open_target(target)
}
