# internal/buffer

O modelo de texto.

**Papel:** bytes, indexação por linha e conversão entre as quatro noções de
posição. É a autoridade única sobre elas.

| Noção | Significado |
|---|---|
| Offset de byte | índice canônico interno |
| Offset de runa | o que "coluna" significa para um caret |
| Grafema | o que o usuário percebe como um caractere |
| Coluna visual | tabulação e largura dupla |

**Fronteira dura:** grafemas e colunas visuais pertencem à **camada de view**.
Este pacote não as calcula, e nenhum outro pacote as calcula fora daqui — quem
converter por conta própria reintroduz a classe de bug que o pacote existe para
impedir (ver a skill `tui-editor-engine`).

**Puro.** Não importa nenhum pacote interno (R4). Testes de propriedade cobrem
marcas combinantes, emoji com modificadores, CJK, RTL, BOM e CRLF.
