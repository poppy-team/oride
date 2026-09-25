# internal/editor

Documentos, seleção, undo e multi-cursor.

**Papel:** o modelo de edição sobre `internal/buffer`.

**Edição transacional:** toda alteração aplica e registra o inverso no mesmo
passo. É o que torna undo, redo e replay possíveis sem casos especiais.

**Grupos de undo fecham em fronteira explícita, nunca por heurística de tempo:**
newline, movimento, save, troca de foco e chamada explícita. A referência Rust
não fechava no newline — `Ctrl+Z` depois de `abc` + Enter + `xyz` apagava os
três (ledger B1).

**Undo e redo restauram a seleção que o grupo produziu**, e descartam cursores
que ficaram além do fim (B2–B4).

**Invariantes de multi-cursor:** cursores ordenados e sem duplicata; edições
aplicam do maior offset para o menor; toda operação que encolhe o buffer
revalida cada cursor.

**Só abre arquivo; não salva.** Não existe caminho de escrita neste pacote — o
dirty flag é limpo por `MarkSaved`, e quem grava é a camada que tem o disco.
