# Keymap

**Canonical source.** This file is the project's only key table. The prose guides
— `docs/polish.md`, `docs/planning/ux-polish-plan.md`, the usage guides —
point here instead of repeating the table. Repeating it is how the divergences
appeared: `:normal` instead of `modal_mode`, `modal_editing` in place of the real id,
and a `:help` that does not exist. All were fixed in M2, and all were of the same class.

The sections group the keys by **domain**, derived from the id prefix by a
deterministic rule — a new id lands in the right group by its own name, with no second
table to maintain. What holds the (key, id) pairs is
`TestDocumentedKeymapMatchesTheBindings`, which compares both directions: a key in
the code and missing here, or here and missing from the code, fails by listing exactly the difference.

**Remapping is configuration, not code.** See `docs/guides/en/config.md`.

<!-- GERADO: keymap -->
### File

| Key | Action |
|---|---|
| `Alt+left` | `prev_tab` |
| `Alt+right` | `next_tab` |
| `Alt+Shift+s` | `save_as` |
| `Ctrl+Alt+s` | `save_all` |
| `Ctrl+n` | `new_tab` |
| `Ctrl+o` | `open_folder` |
| `Ctrl+p` | `open_file_fuzzy` |
| `Ctrl+pagedown` | `next_tab` |
| `Ctrl+pageup` | `prev_tab` |
| `Ctrl+r` | `reload_file` |
| `Ctrl+s` | `save` |
| `Ctrl+Shift+[` | `prev_tab` |
| `Ctrl+Shift+]` | `next_tab` |
| `Ctrl+Shift+s` | `save_as` |
| `Ctrl+w` | `close_tab` |
| `f12` | `save_as` |

### Editing

| Key | Action |
|---|---|
| `backspace` | `backspace` |
| `Ctrl+/` | `toggle_comment` |
| `Ctrl+c` | `copy` |
| `Ctrl+Shift+u` | `undo_tree` |
| `Ctrl+v` | `paste` |
| `Ctrl+x` | `cut` |
| `Ctrl+y` | `redo` |
| `Ctrl+z` | `undo` |
| `enter` | `insert_newline` |
| `f8` | `surround` |
| `tab` | `insert_tab` |

### Movement

| Key | Action |
|---|---|
| `Ctrl+end` | `move_doc_end` |
| `Ctrl+home` | `move_doc_start` |
| `Ctrl+Shift+end` | `move_doc_end_extend` |
| `Ctrl+Shift+home` | `move_doc_start_extend` |
| `down` | `move_down` |
| `end` | `move_line_end` |
| `home` | `move_line_start` |
| `left` | `move_left` |
| `right` | `move_right` |
| `Shift+down` | `move_down_extend` |
| `Shift+end` | `move_line_end_extend` |
| `Shift+home` | `move_line_start_extend` |
| `Shift+left` | `move_left_extend` |
| `Shift+right` | `move_right_extend` |
| `Shift+up` | `move_up_extend` |
| `up` | `move_up` |

### Selection

| Key | Action |
|---|---|
| `Ctrl+a` | `select_all` |
| `Ctrl+Alt+down` | `add_cursor_below` |
| `Ctrl+Alt+u` | `clear_extra_cursors` |
| `Ctrl+Alt+up` | `add_cursor_above` |

### Search

| Key | Action |
|---|---|
| `Ctrl+f` | `find` |
| `Ctrl+h` | `replace` |
| `Ctrl+Shift+f` | `project_find` |
| `Ctrl+Shift+h` | `replace` |
| `f3` | `find_next` |
| `Shift+f3` | `find_prev` |

### Layout

| Key | Action |
|---|---|
| `Alt+p` | `toggle_md_preview` |
| `Ctrl+\` | `focus_tree` |
| `Ctrl+Alt+h` | `split_horizontal` |
| `Ctrl+Alt+left` | `focus_next_pane` |
| `Ctrl+Alt+right` | `focus_next_pane` |
| `Ctrl+Alt+v` | `split_vertical` |
| `Ctrl+b` | `focus_tree` |
| `Ctrl+e` | `focus_editor` |
| `Ctrl+Shift+b` | `toggle_tree` |
| `Ctrl+Shift+e` | `focus_toggle_tree_editor` |
| `Ctrl+Shift+g` | `toggle_scm` |
| `Ctrl+Shift+v` | `toggle_md_preview` |
| `f6` | `focus_next_pane` |

### Terminal

| Key | Action |
|---|---|
| `Alt+-` | `terminal_shrink` |
| `Alt+=` | `terminal_grow` |
| `Ctrl+"` | `toggle_terminal` |
| `Ctrl+'` | `toggle_terminal` |
| `` Ctrl+` `` | `toggle_terminal` |

### Git

| Key | Action |
|---|---|
| `f2` | `show_diff` |

### LSP

| Key | Action |
|---|---|
| `Ctrl+k` | `lsp_hover` |
| `Ctrl+Shift+i` | `lsp_format` |
| `Ctrl+space` | `lsp_complete` |
| `f4` | `lsp_goto_definition` |

### Navigation

| Key | Action |
|---|---|
| `Ctrl+Alt+i` | `jump_forward` |
| `Ctrl+Alt+o` | `jump_back` |

### Discovery

| Key | Action |
|---|---|
| `Alt+/` | `which_key` |
| `Alt+Shift+/` | `welcome` |
| `Ctrl+?` | `help` |
| `Ctrl+g` | `help` |
| `Ctrl+Shift+/` | `help` |
| `Ctrl+Shift+o` | `buffer_picker` |
| `Ctrl+Shift+p` | `command_palette` |
| `Ctrl+Shift+t` | `multi_picker` |
| `f1` | `help` |

### Other

| Key | Action |
|---|---|
| `Alt+z` | `toggle_soft_wrap` |
| `Ctrl+Alt+w` | `close_pane` |
| `Ctrl+q` | `quit` |
| `Ctrl+Shift+d` | `tree_new_dir` |
| `Ctrl+Shift+m` | `toggle_diagnostics` |
| `Ctrl+Shift+n` | `tree_new_file` |
| `delete` | `delete` |
| `esc` | `quit` |
| `f5` | `tree_refresh` |
| `pagedown` | `page_down` |
| `pageup` | `page_up` |

<!-- FIM: keymap -->

## Notes

- `Ctrl+Shift+<letter>` is ambiguous in terminals that do not speak *keyboard
  enhancements*: when the terminal cannot tell the difference, the key arrives as the
  variant without Shift. The editor resolves the ambiguity **from the fields of the key itself**, without
  guessing the terminal — see `docs/ui-ux/degradation.md`.
- A couple of keys use characters that need escaping in code: `Ctrl+"` and `Ctrl+\`.
  They were once lost by a generator that did not handle the escaped quote —
  that is why the table is produced from the Go API, and not by regex over the
  source.
- This document does **not** cover each surface's local keymap (arrow keys inside the
  tree, for example): that is the focus graph, in `docs/ui-ux/focus-graph.md`.
