# internal/keymap

Chords e resolução de tecla para ação.

**Papel:** traduzir uma tecla pressionada em uma `action.Action`.

**Decisão:** `Chord.Code` é um **token** canônico (`"s"`, `"enter"`, `"f1"`),
não um keycode. Assim `String()` e `Parse()` são inversos por construção, e o
round-trip não depende de uma tabela de tradução.

**Aliases:** `esc`/`escape`, `pgup`/`pageup`; e `cmd`/`super` colapsam em `Ctrl`,
`option` em `Alt` — paridade deliberada com o parser da referência.

**Uma camada só.** É um `map[Chord]Action` plano: sem prefixo, sem sequência,
sem camada modal. `ModalMode` existe na config e **não** tem máquina por trás —
quem implementar camadas acrescenta aqui, não configura.

`*Map` nil resolve nada, de propósito: um keymap ausente não deve digitar texto
por acidente.
