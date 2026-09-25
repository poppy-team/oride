# Official Oride Roadmap — From Foundation to Version 1.0 (LTS)

**Status:** Normative  
**Current Version:** **`0.2.0`**  
**Philosophy:** Contained, lightweight, fail-closed TUI mini-IDE with instant boot and strict anti-bloat discipline.  
**Performance Targets:** Startup time `< 10ms`, input/render latency `< 5ms (120 FPS)`, idle RAM `< 25MB`, stripped binary size `<= 12MB`.

---

## 🎯 Core Principles (Anti-Bloat & Performance)

1. **Contained in the Terminal (TUI):** No external browser preview dependencies, no heavy webview processes, no bloated GUI wrappers. Everything runs directly in the user's terminal emulator via Ratatui and Crossterm.
2. **One Concept per PR:** Small, focused, TDD-verified pull requests documented within the same slice.
3. **Fail-Closed Resilience:** Failures in LSP, Git, or PTY report clear, non-intrusive status line diagnostics and never panic or crash the editor.
4. **Binary Budget Discipline:** Zero gratuitous heavy dependencies. Grammars and runtimes are optimized and stripped upon release.
5. **Zero-Copy & Streaming:** The UI renders only visible viewport rows without redundant dynamic string allocations.

---

## 📦 Shipped Releases

### [x] `v0.1.0` — Contained TUI Mini-IDE Baseline
- [x] **Editor Core:** High-performance Ropey buffers, buffer tabs, grouped undo/redo history, selections, and comments.
- [x] **Project Navigation & Tree:** Navigable file tree, folder/file creation/renaming/deletion, Nerd Font icons with clean ASCII fallback.
- [x] **Embedded Terminal:** Interactive real PTY (`portable-pty`), resizing, and focus switching.
- [x] **Integrated Search:** In-buffer search (case, accents, regex) and project-wide search via ripgrep/internal crawler.
- [x] **Basic Git Integration:** SCM status, visual badges on project tree, read-only diff, and blame on status line.
- [x] **Markdown Preview:** Terminal rendering of tables, headers, task lists, and syntax-highlighted code fences.
- [x] **Multi-Server LSP:** On-demand stdio JSON-RPC client for autocompletion, hover, and goto definition.
- [x] **Layout & Splits:** Horizontal and vertical editor splits, multi-cursor editing, and opt-in mouse support (`mouse = false` by default).

### [x] `v0.2.0` — First-Class Languages, Media, Git Sync & Tooling
- [x] **First-Class Languages:** Native support for Rust, C, Bash, Markdown, Ori, Python, TypeScript/TSX, HTML, CSS, D, Lua, Nim, and Ruby via Tree-Sitter or contained lexical engines.
- [x] **Terminal Markdown Media:** Support for terminal graphics protocols (Kitty Graphics, Sixel, iTerm2) with image preview cards and system browser link opening.
- [x] **Git Sync & CLI Staging:** Staging (`s`), unstage (`u`), commit prompt (`c`), push (`P`), pull (`p`), and ahead/behind branch counter in status bar.
- [x] **Session Persistence:** Transparent restoration of scroll offsets, splits, active documents, and tree width via `.oride/session.toml`.
- [x] **Project Search & Replace with Globs:** Full project replace (`Ctrl+Shift+F`) with glob filtering (`*.rs`, `!target/**`).
- [x] **Vim-Style Modal Editing:** `Normal`, `Insert`, `Visual`, `VisualLine`, and command-line `:` modes.
- [x] **Integrated Task Runner:** Declarative execution of scripts in `tasks.toml` with variable interpolation (`$FILE`, `$DIR`, `$WORKSPACE`).
- [x] **Environment Diagnostics (`:health`):** Proactive verification of tools, compilers, and LSP servers in `$PATH`.
- [x] **Dynamic Internationalization (i18n):** External TOML language catalogs (`pt-BR`, `en-US`) and custom theme documentation.
- [x] **Multi-Distro Packaging & Universal Installers:** Packages for Arch Linux (`PKGBUILD`), Debian/Ubuntu (`.deb`), Fedora (`.spec`), Void, and Nix (`flake.nix`), with one-line shell installers for Linux, macOS, and Windows.

---

## 🚀 Planned Slices (Road to Version 1.0)

### [ ] `v0.3.0` — Lightweight Extensibility & Editing Ergonomics
- [ ] **Lua Plugin Engine (`mlua`):** Embedded Lua 5.4 runtime allowing user scripts in `~/.config/oride/plugins/*.lua` with access to palette commands, buffer context, and lifecycle hooks (`on_open`, `on_save`).
- [ ] **Smart Delimiter Auto-Pairing:** Paired insertion of `()`, `[]`, `{}`, `""`, `''`, and ```` `` ````, with skip-over upon typing closing characters and automatic selection wrapping.
- [ ] **Modal Text Objects:** Scoped Vim-style motions and text objects (`ci"`, `da(`, `yiw`, `vi{`).
- [ ] **PTY Scrollback History:** Circular history buffer for the embedded terminal panel with keyboard and mouse scrolling.
- [ ] **LSP Buffer Symbol Outline:** Outline picker for functions, types, and variables in the active document (`Ctrl+Shift+O`).
- [ ] **Self-Update Command (`oride --update` / `:update`):** Automated check against GitHub Releases and in-place binary upgrade when installed via the universal installer or GitHub.

### [ ] `v0.4.0` — Extreme Performance & Large File Management
- [ ] **Large File Mode (> 10MB / > 100MB):** Automatic detection of oversized files with fallback to Ropey chunked streaming view, bypassing heavy Tree-Sitter and soft word wrapping to prevent UI stutter.
- [ ] **Zero-Copy Viewport Rendering:** Direct row slice (`&str`) rendering to Ratatui frames, eliminating per-line intermediate `String` allocations.
- [ ] **Lazy Grammar Initialization:** On-demand compilation of Tree-Sitter queries and parsers on the first file opened for each language (boot time `< 5ms`).
- [ ] **Interactive Git Hunk Staging:** Review and stage/unstage individual diff hunks directly within the SCM interface.

### [ ] `v0.5.0` — Advanced Code Intelligence & Contained Refactoring
- [ ] **LSP Code Actions & QuickFix (`Alt+Enter`):** Interactive prompt to apply compiler/linter suggestions (missing imports, lifetime fixes, lints).
- [ ] **Project-Wide Symbol Rename (`F2`):** Semantic renaming propagated cleanly via LSP `workspace/applyEdit` across open and disk buffers.
- [ ] **Lightweight Snippets with Tabstops:** LSP completion snippet expansion with `$1`, `$2`, `$0` tab navigation.
- [ ] **Direct Diagnostics Navigation (`Alt+N` / `Alt+P`):** Fast jumps to next/previous error or warning with preview in the status line.

### [ ] `v0.6.0` — Large-Scale Fuzzy Matching & Resilient Sessions
- [ ] **High-Performance Fuzzy Matcher (`nucleo-matcher`):** Multi-threaded, zero-allocation fuzzy matching engine maintaining sub-millisecond search across monorepos with 100,000+ files.
- [ ] **Resilient Sessions & Auto-Save Swap:** Periodic recovery state in `.oride/swap/` for clean crash recovery after unexpected terminal disconnects.
- [ ] **Historical Jump List (`Ctrl+O` & `Ctrl+I`):** Persistent jump stack allowing bidirectional return along navigation paths.

### [ ] `v0.7.0` — Visual Diffs & In-TUI Git Tooling
- [ ] **Side-by-Side Visual Diff Viewer:** Two-column synchronized split comparing HEAD vs Working Tree with syntax highlighting.
- [ ] **Subtle Git Blame Virtual Text:** Faded inline blame annotation at the end of the active line (author, commit, relative time), togglable on demand.
- [ ] **Submodules & Nested `.gitignore` Handling:** Efficient traversal of complex repository structures without I/O degradation.

### [ ] `v0.8.0` — Plugin Ecosystem & Secure Sandboxing
- [ ] **Dynamic LanguageProviders via Lua:** Allow Lua scripts to register new file extensions, comment delimiters, and LSP servers without recompiling Oride in Rust.
- [ ] **Declarative Permission Sandboxing:** Configuration flags in `config.toml` restricting external process execution (`allow_exec = false`) or network access.
- [ ] **Expanded Lifecycle Hooks:** Protected asynchronous hooks for buffer events (`on_change`, `on_cursor_move` with debouncing).

### [ ] `v0.9.0` — TUI Accessibility, Unicode Support & Hardening
- [ ] **Full TUI Accessibility:** High-contrast color palettes, optional terminal bell/audio cues, and full keyboard-only workflows independent of color indicators.
- [ ] **Accurate Unicode Alignment (CJK & Emojis):** Strict character width calculations using `unicode-width`, eliminating ghost cursors and gutter misalignment in multi-byte text.
- [ ] **Buffer Fuzzing & Binary Protection:** Safe guards against accidental binary file opening and fuzz-tested protection against malformed UTF-8 inputs.

### [ ] `v1.0.0` — Long-Term Support (LTS) & Production Stability
- [ ] **API Freeze:** Formal stabilization and backwards-compatibility guarantees for the Lua plugin API throughout the 1.x series.
- [ ] **Stabilized Configuration Schema:** Versioned `config.toml` schema with automatic forward migration.
- [ ] **Automated Regression Benchmark Suite (Criterion):**
  - Verified boot time: `< 10ms`
  - Event loop latency: `< 5ms`
  - Compiled release binary: `≤ 12MB`
- [ ] **Canonical Production Documentation:** Comprehensive user manuals, architecture documentation, and printable cheat sheets.
