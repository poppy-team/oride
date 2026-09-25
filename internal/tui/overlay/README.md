# internal/tui/overlay

Tudo que flutua acima das superfícies: palette, which-key, ajuda e os modais.

**Um pacote, não quatro.** As quatro compartilham a mesma moldura — uma lista
filtrada com título e dica, ou uma caixa centrada. Separar significaria ou três
cópias do renderizador de lista, ou um quarto pacote guardando ele; o contrato
pede coesão antes de pedir simetria.

**A regra que este pacote carrega é a de captura.** `Kind.Captures()` responde a
primeira regra de `docs/ui-ux/focus-graph.md`: com uma sobreposição ativa, nem a
superfície com foco nem o keymap global são consultados. Sem isso, uma tecla
digitada no filtro da palette também editaria o buffer atrás dela.

**Duas funções, não quatro wrappers.** `List` e `Box` são os renderizadores; o
modelo escolhe título, itens e dica. Funções `Palette`, `WhichKey` e `Help` que só
repassassem argumentos seriam decoração sem uso, e o §1.6 do contrato proíbe
abstração que não paga.

**A dica fica sempre na última linha**, qualquer que seja o conteúdo: uma dica que
se move é uma dica que ninguém acha duas vezes.

**Vazio tem nome.** Uma palette sem resultados diz isso, em vez de mostrar uma
caixa em branco — que é indistinguível de uma que falhou.

**A caixa deixa margem**, para a superfície de trás continuar visível: uma
sobreposição que cobre tudo não é uma sobreposição, é outra tela.
