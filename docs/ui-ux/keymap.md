# Mapa de teclas

**Fonte canônica.** Este arquivo é a única tabela de teclas do projeto. Os guias em
prosa — `docs/polish.md`, `docs/planning/ux-polish-plan.md`, os guias de uso —
apontam para cá em vez de repetir a tabela. Repetir foi como as divergências
apareceram: `:normal` em vez de `modal_mode`, `modal_editing` no lugar do id real,
e um `:help` que não existe. Todas corrigidas em M2, e todas da mesma classe.

As seções agrupam as teclas por **domínio**, derivado do prefixo do id por regra
determinística — um id novo entra no grupo certo pelo próprio nome, sem uma segunda
tabela para manter. Quem guarda os pares (tecla, id) é
`TestDocumentedKeymapMatchesTheBindings`, que compara os dois sentidos: uma tecla no
código e fora daqui, ou aqui e fora do código, falha listando exatamente a diferença.

**Remapear é configuração, não código.** Veja `docs/guides/pt/config.md`.

<!-- GERADO: keymap -->
### Arquivo

| Tecla | Ação |
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

### Edição

| Tecla | Ação |
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

### Movimento

| Tecla | Ação |
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

### Seleção

| Tecla | Ação |
|---|---|
| `Ctrl+a` | `select_all` |
| `Ctrl+Alt+down` | `add_cursor_below` |
| `Ctrl+Alt+u` | `clear_extra_cursors` |
| `Ctrl+Alt+up` | `add_cursor_above` |

### Busca

| Tecla | Ação |
|---|---|
| `Ctrl+f` | `find` |
| `Ctrl+h` | `replace` |
| `Ctrl+Shift+f` | `project_find` |
| `Ctrl+Shift+h` | `replace` |
| `f3` | `find_next` |
| `Shift+f3` | `find_prev` |

### Layout

| Tecla | Ação |
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

| Tecla | Ação |
|---|---|
| `Alt+-` | `terminal_shrink` |
| `Alt+=` | `terminal_grow` |
| `Ctrl+"` | `toggle_terminal` |
| `Ctrl+'` | `toggle_terminal` |
| `` Ctrl+` `` | `toggle_terminal` |

### Git

| Tecla | Ação |
|---|---|
| `f2` | `show_diff` |

### LSP

| Tecla | Ação |
|---|---|
| `Ctrl+k` | `lsp_hover` |
| `Ctrl+Shift+i` | `lsp_format` |
| `Ctrl+space` | `lsp_complete` |
| `f4` | `lsp_goto_definition` |

### Navegação

| Tecla | Ação |
|---|---|
| `Ctrl+Alt+i` | `jump_forward` |
| `Ctrl+Alt+o` | `jump_back` |

### Descoberta

| Tecla | Ação |
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

### Outros

| Tecla | Ação |
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

## Notas

- `Ctrl+Shift+<letra>` é ambíguo em terminais que não falam *keyboard
  enhancements*: quando o terminal não distingue, a tecla chega como a variante
  sem Shift. O editor resolve a ambiguidade **pelos campos da própria tecla**, sem
  adivinhar o terminal — ver `docs/ui-ux/degradation.md`.
- Um par de teclas usa caracteres que exigem escape em código: `Ctrl+"` e `Ctrl+\`.
  Elas já foram perdidas uma vez por um gerador que não tratava a aspa escapada —
  é por isso que a tabela é produzida a partir da API Go, e não por regex sobre o
  fonte.
- Este documento **não** cobre o keymap local de cada superfície (setas dentro da
  árvore, por exemplo): isso é o grafo de foco, em `docs/ui-ux/focus-graph.md`.
