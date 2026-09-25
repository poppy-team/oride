# internal/tui/findbar

A barra de busca no buffer.

**A altura é calculada, não fixa.** Revelar o campo de substituição faz a barra
crescer, e o layout precisa saber disso **antes** de dividir a tela — por isso
`Height()` é método, e `layout.Wants` pergunta quantas linhas a barra quer.

**A barra fica acima do terminal e abaixo do editor**, tirando linhas do editor
em vez de cobrir o texto procurado. Uma barra que cobre o resultado que ela mesma
produz é uma barra que atrapalha.

**Os modificadores são escritos, não tingidos** (`[Aa acentos regex]`). Nenhum
estado é comunicado apenas por cor, e um modificador que não se vê é um
modificador sobre o qual se lê o resultado errado.

**O erro de expressão aparece no lugar da contagem.** Um padrão que não compila
não casou nada, então mostrar `0/0` ao lado do erro seria uma mentira útil apenas
para confundir.

**O cursor não é desenhado aqui.** O runtime é dono do cursor do terminal, e uma
barra que desenhasse o seu mostraria dois.

**A consulta longa mostra o fim, não o começo** — quem está digitando precisa ver
onde está digitando.
