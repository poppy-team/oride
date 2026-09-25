# Layout e classes de largura

## Classes

| Classe | Largura | O que muda |
|---|---|---|
| `compact` | < 80 | A árvore entra sobreposta em vez de lado a lado; a statusline encurta rótulos; o painel de terminal ocupa a largura toda |
| `standard` | 80–119 | Layout completo: árvore, editor e statusline lado a lado |
| `wide` | ≥ 120 | Como `standard`, com largura extra para a árvore e para a coluna de diagnóstico |

Os limites são **inclusivos embaixo**: 80 é `standard`, 120 é `wide`.

## Regras

1. **A largura é decidida uma vez por frame**, a partir do tamanho que o runtime
   entrega. Nenhuma superfície recalcula a classe por conta própria.
2. **Nada é cortado em silêncio.** O que não cabe é truncado com marcador
   visível; um rótulo que some sem sinal é indistinguível de um rótulo que nunca
   existiu.
3. **Largura é conservadora.** Caractere de largura ambígua é tratado como largo.
   Errar para largo custa uma coluna; errar para estreito desalinha a linha
   inteira.
4. **Zero é um tamanho válido.** Um terminal ainda não medido não pode fazer o
   editor entrar em pânico nem dividir por zero.
