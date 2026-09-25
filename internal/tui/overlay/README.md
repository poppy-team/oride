# internal/tui/overlay

Tudo que flutua acima das superfícies: palette, which-key, ajuda e os modais.

**É uma casca fina sobre o componente `list`**, não uma lista própria. Uma versão
anterior escrevia à mão a lista, o filtro, a janela de scroll e a marcação da
seleção — seis comportamentos que a biblioteca de componentes já tinha, testados,
na versão que o plano de migração nomeava e que este projeto não adicionou. O
resultado foram buracos invisíveis: o build passava, os testes passavam, e o
programa não navegava.

**A regra que este pacote carrega é a de captura.** `Kind.Captures()` responde a
primeira regra de `docs/ui-ux/focus-graph.md`: com uma sobreposição ativa, nem a
superfície com foco nem o keymap global são consultados.

**`Filtering()` existe para o Escape ter dois donos sem ambiguidade.** Enquanto o
filtro está aberto, o Escape é dele; quando não está, é do painel. Fechar no
primeiro Escape descartaria um filtro meio digitado que a pessoa queria corrigir.

**A moldura é configurada aqui**, não pelo chamador: um chamador que esquecesse de
desligar a barra de status receberia texto do componente num painel de duas linhas.

**Título, itens e ciclo de vida são do chamador**; filtro, scroll e seleção são do
componente. Essa divisão é o motivo de o pacote existir.
