//! Definição da menu bar (File / Edit / View / Go / Git / Help).

use oride_i18n::Locale;
use oride_ui::{MenuColumn, MenuItem};

/// Constrói menus localizados de acordo com o `Locale`.
#[must_use]
pub fn menus_for_locale(locale: &Locale) -> Vec<MenuColumn> {
    let msg = locale.messages();
    vec![
        MenuColumn {
            title: msg.file().into(),
            hotkey: 'f',
            items: vec![
                item(msg.new_tab(), "ctrl+n", "new_tab"),
                item(msg.open_file(), "ctrl+p", "open_file_fuzzy"),
                item(msg.open_folder(), "ctrl+o", "open_folder"),
                item(msg.save(), "ctrl+s", "save"),
                item(msg.save_as(), "f12", "save_as"),
                item(msg.save_all(), "ctrl+alt+s", "save_all"),
                item(msg.reload_file(), "ctrl+r", "reload_file"),
                item(msg.quit(), "ctrl+q", "quit"),
            ],
        },
        MenuColumn {
            title: msg.edit().into(),
            hotkey: 'e',
            items: vec![
                item(msg.undo(), "ctrl+z", "undo"),
                item(msg.redo(), "ctrl+y", "redo"),
                item(msg.cut(), "ctrl+x", "cut"),
                item(msg.copy(), "ctrl+c", "copy"),
                item(msg.paste(), "ctrl+v", "paste"),
                item(msg.select_all(), "ctrl+a", "select_all"),
                item(msg.find(), "ctrl+f", "find"),
                item(msg.find_in_project(), "ctrl+shift+f", "project_find"),
                item(msg.replace(), "ctrl+h", "replace"),
                item(msg.toggle_comment(), "ctrl+/", "toggle_comment"),
            ],
        },
        MenuColumn {
            title: msg.view().into(),
            hotkey: 'v',
            items: vec![
                item(msg.command_palette(), "ctrl+shift+p", "command_palette"),
                item(msg.toggle_tree(), "ctrl+shift+b", "toggle_tree"),
                item(msg.toggle_terminal(), "ctrl+\"", "toggle_terminal"),
                item(msg.toggle_scm(), "ctrl+shift+g", "toggle_scm"),
                item("MD preview", "alt+p", "toggle_md_preview"),
                item("Soft wrap", "alt+z", "toggle_soft_wrap"),
                item("Split vertical", "ctrl+alt+v", "split_vertical"),
                item("Split horizontal", "ctrl+alt+h", "split_horizontal"),
                item("Which-key", "alt+/", "which_key"),
                item(msg.color_theme(), "palette", "select_theme"),
                item(msg.display_language(), "palette", "select_locale"),
                item(msg.toggle_mouse(), "palette", "toggle_mouse"),
            ],
        },
        MenuColumn {
            title: msg.go().into(),
            hotkey: 'g',
            items: vec![
                item("Buffer picker…", "ctrl+shift+o", "buffer_picker"),
                item("Jump back", "ctrl+alt+o", "jump_back"),
                item("Jump forward", "ctrl+alt+i", "jump_forward"),
                item("LSP: definition", "f4", "lsp_goto_definition"),
                item("LSP: hover", "ctrl+k", "lsp_hover"),
                item("Diagnostics", "ctrl+shift+m", "toggle_diagnostics"),
                item("Next tab", "ctrl+pagedown", "next_tab"),
                item("Prev tab", "ctrl+pageup", "prev_tab"),
            ],
        },
        MenuColumn {
            title: msg.git().into(),
            hotkey: 'i',
            items: vec![
                item(msg.toggle_scm(), "ctrl+shift+g", "toggle_scm"),
                item("Git pull", "", "git_pull"),
                item("Git push", "", "git_push"),
                item("Diff active file", "f2", "show_diff"),
                item("Refresh git", "f5", "tree_refresh"),
                item("Show path", "", "plugin:show_path"),
            ],
        },
        MenuColumn {
            title: msg.help().into(),
            hotkey: 'h',
            items: vec![
                item("All keybindings…", "f1", "help"),
                item("Essential shortcuts", "alt+shift+/", "welcome"),
                item("Which-key", "alt+/", "which_key"),
                item(msg.command_palette(), "ctrl+shift+p", "command_palette"),
                item("Multi picker…", "ctrl+shift+t", "multi_picker"),
                item("Surround selection…", "f8", "surround"),
                item("Undo history…", "ctrl+shift+u", "undo_tree"),
                item(msg.display_language(), "palette", "select_locale"),
                item("Word count", "", "plugin:word_count"),
            ],
        },
    ]
}

fn item(label: &str, shortcut: &str, action_id: &str) -> MenuItem {
    MenuItem {
        label: label.into(),
        shortcut: shortcut.into(),
        action_id: action_id.into(),
    }
}
