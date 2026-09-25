# internal/tui/statusbar

A linha inferior.

**Puro:** recebe valores, não uma referência ao modelo — o que permite testá-la em
qualquer largura sem terminal.

**A posição do cursor fica na borda direita e ganha o espaço que precisa antes de
todo o resto.** É o campo que se procura num lugar fixo; um valor que se move é um
valor que se perde. À esquerda vão o arquivo, o foco e a mensagem transitória, nessa
ordem de prioridade para truncamento.

**O flag de sujo é escrito, não só tingido** (`*`). Nenhum estado é comunicado
apenas por cor — vale para quem não distingue o tom e para o terminal sem cor.
