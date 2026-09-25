# Current Project State

- Project: **Oride Terminal IDE**
- Framework: **Prumo v0.6.0** (Protocolo v3)
- Current phase: **M0 — Fundação, Oráculo e Prumo**
- Current goal: **Migração Rust → Go com paridade diferencial provada**
- Context methodology: **Lean Progressive Context (LPC)**
- Last updated: `2026-09-25T00:00:00Z`

## M3 — Shell TUI (em andamento)

### M3.0 — Fundamentos verificáveis · **completo**

- ✅ **`internal/architecture`** — o contrato de arquitetura como teste. Sete
  regras (R1–R7) extraídas de `docs/architecture/clean-code-contract.md` §3:
  ausência de ciclos, não importar o consumidor, superfícies de `tui` não
  importarem o modelo, folhas puras permanecerem folhas, nomes de pasta
  proibidos, README por pasta, e o harness não virar dependência de produto.
  Cada regra é dado (`Rule`), então acrescentar uma não significa escrever outro
  teste. **Contraprova:** `TestRulesCanFail` prova que cada regra sabe falhar e
  que sabe silenciar — uma regra que nunca falha não mede nada.
  O R6 obrigou **20 READMEs** que o contrato exigia e o repositório não tinha.
- ✅ **`internal/docdrift`** — contratos de documentação contra o código.
  `docs/ui-ux/keymap.md` é **gerado** das 98 bindings e comparado nos dois
  sentidos: tecla no código e fora do doc, ou no doc e fora do código, falha
  nomeando a diferença.
- ✅ **`docs/ui-ux/`** — o contrato que o gate do M3 exige e que não existia:
  `keymap.md` (canônico, gerado), `focus-graph.md`, `states.md`, `layout.md`
  (classes de largura), `degradation.md` (cor, mouse, keyboard enhancements) e a
  projeção EN apontando para a fonte canônica em vez de duplicá-la.
- ✅ **`docs/architecture/overview.md`** (PT + EN) corrigido: descrevia `Ropey` e
  `Ratatui`/`Crossterm`, nenhum dos três usado pelo produto em Go.

### M3.1 — Decomposição do modelo · próximo

`App` (63 campos planos) vira composição de sub-estados coesos; o `switch` de
`Apply` vira registry por domínio; `StateDump` e `SchemaVersion` **não mudam**.

## Onde o projeto está

A implementação alvo é **Go** (`cmd/`, `internal/`). A implementação **Rust**
(`crates/`) está congelada como **oráculo diferencial**: ela responde o que o
produto faz hoje, e o porte é comparado contra ela sem tolerância. Quando a
paridade for atingida, `crates/` é arquivado em tag.

Paridade **não** é uma afirmação — é um número que o harness produz, em
`conformance/`, e cada divergência intencional está em
[docs/migration/parity-ledger.md](docs/migration/parity-ledger.md).

## Concluído

### M0 — Fundação, Oráculo e Prumo

- ✅ **Oráculo scriptável.** `oride conformance run --case <FILE> --workspace <DIR> --json`
  dá ao Rust um modo de entrada headless determinístico. Antes disso `--headless`
  era um stub que só emitia erro — sem ele, "paridade 100%" não passava de intenção.
- ✅ **Harness de conformance em Go** com duas garantias provadas por teste: o
  oráculo é determinístico (mesmo caso, dois diretórios, relatórios idênticos) e o
  comparador detecta divergência (um comparador que sempre diz "igual" faria a
  paridade subir a 100% sozinha).
- ✅ **Ledger de paridade** com D1–D11 (dissonância doc↔código) e B1–B18 (bugs)
  triados por veredito: *spec vence*, *código vence*, *removido* ou *melhoria*.
- ✅ **Correções aplicadas e guardadas por teste:** `mouse` desligado por default;
  `[editor]` duplicado no exemplo de config; Enter fechando o grupo de undo;
  dependência morta removida; digest de sessão estável (FNV-1a com valor golden
  que o Go precisa reproduzir).
- ✅ **Fundação Go:** módulo, `cmd/oride`, `internal/buildinfo`, `conformance/`,
  CI de três jobs (`rust`, `go`, `conformance`), `CGO_ENABLED=0` verificado a cada
  mudança.
- ✅ **Governança Prumo instalada:** `prumo.json` (Protocolo v3), `docs/PRUMO.md`
  como router, `.prumo/history/`, `project-profile.json`, workforce resolvido
  deterministicamente (14 agents, 58 skills do catálogo + 8 locais, 6 recipes),
  roles de agente geradas dos contratos canônicos, adaptador Antigravity
  recompilado. `prumo validate` e `prumo doctor` passam limpos.
- ✅ **Skills locais** para o que o catálogo não cobre: `bubbletea-v2`,
  `tui-editor-engine`, `conformance-parity`, `doc-drift`, `pty-vt`, `lsp-client`,
  `cross-platform-terminal`, `release-go`. Tudo regerável por
  `scripts/workforce.sh`.
- ✅ **Dissonância documental corrigida:** dos 12 itens fechados no ledger, oito
  eram doc↔código — incluindo uma chave de config que não existe, um comando
  nunca implementado e um default documentado ao contrário do código.

### M1 — Domínio puro (fechado no essencial)

- ✅ **`internal/action`** — 96 ids canônicos gerados do Rust, com guard de drift
  no CI (verificado injetando um typo).
- ✅ **`internal/keymap`** — `Chord` com o código sendo o próprio token canônico,
  o que faz `Parse`/`String` serem inversos por construção. 98 bindings.
- ✅ **`internal/buffer`** — a semântica de índice que o `ropey` dava de graça,
  verificada contra o oráculo em bytes, scalars, linhas, caret e conversão inversa.
- ✅ **`internal/editor`** — seleção, undo com rótulos exatos, documento,
  movimentos, multi-cursor, store. Corrige **B2, B3 e B4** do ledger.
- ✅ **`internal/config`** — modelo, TOML em camadas (padrão ← usuário ← projeto),
  todos os clamps, registry de temas (10 temas, 252 cores, geradas do Rust por
  `scripts/gen-tables.py`, com temas de disco) e resolução de `.editorconfig`.
- ✅ **`internal/i18n`** — catálogos TOML embutidos e de disco, com herança do
  catálogo padrão para traduções parciais.
- ✅ **`internal/fs`** — árvore de projeto: expansão preguiçosa, ordenação,
  filtros, comparada contra o oráculo.
- ✅ **`internal/git`** — status por porcelain `-z`.
- ✅ **`internal/search`** — matcher do buffer: dobra de caso e acento, fronteira
  de palavra, regex, offsets em bytes.

**M1 está completo.** Todo módulo do domínio puro tem testes próprios, e os que
são observáveis no dump são comparados byte a byte contra o oráculo.

### M1-e — O harness diferencial mede de verdade

- ✅ **`internal/app`** — driver headless que executa um caso e produz o mesmo
  dump do oráculo. `TestGoMatchesTheOracle` compara os dois.
- ✅ **Estado comparado:** keymap completo, config efetiva, texto, contagens,
  caret, seleção, cursores extras, rótulos de undo/redo, versão, abas e **árvore**.
- ⏳ **Ainda não produzido** (com o motivo de cada chave, em `NotYetImplemented`):
  `scm`, `find`, `split`, `vim`, `diagnostics`, `lsp_failures`, `status`.
- **5 divergências declaradas**, todas rastreando para B2/B3/B4 — os itens em que
  o ledger diz *spec vence*. O caso `multi-cursor` documenta o B4 ao vivo: o
  oráculo mantém cursores além do fim do buffer e o backspace seguinte não altera
  o texto.

### M2 — Adaptadores (em andamento)

- ✅ **`internal/session`** — persistência com digest FNV-1a 64. Os três valores
  golden do Rust são reproduzidos byte a byte, o que é paridade cross-language
  sem depender do oráculo: um digest divergente faria a sessão de uma
  implementação ser invisível para a outra.
- ✅ **`internal/plugin`** — manifestos `plugin.toml`, descoberta nos dois
  layouts, expansão de `$FILE`/`$DIR`/`$WORKSPACE`, e execução que **não** passa
  por shell: um nome de arquivo com ponto e vírgula é um nome de arquivo, não um
  segundo comando. Um manifesto quebrado é ignorado, não fatal.
- ✅ **`internal/lsp`** — cliente JSON-RPC sobre stdio com framing
  Content-Length. O transporte é um par de streams, **separado do processo**, o
  que deixa o cliente inteiro ser testado por um pipe sem servidor instalado. A
  conversão escalar↔UTF-16 recusa uma posição que parte um par substituto em vez
  de arredondar para o lado errado.
- ✅ **`internal/term`** — terminal PTY embutido (`creack/pty`). O miolo é
  `ApplyChunk`, função pura sobre o scrollback e a coluna do cursor, testada com
  as sequências exatas que zsh e fish emitem. **Não é um emulador de terminal** —
  a referência também não é, e inventar um faria as duas implementações
  mostrarem coisas diferentes.
- ✅ **Busca no projeto** — `rg --json` e fallback em Go.

### 0.2.0 — entregue antes da migração

- i18n dinâmico via catálogos TOML; tutorial de temas; motor modal Vim; task
  runner com substituição de variáveis; diagnóstico `:health`; Markdown rico;
  sync Git; busca e substituição no projeto.

## Próxima ação

- **M1, na ordem de dependência:** `keymap`/`chord` (com o dump de keymap do
  oráculo como verificação), depois `buffer` e `editor` — que é onde vivem os
  itens B1–B4 e B10–B11 do ledger —, depois `config`/`theme`/`i18n`.
- **Depois M2:** adaptadores de I/O (git, fs, search, lsp, pty/term, session,
  plugin). **M3:** shell TUI em Bubble Tea v2 com golden frames. **M4:** gate de
  paridade — 100% dos casos, ledger sem exceções abertas.

## Ordem de recuperação

1. `AGENTS.md` (regras e invariantes do produto).
2. `prumo.json` (manifesto canônico) e `PROJECT_STATE.md` (este arquivo).
3. `docs/PRUMO.md` (router de documentação) — leia só o que a tarefa pede.
4. `docs/migration/parity-ledger.md` antes de mexer em qualquer comportamento.
5. Código: `internal/` (alvo) e `crates/` (oráculo).

Não carregue o repositório inteiro por padrão.
