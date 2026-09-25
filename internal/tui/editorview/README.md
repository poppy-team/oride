# internal/tui/editorview

O viewport do texto.

**As linhas chegam já extraídas**, não como buffer: a superfície não alcança o
documento, e um viewport não precisa do arquivo inteiro para desenhar uma tela.

**O caret é marcado na gutter** (`>`), não pintado sobre um caractere. Um cursor de
terminal é coisa que o runtime desenha; uma superfície que desenhasse o seu também
mostraria dois. E o marcador é texto, então sobrevive ao terminal sem cor — onde um
caret tingido seria invisível e a posição, desconhecida.

**Numeração de linha com largura derivada**, não fixa: num arquivo de 10 mil linhas
uma gutter fixa desalinharia todo o texto.

**A gutter cede primeiro.** Se ela não deixar largura útil para o texto, ela sai —
uma coluna de números sem texto ao lado não é um editor menor, é um editor quebrado.

**A seleção ainda não é pintada aqui.** Ela é *dado* no modelo, e o tratamento
visual chega com o tema; pintar texto sem cor para marcar seleção corromperia o
próprio texto que a superfície existe para mostrar.

**Toda linha volta com largura exata**, e o viewport volta com a altura pedida —
uma tela curta deixaria o que está abaixo subir.
