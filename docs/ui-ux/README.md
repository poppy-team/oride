# Contrato de UI/UX

O comportamento visível do editor, descrito de forma **verificável**.

Cada documento aqui tem uma contraparte no código, e um teste que compara os
dois. Um contrato que ninguém checa é um contrato que decai — e este projeto já
pagou esse preço: `:normal` no lugar de `modal_mode`, `modal_editing` no lugar do
id real, um `:help` que não existia.

| Documento | Descreve | Verificado por |
|---|---|---|
| `keymap.md` | a tabela canônica de teclas | `TestDocumentedKeymapMatchesTheBindings` |
| `focus-graph.md` | o grafo de foco e a captura por sobreposição | teste de foco (M3.2) |
| `states.md` | o vocabulário de estados observáveis | `TestStateVocabularyMatchesTheDump` (M3.2) |
| `layout.md` | as classes de largura e o que muda em cada uma | golden frames (M3.2) |
| `degradation.md` | cor, mouse e keyboard enhancements | golden frames `no-color` (M3.2) |

## Precedência

1. O código e os testes — o que o produto faz.
2. Este diretório — o que o produto **promete** fazer.
3. Guias e planejamento — prosa que aponta para cá.

Quando 1 e 2 divergem, um dos dois está errado e o teste falha. Corrigir a
promessa sem corrigir o código (ou o contrário) é o defeito, não a correção.

## Regra de manutenção

Nenhum documento deste diretório repete uma tabela que exista em código. A tabela
de teclas é **gerada**; regenerar é `go test ./internal/docdrift/ -update`.

A projeção em inglês fica em `docs/en/ui-ux/` e **aponta** para a tabela canônica
em vez de mantê-la em duplicata: duas cópias divergem, e uma tabela de teclas
divergente é pior que nenhuma.
