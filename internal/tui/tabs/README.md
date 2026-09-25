# internal/tui/tabs

A barra de abas.

**A aba ativa é desenhada primeiro quando nem todas cabem.** Uma barra que esconde
qual documento está sendo editado é pior do que uma que esconde os outros — o
corte é marcado com `…`, porque uma barra que descarta abas em silêncio parece uma
barra com menos abas.

**A aba ativa e o estado sujo são escritos, não só tingidos:** `[ ]` em volta da
ativa e `*` no título sujo. Nenhum estado é comunicado apenas por cor — vale para
quem não distingue o tom e para o terminal sem cor.

**A largura da aba é fixa** (um espaço de cada lado), então a barra não se mexe
quando a seleção muda.
