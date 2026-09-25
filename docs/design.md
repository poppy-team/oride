# Arquitetura e Design: TUI Code Editor & Mini-IDE em Rust

## Context

Queremos um **editor de código em terminal (TUI)** modular, rápido e extensível, inspirado na ergonomia do [Micro](https://micro-editor.github.io/) e [Helix], concebido em **Rust** com foco em linguagens de desenvolvimento de sistemas (Rust, C, Bash/Shell), documentação rica (Markdown) e ecossistema Ori (`ori-lang`), além de extensibilidade para linguagens modernas via highlight Tree-Sitter, task runner e LSP.

Diferente de editores puramente textuais simples, o **Oride** inclui:

- Árvore de projeto navegável (criar pastas/subpastas/arquivos, badges git)
- Terminal embutido colapsável/expansível por atalho (`portable-pty`)
- Suporte nativo a **Rust, C, Bash, Markdown e Ori** (+ plugins e gramáticas dinâmicas)
- Modos de edição híbridos: edição padrão intuitiva e modo modal estilo Vim (`:normal`, `i`, `v`, `V`, `:`)
- Task runner declarativo via `tasks.toml` e sistema preventivo de diagnóstico (`:health`)
- Configuração de atalhos, temas TOML e internacionalização dinâmica
- Ícones Nerd Fonts com fallback limpo
- Arquitetura **modular em Rust** (crates + traits desacopladas)

**Decisões de Design:**

| Decisão | Escolha |
|----------|---------|
| Superfície | **TUI** (Ratatui + Crossterm) |
| Arquitetura | **Workspace Cargo modular** com separação estrita entre core e UI |
| Escopo | **“IDE mini” completa** (multi-tab, splits, terminal PTY, git status, multi-LSP sob demanda, temas, keymaps, task runner, palette) |

**Desacoplamento de Linguagens:** toda inteligência avançada de linguagem (diagnostics, completion, hover, goto, format) opera via clientes stdio de **LSP no PATH** (`rust-analyzer`, `clangd`, `ori-lsp`), sem embutir compiladores pesados dentro do binário do editor.

**Nome do projeto:** `oride` (Ori + IDE). Repositório oficial: `https://github.com/ori-team/oride`.

---

## Recomendação de stack

| Camada | Crate / abordagem | Por quê |
|--------|-------------------|---------|
| UI | `ratatui` + `crossterm` | Padrão TUI Rust 2024–26 |
| Buffer de texto | `ropey` | Rope eficiente, undo-friendly |
| Highlight | `tree-sitter` + grammars embutidas | Incremental, AST preciso (Rust, C, Bash, MD, etc.) |
| Terminal embutido | `portable-pty` + scrollback com CSI mínimo | PTY real + shell interativo; sem emulador de tela (não há alt-screen nem wrap) |
| Config | **TOML** (`serde` + `toml`) | Comentários, legível, ecossistema Rust; JSON só se export/import for necessário |
| Keymap | crate próprio + TOML | Camadas: defaults → user → buffer-local |
| LSP client | JSON-RPC stdio (sem tower no host se possível) | Consome `rust-analyzer`, `clangd`, `ori-lsp` e outros servers |
| Git | `git2` **ou** subprocess `git status --porcelain` | Status na árvore; porcelain é mais simples e estável no MVP |
| Fuzzy / palette | `nucleo` ou `fuzzy-matcher` | Command palette + file finder |
| Watch FS | `notify` | Reload / refresh da árvore |
| Clipboard | `arboard` + OSC52 (SSH) | Como Micro |
| Ícones | mapa extensão → glyph Nerd Font | Fallback ASCII se fonte não detectada |
| Async I/O | threads + canais; **sem** `tokio` | Event loop UI síncrono + canais |

**Config: preferir TOML** (`~/.config/oride/config.toml` + `./.oride/config.toml` no projeto). Suportar comentários e overlays; evitar JSON como formato primário.

---

## Arquitetura modular (workspace)

```text
oride/                          # novo repositório
  Cargo.toml                    # workspace
  README.md
  docs/
    design.md                   # este plano (cópia canônica)
    plugin-api.md
    keymaps.md
  crates/
    oride-core/                 # Document, Buffer(rope), Selection, Undo, tabs
    oride-fs/                   # Project tree, create/rename/delete, watch
    oride-config/               # load/merge TOML, schema versionado, temas
    oride-i18n/                 # catálogos de mensagens (pt-BR, en)
    oride-keymap/               # KeyChord → Action (enum + string para plugins)
    oride-syntax/               # LanguageId + tree-sitter queries
    oride-lsp/                  # client multi-server (stdio)
    oride-terminal/             # painel PTY, resize, focus
    oride-git/                  # status por path (M/A/D/? )
    oride-search/               # find/replace buffer + projeto (rg opcional)
    oride-plugin/               # trait Plugin + host + built-ins
    oride-ui/                   # widgets ratatui (tree, editor, status, palette, term)
    oride-app/                  # composition, event loop, layout
    oride/                      # binário CLI: `oride [path]`
    lang-rust/
    lang-markdown/
    lang-web/                   # html, css, js
  assets/
    icons.rs                    # extensão → nerd glyph, com fallback ASCII
    themes/
      default.toml
      dark.toml
    grammars/                   # tree-sitter vendored ou build script
```

### Princípios de modularidade

1. **Core sem UI** — `oride-core` não depende de `ratatui`. Testável sem TTY.
2. **Ações como dados** — teclas e command palette disparam `Action` (enum estável + string livre para plugins).
3. **LanguageProvider trait** — highlight, indents, comment string, LSP command, completions “offline”.
4. **Plugin host** no 0.1 = **crates Rust built-in** + trait; **API externa (Lua/WASM)** no 0.2+ (sem travar ABI nativo frágil).
5. **Fail closed** — LSP/Git/PTY com falha viram status line, não crash.
6. **CLI first via PATH** — inteligência de linguagens via binários padrão no PATH (e.g. `rust-analyzer`, `clangd`, `ori-lsp`), sem embutir compiladores pesados.

### Fluxo de eventos (alto nível)

```text
crossterm events ──► App
                      ├─ Keymap → Action
                      ├─ focus: Tree | Editor | Terminal | Palette | Dialog
                      ├─ DocumentStore (tabs, dirty, undo)
                      ├─ channels ◄── threads: LSP / PTY / notify / git
                      └─ render(oride-ui)
```

Layout padrão:

```text
┌──────────┬─────────────────────────────┐
│  Tree    │  Tabs | buffer              │
│  📁 src  │  1:main.oris ●              │
│  📄 .oris│  … editor …                 │
│          ├─────────────────────────────┤
│          │  Diagnostics / search hits  │  (opcional, toggle)
├──────────┴─────────────────────────────┤
│  Terminal (toggle: Ctrl+`)  ▾ / ▸     │  altura 0 | 30% | maximizado
├────────────────────────────────────────┤
│  status: mode | lang | git | lsp | ln  │
└────────────────────────────────────────┘
```

---

## Funcionalidades pedidas (0.1)

| # | Feature | Notas de implementação |
|---|---------|------------------------|
| 1 | Editor multi-tab | `DocumentId`, dirty `●`, close com confirm se dirty |
| 2 | Árvore de projeto | Expand/collapse, keyboard + mouse se disponível |
| 3 | Criar pasta/arquivo | Dialog inline ou prompt na status; validar path |
| 4 | Terminal embutido | Toggle atalho; resize altura; focus cycle |
| 5 | Highlight nativo | Rust, C, Bash, Markdown, Ori (+ gramáticas dinâmicas) |
| 6 | Completions / suggestions | LSP semântico + keywords offline por linguagem |
| 7 | Config TOML | theme, language, keys, terminal shell, tree width |
| 8 | Ícones | `oride-fs/src/icons.rs` + Nerd Font; fallback ASCII |
| 9 | Temas | cores UI + scopes syntax |
| 10 | Keymaps custom | rebind de Actions; layers |
| 11 | Find/replace | buffer atual; regex opcional |
| 12 | Git status na árvore | M/A/D/? cores no nome |
| 13 | Multi-LSP sob demanda | spawna servidores sob demanda (`[lsp.servers]`) |
| 14 | Command palette | fuzzy de Actions + “open file” |

---

## Funcionalidades **recomendadas** (além do pedido)

Prioridade para caber no 0.1 “IDE mini” sem virar monólito:

### P0 — incluir no 0.1 (alto valor / baixo-médio custo)

| Feature | Por quê |
|---------|---------|
| **Command palette** (`Ctrl+Shift+P`) | Descoberta de comandos; reduz dependência de memorizar keys |
| **Fuzzy open file** (`Ctrl+P`) | Essencial com árvore; fluxo Micro+VS Code |
| **Undo/redo por documento** | Não negociável em editor de código |
| **Status line** | Lang, LSP ready/error, branch, dirty, Ln:Col, encoding |
| **Painel de diagnostics** | Lista erros do LSP; jump com Enter |
| **Save all / “modified” indicator** | Multi-tab seguro |
| **Reload se arquivo mudou no disco** | `notify` + prompt se dirty |
| **Soft wrap** (toggle; default on em `.md`) | Markdown legível |
| **Comment toggle** (`Ctrl+/`) | Por LanguageProvider |
| **Bracket match + auto-indent** | Edição diária |
| **Clipboard sistema + OSC52** | Local e SSH |
| **Which-key / help overlay** (`Ctrl+G` ou `?`) | Onboarding estilo Micro |
| **Session leve** | Reabrir última pasta + lista de tabs (sem full workspace state) |
| **`.editorconfig` básico** | indent_size, tab/spaces, eol |

### P1 — logo após 0.1 (vale planejar a API já)

| Feature | Por quê |
|---------|---------|
| **Split horizontal/vertical de buffers** | Micro tem; complexifica layout — após layout estável |
| **Search in project** (`rg` ou walk+grep) | “Find in files” |
| **Multi-cursor** | Marca do Micro; caro em rope+LSP — pós-0.1 |
| **Markdown preview** (split read-only) | Diferencial DX docs — ver `docs/guides/pt/markdown.md` § Futuro |
| **Code-fence language highlight** | Injections em ` ```lang ` | 
| **MDX com JSX real** | Além do highlight MD genérico |
| **Git gutter** (linha) + stage hunk | Além do ícone na árvore |
| **Format on save** | Via LSP `formatting` |
| **Rename / new file from tree context** | UX árvore completa |
| **Plugins externos (Lua ou WASM)** | Extensibilidade real sem recompilar |
| **DAP / debugger** | Futuro (desacoplado) |
| **Detecção automática de Nerd Font** | Mensagem amigável no first-run |

### Explicitamente **fora** do 0.1

- Debugger completo, remote collab, AI chat embutido, marketplace de plugins, GUI, multi-root workspace, remote SSH host (só editar via SSH local com OSC52).

---

## Plugins: modelo em duas camadas

### 0.1 — Built-in providers (Rust)

```rust
// oride-plugin (conceitual)
pub trait LanguageProvider: Send + Sync {
    fn id(&self) -> &str;                    // "rust"
    fn extensions(&self) -> &[&str];         // [".rs"]
    fn highlight_query(&self) -> Option<&str>;
    fn comment_token(&self) -> Option<&str>; // "//"
    fn lsp_command(&self) -> Option<LspSpawn>; // ["rust-analyzer"]
    fn offline_completions(&self, ctx: &CompletionCtx) -> Vec<CompletionItem>;
}

pub trait Plugin: Send + Sync {
    fn name(&self) -> &str;
    fn on_action(&mut self, action: &str, ctx: &mut PluginCtx) -> PluginResult;
    fn commands(&self) -> &[CommandMeta];    // aparecem na palette
}
```

Registro estático no binário (ou `inventory` / explicit `register` em `main`):

- Providers nativos para Rust, C, Bash, Markdown, Ori, etc.
- `lang-markdown`, `lang-web` — highlight + indent; sem LSP no 0.1 (opcional `vscode-html` etc. depois)

### 0.2+ — Host externo

| Opção | Prós | Contras |
|-------|------|---------|
| **Lua** (como Micro) | Familiar, scripts simples | FFI/embed, sandbox fraco |
| **WASM** (estilo Lapce/Extism) | Sandbox, multi-lang | Mais infra |
| **Rhai** | 100% Rust embed | Menos ecossistema |

**Recomendação:** desenhar `PluginCtx` estável no 0.1; escolher **WASM ou Lua no 0.2** sem quebrar Actions/TOML.

---

## Configuração (TOML)

Exemplo de superfície (não normativa até implementar):

```toml
# ~/.config/oride/config.toml
theme = "dark"
soft_wrap = false
show_line_numbers = true
icons = true                 # false → ASCII
file_icons = true

[editor]
tab_size = 4
insert_spaces = true
format_on_save = false

[terminal]
shell = "/bin/zsh"           # default $SHELL
default_height = 10          # linhas; 0 = colapsado
toggle_key = "ctrl+`"

[tree]
width = 28
show_hidden = false
git_status = true

[lsp.servers]
rust = ["rust-analyzer"]
c = ["clangd"]

[keys]
"ctrl+s" = "save"
"ctrl+p" = "open_file_fuzzy"
"ctrl+shift+p" = "command_palette"
"ctrl+`" = "toggle_terminal"
"ctrl+b" = "toggle_tree"
"ctrl+f" = "find"
"ctrl+h" = "replace"
"f2" = "rename_in_tree"      # p1 se não der tempo

[theme.ui]
background = "#1a1b26"
foreground = "#c0caf5"
# ...
```

Projeto local: `.oride/config.toml` sobrescreve user (merge profundo de seções).

---

## Integração LSP & Ferramentas de Sistema

| Capacidade | Fonte |
|------------|--------|
| Diagnostics / hover / goto / completion / format | Clients stdio preguiçosos por linguagem (`[lsp.servers]`) |
| Servidores recomendados | `rust-analyzer` (Rust), `clangd` (C/C++), `bash-language-server` (Bash), `ori-lsp` (Ori) |
| Grammar highlight | Tree-sitter estático + fallback léxico para novas linguagens |
| Run / Tasks | Task Runner declarativo (`tasks.toml`) e Terminal PTY integrado |

O editor **não** linka compiladores estáticos em seu binário, garantindo leveza extrema, portabilidade e ciclos de vida independentes.

---

## Plano de implementação por fatias (PRs)

Cada PR = um conceito; ordem topologicamente segura.

### Fase 0 — Fundação (1–2 PRs)

| ID | Entrega | Gate |
|----|---------|------|
| **P0.1** | Workspace Cargo, bin `oride`, `oride-core` (rope + undo + seleção), testes unitários | `cargo test` |
| **P0.2** | `oride-ui` + loop: abrir arquivo, editar, salvar, quit, status line | demo TUI manual |
| **P0.3** | `oride-config` TOML + `oride-keymap` + tema default | rebind `ctrl+s` via TOML |

### Fase 1 — IDE shell

| ID | Entrega | Gate |
|----|---------|------|
| **P1.1** | Multi-tab + dirty + close confirm | 3 tabs smoke |
| **P1.2** | `oride-fs` árvore + expand + open file + create file/dir | criar `src/a.oris` pela UI |
| **P1.3** | Ícones + git status na árvore | repo git real |
| **P1.4** | Terminal embutido toggle/focus/resize | `ls` interativo |
| **P1.5** | Command palette + fuzzy open | `Ctrl+P` / palette |

### Fase 2 — Linguagens

| ID | Entrega | Gate |
|----|---------|------|
| **P2.1** | `oride-syntax` tree-sitter + MD/HTML/CSS/JS | highlight visual |
| **P2.2** | Provider de linguagens de sistema + grammar | Rust, C, Bash, Markdown |
| **P2.3** | Comment toggle, indent, soft wrap md | edição confortável |
| **P2.4** | Find/replace no buffer | regex opcional |

### Fase 3 — Inteligência

| ID | Entrega | Gate |
|----|---------|------|
| **P3.1** | `oride-lsp` client + diagnostics panel | LSP stdio desacoplado |
| **P3.2** | Hover / completion / goto (UI) | projetos de exemplo |
| **P3.3** | Format (LSP) + format on save config | formatação via LSP |

### Fase 4 — Polimento 0.2

| ID | Entrega | Gate |
|----|---------|------|
| **P4.1** | Help overlay, clipboard OSC52, editorconfig, session | checklist UX |
| **P4.2** | README, install script, `docs/plugin-api.md` (trait 0.1) | first-run docs |
| **P4.3** | Suite de testes headless (core/keymap/config/fs) + smoke script | CI verde |

**SemVer produto editor:** começar `0.1.0-alpha` até P3 estável; `0.1.0` com P4.

### Pós-0.1 (decisões atuais)

Ver plano canônico: **[`docs/planning/post-0.1-roadmap.md`](planning/post-0.1-roadmap.md)**.

**Roadmap atual (normativo):** [`docs/planning/alpha6-roadmap.md`](planning/alpha6-roadmap.md)  
(`0.2.0`). P5–P9, languages L1 (Rust, Python, TS, D, Lua, etc.), MD links/imagens, multi-LSP sob demanda L2, SCM sync (pull/push/ahead-behind) e hygiene anti-bloat já entregues e validados.

---

## Estrutura de diretórios no primeiro commit

```text
oride/
  Cargo.toml
  README.md
  LICENSE
  crates/oride-core/...
  crates/oride/...
  docs/design.md
```

Scaffold inicial só com `oride-core` + binário vazio/`hello buffer` — depois preencher conforme fases.

---

## Riscos e mitigações

| Risco | Mitigação |
|-------|-----------|
| Terminal embutido (PTY) complexo no TUI | Isolar crate cedo; fallback “abrir `$SHELL` externo” se PTY falhar |
| Tree-sitter build / grammars | Vendor grammars + `cc` build; CI com cache |
| Multi-cursor / splits atrasam 0.1 | Fora do escopo; API de Selection já lista de ranges se possível |
| Acoplar a compiladores específicos | Só LSP via PATH; binário desacoplado |
| Plugin ABI nativo | Não expor `cdylib` no 0.1; só traits internos |
| Escopo “IDE mini” estourar | Cortes: multi-cursor, project search, preview md, rename tree → P1 |

---

## Verification (como validar o 0.1)

1. **Unit:** `oride-core` (insert/delete/undo), keymap resolve, config merge, tree create path.
2. **Integração headless:** abrir buffer de fixture, apply actions sem TTY (`App::from_test`).
3. **Manual TUI checklist:**
   - Abrir pasta com projeto de código
   - Highlight `.rs` / `.c` / `.sh` / `.md` / `.html`
   - Criar subpasta e arquivo pela árvore
   - Terminal: toggle, executar comandos, colapsar
   - Introduzir erro de sintaxe/tipo → diagnostic no painel via LSP
   - Completion / hover / goto
   - Rebind tecla em `config.toml` e reiniciar (ou hot-reload se implementado)
   - Git: modificar arquivo → status `M` na árvore
   - Multi-tab dirty save
4. **CI:** `cargo fmt`, `clippy -D warnings`, `cargo test --workspace` no repo `oride`.

---

## Resumo da Arquitetura
 
- **TUI modular em Rust** com workspace de crates desacoplados e **Actions** como lingua franca.
- **TOML** para config/keymaps/themes e tarefas (`tasks.toml`).
- **Tree-sitter + LSP via PATH** para inteligência de linguagem eficiente e desacoplada.
- **Task Runner e Diagnóstico Integrados** para experiência completa de desenvolvimento no terminal.
- **Extensibilidade Dinâmica:** temas TOML sem recompilação, catálogos de tradução i18n externos e plugins modulares.
