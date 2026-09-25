# internal/tui/layout

Geometria da tela, pura.

**Papel:** responder **onde** cada superfície cai, e **como** o texto cabe nela. O
modelo diz o que quer mostrar; este pacote diz onde aquilo aterrissa.

**Puro, e de propósito:** sem terminal, sem Bubble Tea, sem modelo. É o que
permite conferir as regras de `docs/ui-ux/layout.md` em qualquer largura sem
renderizar um frame — e é por isso que as classes de largura têm teste direto.

**Contrato de entrada:** `Wants` é um valor, não quatro parâmetros — três deles
booleanos, que é a aridade que o contrato de clean-code chama de cheiro.

**Regras que o teste guarda:**
- as fronteiras de classe são inclusivas embaixo (80 é `standard`, 120 é `wide`);
- nada é cortado em silêncio — o truncamento marca a perda com `…`;
- largura é medida em **células**, com caractere ambíguo contando como largo:
  errar para largo custa uma coluna, errar para estreito desalinha a linha;
- `Pad` devolve **exatamente** a largura pedida, porque uma linha curta deixa o
  que está à direita deslizar para a esquerda;
- zero é tamanho válido: um terminal ainda não medido produz regiões vazias, não
  um panic nem largura negativa;
- o editor cede por último — a árvore vira sobreposição antes de espremer o
  editor abaixo do mínimo.

**Dependência:** `mattn/go-runewidth`, já presente via lipgloss. É medida de
célula, não renderização.
