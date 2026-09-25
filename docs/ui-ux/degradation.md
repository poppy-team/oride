# Degradação de capacidade

Cada capacidade é **declarada no view** ou **ausente com elegância**. Nunca
alternada imperativamente a cada frame — é assim que um terminal fica quebrado
depois de um encerramento abrupto.

## Cor

| Capacidade | O que acontece |
|---|---|
| truecolor | cores exatas do tema |
| 256 | cada cor é reduzida ao índice mais próximo |
| 16 | reduzida à paleta do terminal |
| nenhuma | **todo token não primitivo resolve para nada** |

Um tema `no-color` que pinta qualquer coisa é um erro, não uma escolha. E
**nenhum estado é comunicado só por cor**: o que está tingido também está escrito.
Cor é reforço; nunca é a informação.

## Mouse

Desligado por padrão. Ligado, é estado do modelo e vira `MouseMode` no view — o
runtime cuida de capturar e soltar. Um editor que captura o mouse sem o usuário
pedir rouba a seleção de texto do terminal.

## Keyboard enhancements

Pedidos **apenas quando o programa usa**, e o terminal que não suporta é tratado
como caso normal, não como erro.

Consequência concreta: `Ctrl+Shift+<letra>` é ambíguo. Sem *enhancements*, o
terminal entrega a tecla sem distinguir o Shift. A ambiguidade é resolvida **pelos
campos da própria tecla** — nunca por uma suposição sobre qual terminal está do
outro lado.

## Barra de foco

Reportada como estado do view. Um terminal que a suporta não precisa que o editor
adivinhe quando perdeu o foco.
