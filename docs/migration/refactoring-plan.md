# Plano de refatoração, saneamento e dependências

**Status:** ativo · **Substitui:** nada · **Relacionado:** `docs/migration/parity-ledger.md`, plano de migração aprovado

Este documento nasceu de uma constatação desconfortável: o gate media frames,
testes e paridade, e **nenhum dos três pergunta se uma tecla faz alguma coisa**.
Uma TUI em que não se fecha o programa, não se navega a árvore e não se abre um
menu passava em todos os nove checks. O que segue é o que precisa mudar para que
isso não volte.

---

## 1. Diagnóstico

### 1.1 Configuração portada e inerte

O padrão se repetiu e ninguém notou, porque nada falha: **configuração é parseada,
testada e não tem consumidor**.

| Config | Papéis | Consumido por | Veredito |
|---|---|---|---|
| `Syntax.*` | **20** | o tema, que expõe **um** (`Comment`) | **19 sem motor.** Não existe realce de sintaxe no Go |
| `Markdown.*` | — | só `config/load.go` | parseado, ninguém age |
| `Terminal.*` | — | só `config/load.go` | idem; não há painel de terminal na TUI |
| `LSP.*` | — | só `config/load.go` | idem; `internal/lsp` existe e a app nunca sobe cliente |
| `Languages.*` | — | **ninguém** | nunca referenciado |

O caso do `Syntax` é o mais grave, e não é um item de M4: o plano colocava
"syntax core (chroma mapping + markdown list continuation)" no **M1**. Falta desde
o primeiro marco.

O `i18n` já esteve nesta lista — portado, completo, sem consumidor — até a
`menubar` passar a lê-lo. É o mesmo defeito, ainda em quatro lugares.

### 1.2 Componentes escritos à mão em paralelo com a biblioteca

O plano nomeava `bubbles/v2` v2.2.1. Duas das três bibliotecas foram adicionadas e
a do meio não — e **nada falhou**: nem build, nem teste, nem gate.

| Componente | Deveria ser | Estado |
|---|---|---|
| `overlay` (palette, which-key, ajuda) | `list` | ✅ substituído |
| `findbar` (campos) | `textinput` | ⏳ à mão |
| `editorview` (scroll) | `viewport` | ⏳ à mão |
| `tree` (painel) | `tree` | ⏳ à mão |
| which-key / ajuda | `key` / `help` | ⏳ à mão |

### 1.3 Três implementações de largura

A classe de bug mais frequente deste projeto — **cinco ocorrências** — é "um valor
medido numa unidade e afirmado noutra". Hoje convivem três medidas:

1. `layout.Width` — minha, `runewidth` com **ambíguo = largo** (conservadora)
2. `bubbles` — `clipperhouse/displaywidth`
3. `x/ansi` — `StringWidth`

Três medidas que podem discordar é o terreno exato do sexto bug. **Consolidar numa
só é pré-requisito** de qualquer trabalho novo de layout.

### 1.4 O que a matriz de golden *não* pode pinar

O campo de filtro do `bubbles/list` **pisca**: a saída depende do relógio, e um
golden passaria numa máquina rápida e falharia numa lenta. Os casos de
sobreposição foram removidos da matriz com o motivo escrito. O mesmo vale para
qualquer componente animado (`spinner`, `progress`, `timer`) — a regra é:
**golden só para o que é função pura do estado.**

---

## 2. Saneamento

Ordem de execução. Cada item é verificável e nenhum depende do seguinte, exceto
onde anotado.

### S1 — Extrair o realce de sintaxe (desbloqueia M4)

Adotar `chroma/v2` e ligar os 20 papéis de `config.Syntax`. É o item de maior
impacto visual e o que torna a configuração honesta.

**Aceite:** um arquivo `.go` aberto mostra tokens coloridos; os 20 papéis têm
consumidor, ou o papel é removido da configuração.

### S2 — Consolidar a medida de largura

Escolher **uma** implementação e apontar as outras para ela. A recomendação é
manter a minha como autoridade de medida da aplicação (a regra conservadora é uma
decisão, não um acidente) e expor uma ponte para o que o `bubbles` mede, com um
teste de propriedade que compare as duas em texto largo, ambíguo e com ANSI.

**Aceite:** um teste afirma que as duas concordam, ou documenta exatamente onde
discordam.

### S3 — Substituir os quatro componentes restantes

`findbar` → `textinput`; `editorview` → `viewport`; `tree` → `tree`;
which-key/ajuda → `key` + `help`. Um commit por componente, como o do `overlay`.

**Aceite, por componente:** o código manual some, os testes do que deixou de ser
nosso são **apagados** (não adaptados), e o que é nosso ganha teste novo.

### S4 — Regra de aptidão: configuração sem consumidor

Acrescentar a `internal/architecture` uma regra que falhe o build quando um campo
de configuração não tem leitor fora do `config/`. Um mapa explícito de exceções
(com motivo) mantém a regra útil em vez de barulhenta.

**Aceite:** reintroduzir `Languages` sem consumidor **falha** o build. Contraprova
obrigatória, como nas outras sete.

### S5 — Teste de fumaça da TUI

O item que faltou. Um teste que dirige a TUI **como um usuário**: abre, anda na
árvore, abre um arquivo, digita, busca, fecha. Não compara frames — exercita
caminhos.

**Aceite:** remover o `Ctrl+Q` **falha** o teste; remover a navegação da árvore
**falha** o teste.

### S6 — Limpeza de código morto

`bubbles` trouxe 29 dependências transitivas. Auditar:
`atotto/clipboard` (o plano listava como **não-candidata**), e qualquer coisa que
o `go mod why` não justifique.

---

## 3. Dependências

### 3.1 Adotar, sem decisão pendente

| Biblioteca | Para | Nota |
|---|---|---|
| `chroma/v2` v2.27.0 | realce de sintaxe | API de token stream, não só string |
| `doublestar/v4` v4.10.2 | globs de busca | substitui o `globAllows` caseiro |
| `go-git/v5/plumbing/format/gitignore` | fecha a divergência **B25** | só o subpacote; o módulo é grande |
| `glamour/v2` | preview Markdown | escolha do plano |
| `golang.org/x/image` v0.46.0 | dimensões de imagem | `image.DecodeConfig` cobre PNG/JPEG/GIF/WebP |
| `fsnotify/fsnotify` v1.10.1 + `bep/debounce` v1.2.1 | reload em disco | `fsnotify` não é recursivo: watch por diretório |
| `teatest/v2` | testes de interação | `x/exp/golden` **já está no árbol**, transitivo |

### 3.2 Exige decisão

**`charmbracelet/x/vt`** — emulador VT real (ledger **B5**). **Não tem versão
marcada**, só pseudo-versão. Decisão: fixar por commit e pinar, aceitando que uma
atualização é manual. Sem isso, o terminal embutido continua sendo o scrollback
com CSI mínimo, que não faz alt-screen nem scrollback real.

**`tree-sitter/go-tree-sitter` + grammars** — **exige CGO**, e o plano fixou
`CGO_ENABLED=0` em todos os alvos. O ledger (B17) já resolveu: *removido, com
substituição por `-tags treesitter`*. Concretamente:

- **padrão:** `chroma` faz o realce, sem CGO
- **opcional:** `-tags treesitter` para realce preciso, documentado no README de build

É a única decisão de stack desta lista, e afeta o release — não o `go.mod`.

### 3.3 Não adotar

`huh/v2` e `wish/v2` (o ch14 do Prumo os nomeia para a TUI **dele**; o Oride não
tem formulários nem servidor SSH), `creativeprojects/go-selfupdate` (roadmap v0.3),
e o `go-git` inteiro — só o subpacote de gitignore.

---

## 4. Waves

### M4 — Paridade de features (em andamento)

**Feito:** find no buffer (modelo + barra + realce + digitação ao vivo), tema com
perfis e degradação, seleção pintada, sessão persistida, sobreposições pelo
`bubbles/list`.

**Falta, na ordem:**

1. **S1–S3** deste documento (realce, largura, componentes) — pré-requisito, não
   paralelo: substituir componentes mexe em tudo que vem depois
2. Painel SCM + diff
3. Splits
4. Multi-cursor com gesto
5. Preview Markdown (`glamour` + `x/image`)
6. Pickers de tema e locale
7. Motor modal Vim (o keymap é **plano**, sem camadas — isto é trabalho real)
8. Task runner, `:health`, jump list, undo tree, buffer picker, surround
9. Disk watch (`fsnotify`)
10. Bracketed paste (`PasteMsg` nativo do Bubble Tea v2)

**Gate M4:** 100% dos casos `parity/` verdes · ledger sem exceções abertas ·
`--stat`/`--demo`/`--help`/`--version` equivalentes · `docs/migration/parity-report.md`
publicado · **mais o teste de fumaça (S5)**, que o gate original não previa.

### M5 — Cross-platform e release

ConPTY e detecção de shell no Windows, PATHEXT, separadores de caminho,
portabilidade de `/dev/null` e `sh -c` — o `oride-osutil` do Rust já tem o
precedente; o Go precisa do equivalente.

Release: matriz Go de 6–8 alvos **sem split glibc/musl**, ldflags com versão de
fonte única, goreleaser, checksums, e a **regeneração de todo o packaging** — hoje
12 arquivos hardcodam `0.2.0` e o nome `crates/` (PKGBUILD, `.spec`, control,
`build-deb.sh`, template Void, `flake.nix`, `default.nix`, `install.sh`,
`install.ps1`, `release.yml`). Self-update (`--update`/`:update`) entra aqui.

**Gate M5:** matriz CI verde · `CGO_ENABLED=0` em todos os alvos, verificado por
script · instaladores testados em container · `SHA256SUMS` conferido ·
`buildGoModule` no Nix funcionando.

### M6 — Fechamento Prumo-nativo

Docs canônicas reescritas (PT como fonte, EN como **projeção** com
`docs/i18n/translations.json` e frescor por digest — hoje `docs/en/` está abridged:
`design.md` 410→85 linhas). Pack `docs/ui-ux/` completo, contratos em
`docs/contracts/`, `docs/media/records.json` com os golden frames, goals fechados,
`crates/` → tag `v0.2.0-rust-final` + `archive/rust-crates/` (**arquivado, não
deletado** — mesmo precedente do ADR 013 do Prumo).

**Camada de tokens semânticos** (`surface`, `panel`, `border-subtle`, `text`,
`text-muted`, `accent`, `success`, `warning`, `error`, `focus`) — **o Oride não tem
essa camada hoje**: temas são `[ui]`/`[syntax]` com nomes de cor diretos. É trabalho
real do M6, e é o que habilita o tema `no-color` obrigatório: todo token não
primitivo resolve para `none`, e uma paleta que **alega** suprimir cor pintando cor
é **erro**, não aviso.

**Gate M6:** `prumo-agent docs audit`, `docs readiness --goal`, `ui verify` e
`docs release` verdes — o último **recusa o release** com evidência stale, que é o
mecanismo que impede a dissonância de voltar.

---

## 5. As cinco regras que este projeto pagou para aprender

Cada uma nasceu de um defeito concreto. Elas valem para qualquer código daqui, e
três delas são verificáveis por teste.

| O defeito | A regra |
|---|---|
| Um marcador de cursor reservava três células enquanto a função de medida do próprio pacote contava duas | **Se você expõe `Width()`, não pode embarcar constante que assume largura.** Derive da medição |
| Uma cor não resolvia e todo teste passava, porque os testes usavam hexadecimal e o produto usava nomes | **Teste a saída renderizada, não o valor parseado.** Nome não resolvido é no-op silencioso |
| Uma condição de corrida real, porque um teste escrevia no ambiente do processo | **Capacidade se injeta, nunca se lê do ambiente** |
| Um retorno antecipado pulava a pintura inteira quando a gutter estava desligada | **Caminho que parece correto por nunca ser exercitado** |
| Dois overlays distintos, e a regra de captura checada em um só — digitar ia para o documento atrás da barra | **Regra checada num lugar é regra não checada** |

---

## 6. O que mudou no processo

**O gate mede o que ele mede, e nada além.** Nove checks verdes coexistiam com uma
TUI inutilizável. Frame, teste e paridade não perguntam se uma tecla faz algo — e
"marco completo" era verdade sobre a medição e mentira sobre o produto.

Três consequências, todas já incorporadas acima:

1. **Teste de fumaça por wave** (S5) — caminhos, não frames
2. **Regra de configuração sem consumidor** (S4) — o tipo de defeito que passa por
   build, teste e revisão
3. **Golden só para o que é função pura do estado** (§1.4) — um golden que depende
   do relógio é um teste que falha por motivo alheio ao produto

E uma que não é verificável por máquina: **declarar desvio do plano quando ele
acontece**. O `bubbles` foi omitido em silêncio porque nada falhou. Uma linha no
relatório de marco teria custado nada.
