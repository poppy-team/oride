# internal/term

Terminal PTY embutido.

**Papel:** rodar um shell interativo no painel inferior.

**Não é um emulador de terminal** — e a referência também não é. Não há modelo de
tela nem parsing geral de escape: há um scrollback de texto e as sequências que
um prompt de fato emite (`\r`, `\n`, `\b`, `\t`, CSI `K J D C G H f`, OSC, SGR).
Implementar um emulador faria as duas implementações mostrarem coisas diferentes.

**`chunk.go` é puro:** `ApplyChunk(scrollback, cursorCol, texto)` não toca em
nada além dos dois ponteiros. É o miolo testado com as sequências exatas de zsh e
fish.

**`terminal.go` é a fronteira:** `creack/pty`, Unix. No Windows exige ConPTY —
falha com erro, não com panic.
