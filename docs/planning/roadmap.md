# Roadmap Oficial do Oride — Da Fundação à Versão 1.0 (LTS)

**Status:** Normativo  
**Versão Atual:** **`0.2.0`**  
**Filosofia:** TUI Mini-IDE **contida, enxuta, resiliente (fail-closed)**, com inicialização instantânea e disciplina estrita contra inchaço (anti-bloat).  
**Alvos de Performance:** Inicialização `< 10ms`, latência de digitação `< 5ms (120 FPS)`, consumo de RAM `< 25MB`, binário stripped `<= 12MB`.

---

## 🎯 Princípios Norteadores (Anti-Bloat & Performance)

1. **Contido no Terminal (TUI):** Sem visualizadores em navegadores externos, sem processos webview pesados, sem shells gráficos inflados. Tudo roda no próprio emulador de terminal do usuário via Ratatui e Crossterm.
2. **Um Conceito por PR:** Fatias pequenas, focadas, testadas com TDD e documentadas no mesmo pull request.
3. **Resiliência (Fail-Closed):** Falhas em LSP, Git ou terminal PTY reportam mensagens claras na status bar e nunca travam nem causam pânico no editor.
4. **Disciplina de Binário:** Nenhuma dependência pesada desnecessária. Gramáticas e interpretadores são otimizados, limpos e estripados no release.
5. **Zero-Copy & Streaming:** A UI renderiza apenas o viewport estritamente visível sem alocações dinâmicas redundantes por linha.

---

## 📦 Versões Entregues (Shipped)

### [x] `v0.1.0` — Baseline da Mini-IDE TUI Contida
- [x] **Core do Editor:** Buffers eficientes com Ropey, abas de documentos, histórico de undo/redo com agrupamento de edições, seleção e comentários.
- [x] **Navegação & Árvore:** Árvore de projeto navegável, criação/renomeação/remoção de pastas e arquivos, ícones Nerd Fonts com fallback ASCII.
- [x] **Terminal Embutido:** Terminal interativo com PTY real (`portable-pty`), redimensionamento e chaveamento de foco.
- [x] **Busca Integrada:** Busca em buffer (case, acentos, regex) e busca recursiva no projeto via ripgrep/crawler interno.
- [x] **Git Básico:** Status SCM, badges visuais na árvore de arquivos, diff somente leitura e comando de blame na status bar.
- [x] **Markdown Preview:** Renderização de tabelas, cabeçalhos, listas e code fences com realce no terminal.
- [x] **LSP Multi-Server:** Cliente stdio JSON-RPC sob demanda para autocompletion, hover e goto definition.
- [x] **Layout & Splits:** Divisão horizontal e vertical de janelas, múltiplos cursores e suporte opcional ao mouse (`mouse = false` por padrão).

### [x] `v0.2.0` — Linguagens First-Class, Mídia, Git Sync & Ferramentas
- [x] **Linguagens First-Class:** Suporte nativo para Rust, C, Bash, Markdown, Ori, Python, TypeScript/TSX, HTML, CSS, D, Lua, Nim e Ruby com Tree-Sitter ou fallback léxico contido.
- [x] **Markdown com Mídia no Terminal:** Suporte aos protocolos gráficos de terminal (Kitty Graphics, Sixel, iTerm2) com visualização de imagens locais e links clicáveis no navegador do sistema.
- [x] **Git Sync & Staging CLI:** Staging (`s`), unstage (`u`), prompt de commit (`c`), push (`P`), pull (`p`) e contador de commits à frente/atrás na barra de status.
- [x] **Persistência de Sessão:** Restauração transparente de scroll, splits, arquivos abertos e largura da árvore em `.oride/session.toml`.
- [x] **Busca e Substituição com Globs:** Substituição no projeto inteiro (`Ctrl+Shift+F`) com filtragem por globs (`*.rs`, `!target/**`).
- [x] **Edição Modal Estilo Vim:** Modos `Normal`, `Insert`, `Visual`, `VisualLine` e linha de comando `:`.
- [x] **Task Runner Integrado:** Execução declarativa de scripts em `tasks.toml` com interpolação de variáveis (`$FILE`, `$FILE_NAME`, `$FILE_STEM`, `$FILE_DIR`, `$WORKSPACE`, `$LINE`, `$COL`).
- [x] **Diagnóstico de Ambiente (`:health`):** Verificador preventivo de ferramentas, compiladores e servidores LSP no `$PATH`.
- [x] **Internacionalização Dinâmica (i18n):** Catálogos externos TOML (`pt-BR`, `en-US`) e guias de temas customizados.
- [x] **Empacotamento Multi-Distro & Instaladores Universais:** Pacotes para Arch Linux (`PKGBUILD`), Debian/Ubuntu (`.deb`), Fedora (`.spec`), Void e Nix (`flake.nix`), instaladores one-line para Linux, macOS e Windows.

---

## 🚀 Próximos Slices Planejados (Até a Versão 1.0)

### [ ] `v0.3.0` — Extensibilidade Leve & Ergonomia de Edição
- [ ] **Motor de Plugins em Lua (`mlua`):** Interpretador Lua 5.4 embutido permitindo scripts em `~/.config/oride/plugins/*.lua` com acesso a comandos da palette, buffer e eventos (`on_open`, `on_save`).
- [ ] **Auto-Pairing Inteligente de Delimitadores:** Inserção em pares de `()`, `[]`, `{}`, `""`, `''` e ```` `` ````, com skip-over ao digitar fechamento e wrap automático de seleções ativas.
- [ ] **Text Objects no Modo Modal:** Comandos de ação por escopo estilo Vim (`ci"`, `da(`, `yiw`, `vi{`).
- [ ] **Scrollback Navegável no PTY:** Buffer circular de histórico de saída do terminal embutido com rolagem por teclado e mouse.
- [ ] **Outline de Símbolos do Buffer via LSP:** Seletor de funções, tipos e variáveis do documento atual (`Ctrl+Shift+O`).
- [ ] **Comando de Auto-Update (`oride --update` / `:update`):** Verificação automatizada de releases no GitHub e atualização in-place do executável para instalações realizadas via instalador universal ou release do GitHub.

### [ ] `v0.4.0` — Performance Extrema & Gestão de Arquivos Grandes
- [ ] **Large File Mode (> 10MB / > 100MB):** Detecção preventiva de arquivos pesados com fallback automático para visualização em streaming via chunks do Ropey, desativando Tree-Sitter e wrap de linhas para evitar congelamento da UI.
- [ ] **Zero-Copy Viewport Rendering:** Renderização direta de fatias de texto (`&str`) para o frame do Ratatui, eliminando alocações intermediárias de `String` por linha.
- [ ] **Lazy Initialization de Gramáticas:** Inicialização sob demanda de parsers e queries Tree-Sitter no primeiro arquivo da linguagem aberto (tempo de boot `< 5ms`).
- [ ] **Staging Interativo de Hunks no Git:** Visualização e staging/unstaging por blocos individuais (`hunks`) diretamente na interface SCM do editor.

### [ ] `v0.5.0` — Inteligência de Código Avançada & Refactoring Contido
- [ ] **LSP Code Actions & QuickFix (`Alt+Enter`):** Menu interativo para aplicação instantânea de sugestões do compilador (importações ausentes, correções de lifetime, lints).
- [ ] **Rename Symbol no Projeto (`F2`):** Renomeação semântica com propagação automática via `workspace/applyEdit` do LSP em múltiplos buffers.
- [ ] **Snippets Leves com Tabstops:** Expansão de snippets do LSP com preenchimento guiado por `$1`, `$2`, `$0` e navegação rápida via `Tab`.
- [ ] **Navegação Rápida de Diagnósticos (`Alt+N` / `Alt+P`):** Salto direto para o próximo erro ou warning com preview do problema no rodapé.

### [ ] `v0.6.0` — Busca Fuzzy em Larga Escala & Sessões Resilientes
- [ ] **Fuzzy Matcher de Alta Performance (`nucleo-matcher`):** Motor de busca fuzzy multi-thread e zero-allocation em cache, mantendo busca instantânea em monorepos com 100.000+ arquivos.
- [ ] **Sessões Resilientes & Auto-Save com Swap:** Arquivos de recuperação periódicos em `.oride/swap/` para restauração transparente após quedas de terminal ou reinicializações.
- [ ] **Jump List Histórica (`Ctrl+O` e `Ctrl+I`):** Pilha persistente de posições de navegação para retornar exatamente ao ponto anterior após navegações profundas.

### [ ] `v0.7.0` — Diffs Visuais & Ferramentas Git In-TUI
- [ ] **Visualizador de Diff Side-by-Side:** Split de duas colunas comparando HEAD vs Working Tree com realce sintático e scroll sincronizado.
- [ ] **Git Blame Virtual Text Sutil:** Informações virtuais esmaecidas no final da linha ativa (autor, commit e data relativa) ativáveis sob demanda.
- [ ] **Submódulos & Resolução Eficiente de `.gitignore`:** Tratamento transparente de múltiplos `.gitignore` aninhados e pastas de submódulos sem sobrecarga de I/O.

### [ ] `v0.8.0` — Ecossistema de Plugins & Sandboxing Seguro
- [ ] **LanguageProviders Dinâmicos via Lua:** Possibilidade de scripts registrarem suporte a novas extensões, delimitadores de comentários e comandos LSP sem recompilar o Oride em Rust.
- [ ] **Sandboxing de Permissões:** Políticas declarativas em `config.toml` restringindo acesso a processos externos (`allow_exec = false`) ou rede.
- [ ] **Hooks de Ciclo de Vida Expandidos:** Ganchos assíncronos protegidos para eventos de buffer (`on_change`, `on_cursor_move` com debounce).

### [ ] `v0.9.0` — Acessibilidade TUI, Suporte Unicode & Hardening
- [ ] **Acessibilidade TUI Completa:** Temas de alto contraste, pistas sonoras/assistivas opcionais e garantia de navegação 100% livre de dependência estrita de cores.
- [ ] **Alinhamento Unicode Preciso (CJK & Emojis):** Cálculo estrito de largura de glifos com `unicode-width`, eliminando desalinhamento de cursores e gutters ao editar caracteres asiáticos ou símbolos complexos.
- [ ] **Fuzzing de Buffers & Proteção Binária:** Blindagem contra abertura acidental de arquivos binários e testes de estresse contra entradas malformadas.

### [ ] `v1.0.0` — Versão de Longo Prazo (LTS) & Estabilidade de Produção
- [ ] **API Freeze:** Congelamento da API Lua com garantia de estabilidade e retrocompatibilidade na série 1.x.
- [ ] **Esquema de Configuração Estabilizado:** Schema versionado do `config.toml` com migração automática.
- [ ] **Suite de Benchmarks de Regressão (Criterion):**
  - Tempo de inicialização verificado: `< 10ms`
  - Latência do loop de eventos: `< 5ms`
  - Tamanho do executável compilado: `≤ 12MB`
- [ ] **Documentação Canônica de Produção:** Manuais completos, especificações de arquitetura e cheatsheets de bolso.
