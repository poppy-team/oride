# internal/tui/tree

O painel da árvore de projeto.

**Usa `component.Window`** para decidir a fatia visível: a seleção está sempre na
tela, e a janela não pula enquanto a seleção se move dentro do que já aparece.

**Os marcadores são texto, não glifos.** `v` e `>` dizem o que a entrada é e se
está aberta; uma seta seria caractere de largura ambígua — e a presença dela seria
o **único** sinal de que a linha pode ser expandida. Os ícones do `oride-fs`
entram quando houver tema, e o marcador de texto continua sendo o que carrega a
informação.

**`(vazio)`** distingue um painel sem entradas de um painel que falhou ao carregar.
Um retângulo em branco não distingue os dois.
