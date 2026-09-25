# internal/tui/menubar

A linha superior de rótulos de menu.

**O texto vem do catálogo de locale** — o `internal/i18n` existia com catálogos
completos e nenhum consumidor; esta é a primeira superfície que o usa. A barra não
tem string fixa.

**A ordem fica aqui, não no catálogo.** Uma tradução fornece as palavras; a
disposição é decisão da barra. Se a ordem morasse no arquivo de locale, um locale
poderia reordenar os menus.

**O menu aberto é marcado** (`[ ]`), não só tingido — mesma regra da statusbar.
