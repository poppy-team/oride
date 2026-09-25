> **Plano ativo de refatoração e dependências:** [`refactoring-plan.md`](./refactoring-plan.md).
> Ele registra o saneamento pendente, as bibliotecas a acoplar e os próximos passos
> de M4, M5 e M6 — inclusive o motivo de o gate ter passado com a TUI inutilizável.

# Ledger de Paridade — Oride Rust → Go

Registro de toda diferença entre o **oráculo Rust** (`crates/`) e a
**implementação Go**, e de toda contradição entre o que o produto faz e o que a
documentação afirma.

**Regra:** o ledger é a única fonte de exceções. Um caso de conformance que
divirja do oráculo sem entrada aqui é uma regressão, não uma decisão. Nenhuma
entrada pode ficar aberta no gate do M4.

**Triagem:**

| Veredito | Significado |
|---|---|
| **spec vence** | O código Rust está errado. O Go implementa a spec corrigida, e o caso de conformance registra a divergência esperada |
| **código vence** | A documentação está errada. Corrige-se o documento; o comportamento é paridade |
| **removido** | A funcionalidade não sobrevive à migração. Remoção documentada e aprovada |
| **melhoria** | O Go entrega comportamento superior. Não é paridade e não se disfarça de paridade |

---


> **Achado de escopo (M2, terminal).** O `oride-terminal` da referência tem 456 linhas e **não é um emulador de terminal**: não há modelo de tela nem parsing de sequências de escape alem das poucas que um prompt emite. A estimativa inicial desta migração tratava o terminal como o módulo mais caro de M2, equivalente a um marco inteiro; lido o código, são PTY mais buffer de linhas. A paridade aqui é com essa lista curta de sequências — `\r`, `\n`, `\b`, `\t`, CSI `K J D C G H f`, OSC e SGR — e o Go reproduz cada uma, inclusive as peculiaridades (ver B26 e a nota sobre `K=1`), em vez de implementar um emulador que mostraria algo diferente.

## Nota medida — o `expect` do snippet de LSP (item do B11)

O registro do B11 listava `oride-lsp/src/client.rs:533` (`expect("cursor within
snippet")`) como panic em caminho de usuário. **Medido, não suposto:** o
`expect` é inalcançável hoje, e o panic vizinho — `&snippet[cursor..]` sobre um
índice que não seja fronteira de caractere — também não dispara. Todas as rotas
do laço avançam `cursor` até uma fronteira ASCII ou um caractere inteiro.

Entrada adversarial em teste (`malformed_snippets_do_not_panic`): `""`, `"$"`,
`"${"`, `"${}"`, `"${:}"`, `"${a:"`, `"${a:${b}"`, `"$1"`, `"$99999999999999999999"`,
`"まで${1:日本語}"`, `"${1:😀"`, `"$😀"`, `"${1:${2:x}}"` — nenhuma derruba.

O que resta é fragilidade, não bug: a invariante depende da aritmética de três
ramos, e uma edição futura a quebra silenciosamente em panic. O teste fica como
rede. **Não é bug vivo** — e por isso o item muda de natureza em vez de virar
"corrigido".

 — onde rodar `prumo validate` / `doctor`

Confirmado de novo neste marco, e vale para quem for repetir: **rode os dois de
dentro do clone do Prumo** (`~/Documentos/Projetos/prumo`), não do repositório
Oride. Rodando do repo errado, o binário não encontra o registro de schemas e o
resultado são falsos positivos com cara de problema real:

| CWD | `validate` | `doctor` |
|-----|-----------|----------|
| dentro do clone Prumo | passa | limpo |
| dentro do repo Oride | falha: `schema registry: open schemas: no such file or directory` | 6 avisos `Selected recipe '...' is not registered` |

O sinal de que é artefato e não drift: ele acusa **todas** as seis receitas
selecionadas como não registradas, inclusive as que existem em
`.ai/recipes/manifest.json`. Um registro ilegível acusa tudo; drift real acusaria
alguma. O `prumo doctor` de dentro do clone passa limpo.

Os artefatos do projeto estão íntegros — verificado dos dois lados.

1. Dissonâncias entre documentação e código

| # | Dissonância | Evidência | Veredito | Fase | Estado |
|---|---|---|---|---|---|
| D1 | **`mouse` default divergente.** Docs diziam desligado, código dizia ligado | `crates/oride-config/src/model.rs:45` (`mouse: true`) vs `README.md:21`, `README.pt-BR.md:21`, `assets/config.example.toml:10`, `CHANGELOG.md:51`, `crates/oride-app/src/app/mouse_handler.rs:16`, `docs/planning/roadmap.md:30` | **spec vence** — `false`. Alinha com a visual-constitution: teclado completo, mouse é aceleração opcional | M0-e | **corrigido** |
| D2 | `docs/plugin-api.md` afirma que o Oride **não** carrega plugins externos, mas `plugin/external.rs` implementa plugins por manifesto | doc: *"não carrega plugins externos (Lua/WASM/dynload) — anti-bloat"* vs `crates/oride-plugin/src/{external,manifest}.rs` | **spec vence** — o doc é reescrito. Manifestos declarativos com isolamento de processo (Prumo ch17: *instalar ≠ conceder acesso*) | M2 | corrigido |
| D3 | CHANGELOG 0.2.0 alegava grammars tree-sitter oficiais para Python, TypeScript, TSX e Ruby | Só existem Rust, C, Bash, Markdown e OriScript (`crates/oride-syntax/src/highlight.rs::language_ts`); o resto é o highlighter lexical | **código vence** — corrigido, declarando explicitamente o que é lexical e o que é grammar dinâmica | M0-e2 | **corrigido** |
| D4 | `docs/guides/pt/markdown.md` marca links clicáveis (M1) e imagens (M2) como "planejado" | `CHANGELOG.md` e `ROADMAP.md` marcam ambos como entregues; `crates/oride-syntax/src/md_preview.rs` implementa | **código vence** — corrigir o guia | M2 | corrigido |
| D5 | `docs/en/README.md` descreve `plugin-api.md` como protocolos stdin/stdout JSON | contradiz o documento real | **código vence** — corrigir o índice | M2 | corrigido |
| D6 | `docs/design.md` referencia crates e arquivos que não existem | `oride-theme`, `plugins/`, `docs/keymaps.md`, `docs/config.md`, `tokio`, `vte`, `nucleo`, `icons.toml`, `git2`. O PT também omite `oride-i18n`, que o EN inclui | **código vence** — reescrever como *as-built* | M2 | corrigido |
| D7 | Rótulo da fase P4 conflita | `docs/design.md:338` "Polimento 0.1" vs `AGENTS.md` e `PROJECT_STATE.md` "Polimento 0.2" | **código vence** — unificado em "Polimento 0.2" | M0-e2 | **corrigido** |
| D8 | ROADMAP afirmava **undo/redo ramificado** | `docs/planning/roadmap.md:23`, `docs/en/planning/roadmap.md:23`. O undo é uma pilha linear (`crates/oride-core/src/undo.rs`: `undo`, `redo`, `open`). Nota: splits e multi-cursor atribuídos a v0.1.0 **estão certos** — chegaram em `0.1.0-alpha.6`, que é parte da linha v0.1.0. Minha triagem inicial chamou isso de erro; era exagero meu | **código vence** — corrigido nos dois roadmaps | M0-e2 | **corrigido** |
| D9 | Variáveis do task runner divergiam entre roadmap e guias | Código (`crates/oride-app/src/tasks.rs:41-77`) implementa `$FILE`, `$FILE_NAME`, `$FILE_STEM`, `$FILE_DIR`, `$WORKSPACE`, `$LINE`, `$COL`. Os guias documentavam `${file}`, `${dir}`, `${workspace_root}` — **nenhum existe**. O roadmap listava `$DIR`, também inexistente | **código vence** — guias PT/EN e roadmap corrigidos; o teste `resolve_magic_variables` já pina a convenção real | M0-e2 | **corrigido** |
| D10 | `assets/config.example.toml` declarava `[editor]` duas vezes — TOML inválido | linhas ~14 e ~90 do arquivo | **código vence** — corrigido, e agora **guardado por teste**: `crates/oride-app/tests/config_example.rs` faz o exemplo falhar o build em tabela duplicada, id de ação inexistente, chord inválido ou default documentado que divergiu do código (verificado: com a duplicata reintroduzida, os 5 testes falham) | M0-e | **corrigido** |
| D11 | READMEs e guias mandavam digitar `:normal` para ativar o modo modal | `execute_vim_command` (`crates/oride-app/src/app/input.rs:778-846`) implementa `w`/`write`, `q`/`quit`, `q!`, `wq`/`x`, `qa`/`qall`/`quitall`, `health`/`checkhealth`, `tasks`, `run`, `e`/`edit`, `theme`, `lang`, `noh`/`nohlsearch` — **não há `:normal`** | **código vence** — READMEs e guias corrigidos para a ativação real (`modal_mode = true` ou o toggle na palette) | M0-e2 | **corrigido** |
| D12 | **Três erros adicionais encontrados ao corrigir o D11** | (a) os guias prescreviam `modal_editing = true`; a chave é `modal_mode` (`crates/oride-config/src/model.rs:70`); (b) prescreviam `language = "pt-BR"`; a chave é `locale`; (c) `:help` era documentado mas **não** existe — a tela de atalhos é `F1`/`Ctrl+G`/`Ctrl+Shift+/`. E `:lang` e `:noh`, que existem, não estavam documentados | **código vence** — guias PT/EN corrigidos; `:lang` e `:noh` passaram a constar | M0-e2 | **corrigido** |

---

## 2. Bugs de código

| # | Bug | Evidência | Veredito | Fase | Estado |
|---|---|---|---|---|---|
| B1 | **Enter não fechava o grupo de undo.** Digitar, pressionar Enter e digitar coalescia num único grupo: um só `Ctrl+Z` apagava as duas digitações e o newline. Movimento **é** fronteira (funciona), newline não era | `grep commit_group crates/`: `move_head_to` (`crates/oride-core/src/document.rs:215-221`) e `set_carets_from_offsets` (`:205-213`) fechavam o grupo; `insert_text` e `insert_newline_smart` não. Caso `editor/typing-undo-groups.json` antes do fix: passo 4 com 27 bytes, undo do passo 5 voltava a 20 — 7 bytes num grupo só | **spec vence** — corrigido no oráculo via `Document::commit_newline_boundary`: o Go reproduz byte a byte, sem exceção Tier A | M0-e | **corrigido** |
| B2 | **Redo não restaura o caret.** Após redo o caret fica na posição clampada do undo, não onde a edição aconteceu | `crates/oride-core/src/document.rs:643-652`: `redo` faz `head.min(len)` e recolapsa a seleção. Caso `editor/typing-undo-groups.json`: passo 4 caret `(3,3)` → passo 7 (redo) caret `(2,0)` | **spec vence** — corrigido no Go: cada grupo de undo carrega a seleção com que começou e terminou (`EditGroup.before`/`after`), então undo volta ao estado anterior ao passo e redo ao posterior. Pinado por `TestRedoRestoresTheCaret` | M1 | **corrigido** |
| B3 | Undo e redo **recolapsam uma seleção ativa** num caret | `crates/oride-core/src/document.rs:621-652`: ambos fazem `Selection::caret(...)` | **spec vence** — resolvido pela mesma mudança do B2: a seleção restaurada é a gravada no grupo, não um caret derivado do head clampado. Uma seleção ativa sobrevive ao ciclo undo/redo | M1 | **corrigido** |
| B4 | **Cursores extras não são clampados após undo/redo.** `extra_carets` pode apontar além do fim do buffer | `crates/oride-core/src/document.rs:621-652` clampa só `selection.head`; `undo.rs` mantém `extra_carets` intactos. Próximo insert falha com erro de buffer | **spec vence** — corrigido no Go: `restoreSelection` descarta cursores além do novo fim e o teste termina com um insert que precisa funcionar. Pinado por `TestUndoDropsCursorsPastTheEndOfTheBuffer` | M1 | **corrigido** |
| B5 | Terminal emulado ignora SGR, não faz wrap, não tem alt-screen nem histórico | `crates/oride-terminal/src/lib.rs::apply_terminal_chunk` (subset CSI escrito à mão) | **melhoria** — `charmbracelet/x/vt` entrega os quatro. Não é paridade e não se apresenta como tal | M2 | aberto |
| B6 | Sem bracketed paste; `Event::Paste` é descartado | `crates/oride-app/src/run.rs:76` cai em `_ => {}` | **melhoria** — `PasteMsg` nativo do Bubble Tea v2 | M3 | aberto |
| B7 | `--headless` sem `--stat` é um stub que só emite erro | `crates/oride/src/main.rs` (antes de M0-a) | **resolvido** — substituído pelo modo conformance, que é o headless real | M0-a | **corrigido** |
| B8 | `project_replace` existe como ação mas não tem binding default | `crates/oride-config/src/model.rs::default_key_bindings` não o inclui; alcançável só pela palette | decidir binding e documentar | M4 | aberto |
| B9 | `strip-ansi-escapes` declarado e nunca usado | `crates/oride-terminal/Cargo.toml:12` | **código vence** — removido do crate e do workspace | M0-e | **corrigido** |
| B10 | Vazamento ilimitado por `Box::leak` | `crates/oride-syntax/src/language.rs::from_str_or_custom` vaza **todo** id de linguagem desconhecido; `plugin/external.rs` e `plugin/language.rs` vazam nomes e metadados de comando | **spec vence** — internamento em mapa limitado | M2 | corrigido (parcial) |
| B11 | `panic!`/`unreachable!`/`expect` em caminho de usuário | `oride-core/src/undo.rs:22,28` (invariante de offset), `oride-i18n/src/lib.rs:74,76,667`, `oride-lsp/src/client.rs:533`, `oride/src/main.rs` (demo/stat), `oride-app/src/app/input.rs:1170` | **spec vence** — Prumo: nunca panic para erro de usuário ou config; panic só para invariante de programador | M2 | parcial — medido |
| B12 | Unix-only: `/dev/null` literal, `sh -c`, `/bin/sh`, sem PATHEXT | `oride-git/src/lib.rs` (`diff --no-index`), `oride-app/src/app/actions.rs:1640` (tarefas em background), `oride-terminal/src/lib.rs:435` (fallback de shell), `oride-app/src/health.rs` (busca em PATH) | **spec vence** — `internal/osutil` cross-platform | M2 | corrigido |
| B13 | Código morto: `Component`/`ComponentRegistry` (429 LOC) nunca despachados | `crates/oride-app/src/component.rs` existe mas `run.rs` chama `handle_key`/`handle_mouse` direto | **código vence** — implementar de verdade: é a costura certa para a fase de cliente do harness | M3 | aberto |
| B14 | `menus.rs::default_menus` morto | `crates/oride-app/src/menus.rs:8` (`#[allow(dead_code)]`) | **código vence** — remover | M2 | corrigido |
| B15 | Fuzzy é substring/subsequência sem score | `crates/oride-app/src/browser.rs`, `crates/oride-app/src/app/state.rs:256` (`fuzzy_match`) | **melhoria** — `sahilm/fuzzy` dá score real estilo Sublime/VSCode | M2 | aberto |
| B16 | Testes acoplados ao ambiente: escrevem `KITTY_WINDOW_ID` no env do processo e ficam flaky sob paralelismo | `crates/oride-app/tests/ui_ux.rs:1050-1070` | **spec vence** — injetar capacidade por opção do modelo de teste, não pelo ambiente | M2 | parcial — costura pronta |
| B17 | Grammar dinâmica (`.so`/`.dylib`/`.dll`) não sobrevive ao default sem CGO | `crates/oride-syntax/src/dynamic_grammar.rs` (libloading). Bindings oficiais de tree-sitter em Go linkam a grammar em tempo de build | **removido** com substituição: recompilar com `-tags treesitter`. É o único ponto onde paridade é impossível por decisão de stack | M5 | aberto |
| B18 | Session filename usa `DefaultHasher`, instável entre versões do Rust — sessões "desapareciam" após update | `crates/oride-app/src/session.rs` (antes de M0-e) | **resolvido** — FNV-1a 64 sobre o caminho canônico, com digests golden pinados para o Go reproduzir | M0-e | **corrigido** |
| B19 | **Um caret não representa posição dentro da quebra de linha.** Offset 8 e 9 de `"linha um\r\n"` (o `\r` e o `\n`) colapsam ambos para caret `{0, 8}` | `crates/oride-core/src/buffer.rs::byte_to_caret` clampa `column_bytes` ao conteúdo sem EOL; a implementação Go faz igual e `TestOffsetInsideALineBreakCollapsesToEndOfContent` pina. **Propriedade a preservar, não bug:** a seleção carrega offsets de byte exatos, então editar é correto; o caret é uma visão derivada para exibição | Verificado contra o oráculo por `TestBufferMatchesTheOracle` |
| B20 | **O conjunto de quebras de linha não estava documentado.** O `ropey` conta `\n`, `\r\n`, `\r` **e** U+0085, U+2028, U+2029 — nada no código dizia isso | Descoberto empiricamente: `conformance/cases/editor/line-break-semantics.json` com os três separadores Unicode e um `\n` produz `lines: 5` no oráculo, não 2 | Implementado e agora pinado pelo caso e pelo teste diferencial |
| B21 | **O texto do diagnóstico de regex difere entre os motores.** O `regex` do Rust e o `regexp` do Go descrevem um padrão inválido com palavras diferentes | Caso `search/regex-error.json`: o comportamento é idêntico — erro reportado, zero matches, editor utilizável — e só a redação muda | **divergência aceita**, declarada no caso. Reproduzir a redação do Rust exigiria reimplementar o parser de erro dele | M2 | declarado |
| B22 | **A precedência do `.editorconfig` está invertida na referência.** Ela aplica do arquivo mais próximo para o mais distante, então um `.editorconfig` na raiz do monorepo sobrepõe o do subprojeto — o oposto do que a especificação determina | `crates/oride-config/src/editorconfig.rs:29-44` itera `ancestors_of(file)`, que vai do diretório do arquivo para cima, e cada aplicação sobrescreve a anterior. Só funciona por acidente quando há `root = true` para interromper a subida | **spec vence** — o Go lê a cadeia de fora para dentro, então o mais próximo vence. `root = true` continua interrompendo. Pinado por `TestWithoutRootTrueTheNearestWins` | M1 | **corrigido** |
| B23 | **Um catálogo de idioma parcial deixa a interface em branco.** O catálogo é desserializado com os campos ausentes virando strings vazias | `crates/oride-i18n/src/lib.rs`: `LocaleDefinition` não tem herança. Quem traduz um idioma e não termina vê menus vazios, que parecem defeito e não trabalho inacabado | **melhoria** — o Go preenche os campos que o arquivo não traz a partir do catálogo padrão (`withFallback`). Não é paridade e não se apresenta como tal | M1 | aplicado |
| B24 | **O URI de arquivo é inválido no Windows na referência.** `format!("file://{}", path)` produz `file:///home/x` em Unix (correto por acidente, porque o caminho já começa com barra) mas `file://C:\dir\file` no Windows, que não é um URI | `crates/oride-lsp/src/protocol.rs::path_to_uri`. Sem efeito medido ainda: nenhum caso exercita LSP, porque exigiria um servidor nos fixtures | **melhoria** — o Go monta o URI pela via do `net/url`, e ainda escapa caracteres que não valem em caminho de URI (um nome com espaço produzia um URI inválido). Em Unix o resultado é idêntico | M2 | aplicado |
| B25 | **Os dois backends de busca não ignoram os mesmos arquivos.** `rg` só pula o que o git ignora — e só honra um `.gitignore` que esteja **dentro de um repositório** —, enquanto o fallback em Go pula `target/`, `node_modules/`, `vendor/` e entradas ocultas incondicionalmente | `crates/oride-search/src/walk.rs` vs. `internal/search/project.go` | **divergência declarada** — num projeto real, com `.gitignore`, os dois concordam (há teste que fixa isso); sem `.gitignore` o fallback acha menos. Medida, não suposta: o primeiro teste que escrevi afirmava que concordavam e falhou | M2 | declarado |
| B26 | **Detecção do shell interativo por substring.** A referência liga `-i` se o caminho *contiver* `bash`, `zsh` ou `fish` — então `/home/user/bashrc-tools` ganharia `-i` | `crates/oride-terminal/src/lib.rs:59-67` | **divergência declarada e deliberada**: o Go compara o nome base exatamente (`filepath.Base(shell) == "bash"`). Para qualquer shell real o resultado é idêntico; a diferença só aparece para um caminho cujo nome não é exatamente um shell mas contém a substring, que não é um shell. Divergência teórica, medida por teste | M2 | declarado |

---

## 3. Correções já aplicadas

| Item | O que mudou |
|---|---|
| B7 | `--headless` dá lugar ao subcomando `conformance run`, com `--json` reservando stdout e diagnósticos no stderr |
| B18 | `workspace_digest` FNV-1a 64 com três digests golden (`/tmp/oride`, entrada vazia, `/tmp/test_dir_alpha`) que o porte Go precisa reproduzir |
| H1 | **O dump vazava o ambiente.** A linha raiz da árvore exibia o basename do diretório do workspace, então uma fixture gravada em `/tmp/a` falhava em `/tmp/b`. Normalizada para `<workspace>` (`tree_row_name`), e o `run_case` passou a canonicalizar o workspace antes de comparar — `ProjectTree::open` canonicaliza, e comparar um caminho canonicalizado com um não-canonicalizado quebraria atrás de symlink. Encontrado por `TestOracleIsDeterministic`, que agora prova a propriedade em vez de assumi-la |
| H2 | O harness recusa um passo ambíguo (`kind: action` com `chord` também preenchido) e um passo que não afirma nada (`text` vazio). O oráculo ignorava o payload extra em silêncio, então o caso parecia testar uma coisa e testava outra |
| H3 | **O gerador de tabelas engoliu um binding em silêncio.** `("ctrl+\"", "toggle_terminal")` no Rust não casava com o extrator por causa da aspa escapada: 97 bindings gerados em vez de 98, e `Ctrl+"` deixaria de funcionar sem aviso. Encontrado por `TestDefaultKeymapMatchesTheOracle` — o **primeiro teste diferencial real** do repositório, que compara o keymap Go contra o dump do oráculo em vez de comparar o oráculo consigo mesmo. Corrigido com parsing de escape de literal Rust, escape de string Go na saída, e uma regra nova no gerador: **linha que parece um binding e não é interpretada é erro fatal**, nunca pulo silencioso |

## 4. Adições que a migração traz

O porte não é só transporte. Estes artefatos passam a existir e sustentam o
resto:

| Artefato | Papel |
|---|---|
| `Action::ALL` + `Action::id()` | Torna id↔variante bidirecional e verificável. Antes, o `enum` e a string TOML eram duas identidades unidas por `parse_action` e um `ActionParseError` que podia discordar |
| `conformance` (modo Rust) | Oráculo diferencial scriptável. Antes não havia entrada headless: `--headless` era stub |
| `conformance/` (Go) | Suíte diferencial com gate de paridade por feature |
| Tokens semânticos de tema | Camada primitive → semantic → component, exigida pela visual-constitution. Habilita o tema `no-color` enforçado |
| `docs/ui-ux/` | `interaction.md`, `interface-map.json`, `state-matrix.json`, `flows.md`, `theming.md`, `design-tokens.json` |
| Tokens / contrato de documentação | `docs/PRUMO.md`, `SOURCE_MAP.json`, `AUTHORITY_MAP.json`, `glossary.json`, `lifecycle.json` |
| Workforce | Skills e agents de §5 do plano, no formato do Prumo |

## 5. Achados da instalação do Prumo

O Prumo é a ferramenta que governa o projeto; quando ela erra, o erro vira fato
do projeto se ninguém anotar. Estes achados foram observados ao instalar a
governança, com `prumo` v0.6.0.

| # | Achado | Evidência | O que fizemos |
|---|---|---|---|
| P1 | **`prumo adopt <sub> [path]` está quebrado nesta build.** A forma posicional consome o próprio subcomando como caminho | `prumo adopt facts "$PWD"` → `repository.root == "facts"`, 0 arquivos. `prumo adopt facts --path "$PWD"` → 462 arquivos | Usamos `--path` em tudo e registramos aqui |
| P2 | **Falso positivo na adoção: `persistence: ["sql"]`.** O Oride não tem banco | A única ocorrência de `sql` é um literal na lista de extensões de `crates/oride-search/src/walk.rs:166` | Não entrou no `prumo.json`; o perfil declara só `toml`/`json` |
| P3 | **A adoção não viu o testing do Rust.** Reportou `testing: ["go-test"]` | 50 arquivos Rust com `#[test]` mais `cargo test --workspace` no CI | Perfil e manifesto não dependem desse campo |
| P4 | **Achados reais e úteis.** Contratos de documentação ausentes: `cli.reference` e `installation.lifecycle` | `prumo adopt facts --path .` → `doc_coverage.required_contracts` com 7, 2 sem fonte | **Dívida registrada** — a ser paga no M6 |
| P5 | **`prumo compile` lista como seus arquivos que não reescreveu.** `.agents/rules/*` e `.agents/subagents/*` continuam sendo o conteúdo Atlas, mas aparecem em `managed: true` | `git diff --stat -- .agents/rules .agents/subagents` vazio depois de `compile --target antigravity`, com os caminhos presentes em `.prumo-generated.json.created_paths` | Conteúdo herdado fica até ser reescrito com as roles novas; **não** o tratamos como gerado |
| P6 | **`created_paths` usa caminhos absolutos** | `.prumo-generated.json.created_paths[0] == "/home/raillen/…/GEMINI.md"` | Manifesto não é portável entre máquinas; não usamos como fonte |
| P7 | **`adopt scaffold` não escreve nada.** Os next-steps mandam rodar `prumo adopt` sem `--audit-only`, mas esse flag não existe em `adopt --help` | `git status --porcelain` idêntico antes e depois | `prumo.json` e `project-profile.json` foram escritos à mão e validados por `prumo validate` |
| P8 | **Imprecisão do catálogo na seleção.** Quatro skills entraram por tokens legítimos do Oride, mas não servem a um editor de terminal | `editor-tooling` ("desktop engine editors, scene inspectors, gizmo overlays") via `code-editor`; `input-handling` ("gamepad support") via `input`; `zoom-reflow` (400% zoom, viewport de 320px) via `accessibility`; `wireframe-styleguide` via `design-tokens` | **Mantidas.** Ajustar o perfil para evitá-las seria esconder a imprecisão e declarar menos features do que o produto tem |
| P9 | **Dois tokens genéricos causaram seleções erradas minhas** | `conformance` sozinho puxou `mcp-tooling` (Oride não tem servidor MCP); `planning` puxou `traycer-orchestration` (Oride usa Prumo) | Trocados pelo token exato `surface-protocol-conformance`; `planning` removido em favor de `goals`/`handoff`/`context-management` |
| P10 | **`prumo` bloqueia no stdin** | `prumo-agent --version` pendura sem `</dev/null` | Toda invocação precisa de stdin fechado — inclusive em CI |
| P11 | **`workforce sync` não instala agentes nem recipes**, apesar da descrição dizer "Synchronizes skills, agents, and recipes". Só `--skills` existe | Depois de `workforce sync --offline --path .`, `.ai/agents/` continua com apenas `manifest.json` e `README.md`; nenhum `AGENT.md` é instalado. O corpo em prosa dos agentes existe em `src/prumo/resources/workforce/agents/<id>/AGENT.md` no FS embutido, mas nenhum comando o expõe | `scripts/workforce.sh` gera `.ai/agents/<id>/AGENT.md` a partir de `prumo --json explain agent <id>` (o contrato estruturado). O corpo em prosa continua indisponível — **lacuna registrada** |
| P12 | **`workforce sync` sem `--skills` trunca o manifest** para um conjunto default (33 skills), derrubando as skills locais da seleção | Rodar `sync --offline --path .` levou o manifest de 66 para 33 skills | `scripts/workforce.sh` sempre passa `--skills` explícito e re-adiciona as locais depois |
| P13 | **`prumo compile` não propaga os corpos de agente** para `.agents/subagents/` | `git diff --stat -- .agents/subagents` vazio após `compile`; os três stubs herdados do Atlas (2 linhas, sem frontmatter) permanecem | O contrato canônico vive em `.ai/agents/`; o adaptador não é editado à mão, porque `compile` o declara gerenciado |
| P14 | **O binário `prumo` foi substituído no meio da sessão** por um build que resolve os recursos embutidos **pelo disco, relativo ao diretório atual** | Tamanho passou de 10.486.025 para 19.471.803 bytes (`~/.local/bin/prumo`, mtime `set 24 22:34`). De `/tmp` ou do repo do Oride: `validate` falha com `schema registry: open schemas: no such file or directory` e `framework-check` lista 11 schemas e 7 adapters ausentes. De dentro do checkout do Prumo, `validate` passa e `doctor` fica limpo para o Oride | **O estado do projeto está intacto** — o que mudou foi a resolução de recursos da ferramenta, não os artefatos. Enquanto isso durar, os gates de governança só funcionam rodando de dentro do checkout do Prumo, e o CI precisaria de um binário com os assets embutidos |

**Regra que sai disso:** os manifestos de workforce (`.ai/agents`, `.ai/recipes`,
`.ai/orchestration`) e as roles de agente são **projeção da resolução
determinística** de `prumo resolve project-profile.json`, não transcrição manual.
Um comando regera tudo:

```bash
scripts/workforce.sh [--target antigravity]
```

Ele resolve o perfil, sincroniza as skills do catálogo, re-adiciona as **skills
locais** que o catálogo não cobre, gera as roles de agente a partir de
`prumo explain agent`, compila o adaptador e roda `validate` + `doctor`. O passo
de re-adicionar as locais existe porque `workforce sync` reescreve o manifest
para exatamente o que recebeu (P12) — sem ele, rodar o sync apaga as skills
autorais da seleção.

## 6. Como um item sai deste ledger

1. A decisão é aplicada na implementação Go **e**, quando for o caso, no
   oráculo Rust.
2. O caso de conformance que reproduz a divergência é adicionado ou ajustado, e
   passa a citar o número do item no campo `description`.
3. O estado muda para `fechado` com a referência do commit.
4. Um item `spec vence` só fecha quando o teste que impede a reincidência
   existir — não basta a correção.
