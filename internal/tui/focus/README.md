# internal/tui/focus

O grafo de foco: onde a entrada pode aterrissar, em que ordem, e o que tem
precedência.

**É dado, não cascata de `if`.** A referência resolvia foco pela ordem de condições
numa função de dois mil linhas: a ordem existia, mas nada podia afirmá-la, então um
`if` inserido no topo roubava uma tecla de um modal sem ninguém notar. Aqui a ordem
é um valor, e as regras de `docs/ui-ux/focus-graph.md` são testáveis direto.

**Não conhece `internal/app`.** O composition root mapeia uma `Surface` para o
vocabulário de foco do modelo — é o que mantém este pacote reutilizável e uma
superfície descartável (e o que a regra R3 de `internal/architecture` exige).

**Regras guardadas por teste:**
- `Tab` percorre as superfícies **visíveis** na ordem declarada e dá a volta;
- superfície oculta é **pulada**, porque o foco nunca pode parar onde nada é
  desenhado — a próxima tecla iria para um lugar invisível;
- com só o editor visível, o foco **fica parado**: `Tab` não pode ser um jeito de
  perder o cursor;
- uma superfície que saiu do grafo cai no editor, em vez de deixar o foco solto;
- toda superfície declarada é **alcançável** — uma que o `Tab` nunca encontra é uma
  que ninguém consegue focar;
- `Order` devolve cópia, para um chamador não reordenar o grafo por fora.

**Captura por sobreposição** (regra 1 do documento) é do modelo, que sabe quando há
uma sobreposição ativa: enquanto houver, nem o nó com foco nem o keymap global são
consultados.
