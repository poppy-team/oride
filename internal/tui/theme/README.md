# internal/tui/theme

Traduz as cores da configuração em estilos.

**É o único lugar que sabe que cor existe.** As superfícies pedem um **papel** —
seleção, gutter, status — e recebem um estilo; trocar a paleta não alcança o
código que desenha.

**O perfil é parâmetro, não ambiente.** Um teste precisa pedir `NoColor` sem
escrever no ambiente do processo, que é estado compartilhado — a mesma regra que
o `bubbletea-v2` enuncia como *capability is injected by option, never by writing
to the process environment*.

**`NoColor` devolve o estilo zero**, que não aplica atributo nenhum. Esse é o
invariante inteiro do modo sem cor, e ele mora num lugar só: um tema que afirma
suprimir cor e pinta uma é um erro, não uma escolha.

**`""` e `"reset"` significam "sem cor"**, e ambos são normais nos temas que o
produto entrega. Filtrar aqui evita entregar à biblioteca de estilo um valor que
ela não resolve.

**Uma função de papel por enquanto**, não vinte: o mapeamento de token do chroma
para papel chega com o motor de realce, e declarar dezenove acessores sem uso
seria abstração que não paga.

**Depende de `layout.Width` medir texto com ANSI**, o que é o que permite uma
linha estilizada continuar ocupando exatamente as células que ocupa sem estilo.
