# User Guide — Oride

Welcome to **Oride** (Ori + IDE), a lightweight, modular, and extensible terminal code editor and mini-IDE built in Rust. This practical guide covers everything from basic navigation and editing to advanced Vim-style modal editing, the declarative task runner, dynamic splits, and environment health diagnostics.

---

## 1. Getting Started & Installation

### Building and Installation
Oride is built with the standard Rust toolchain (Cargo):

```bash
# Build an optimized release binary
cargo build --release

# Install locally to ~/.local/bin/oride
./scripts/install.sh
```

### Launching the Editor
You can launch Oride from your terminal in several ways:

```bash
# Open the current working directory as a workspace
oride .

# Open a specific file
oride src/main.rs

# Open a project folder directly
oride /path/to/project

# Check the installed version
oride --version
```

---

## 2. Interface Layout

![Oride Interface](../../assets/oride-interface.png)

Oride's user interface is crafted for optimal terminal ergonomics using the Ratatui library:

- **Top Menu Bar:** Accessible via standard mnemonics: `Alt+F` (File), `Alt+E` (Edit), `Alt+V` (View), `Alt+G` (Go), `Alt+I` (Git), and `Alt+H` (Help).
- **Buffer Tabs:** Displays all open documents. The active buffer is highlighted, and unsaved changes are marked with an asterisk (`*`).
- **Project File Tree:** Left panel providing recursive file browsing, directory collapse/expansion, Nerd Font icons, and Git status badges (`M`, `A`, `?`).
- **Editor Workspace:** Supports line numbers (gutter), multi-cursor editing, soft word wrapping (`Alt+Z`), and Tree-Sitter syntax highlighting.
- **Git / SCM Panel:** Accessible on the right via `Ctrl+Shift+G`, showing modified files across your working tree.
- **Embedded Terminal (PTY):** A collapsible, full-featured interactive shell at the bottom (`Ctrl+\`` or `Ctrl+"`).
- **Status Line:** Displays caret position (line, column), file encoding, detected language, active editing mode, and system feedback.

---

## 3. Editing Paradigms

Oride seamlessly integrates two editing paradigms:

### Standard Mode (Micro Style)
By default, Oride operates as an intuitive modern text editor:
- Arrow keys for navigation, `Home`/`End`, `PageUp`/`PageDown`.
- Selection via `Shift + Arrow keys` or mouse drag (when `mouse = true` in config).
- Universal keybindings: `Ctrl+S` (save), `Ctrl+C` (copy), `Ctrl+V` (paste), `Ctrl+Z` (undo), and `Ctrl+Y` (redo).

### Modal Mode (Vim Style)
For developers who prefer home-row navigation without leaving the keyboard:

1. **Activation:** Set `modal_mode = true` in your `config.toml`, or open the Command Palette (`Ctrl+Shift+P`) and choose **Toggle modal mode (Vim / CUA)** to switch without leaving the editor.
2. **Available Modes:**
   - **NORMAL:** Default mode for navigation, deletions, and commands. Cursor is drawn as a solid block.
   - **INSERT (`i`, `a`, `o`, `I`, `A`, `O`):** Direct text insertion. Press `Esc` to return to NORMAL mode.
   - **VISUAL (`v`):** Character-wise visual text selection.
   - **VISUAL LINE (`V`):** Line-wise visual text selection.
   - **COMMAND (`:`):** Command-line prompt in the status bar for editor actions.

3. **Essential Normal Mode Motions:**
   - `h`, `j`, `k`, `l`: Left, down, up, right.
   - `w`, `b`: Next word beginning, previous word beginning.
   - `0`, `$`: Line start, line end.
   - `gg`, `G`: First line of buffer, last line of buffer.
   - `x`: Delete character under cursor.
   - `u`: Undo change.

4. **Command-Line Commands (`:`):**
   - `:w` — Save current buffer.
   - `:q` — Close active buffer or quit if last buffer.
   - `:q!` — Force quit without saving unwritten changes.
   - `:wq` — Save and quit.
   - `:e <path>` — Open a file into a buffer.
   - `:tasks` — Open the integrated Task Runner picker.
   - `:health` — Open the system and LSP health diagnostics modal.
   - `:theme <name>` — Switch color theme (e.g., `:theme tokyo-night`).
   - `:lang <code>` — Switch the interface language (e.g., `:lang en-US`).
   - `:noh` — Clear the current search selection.
   - `:run <task>` — Run a task runner entry by name.

   To see every binding, use `F1`, `Ctrl+G` or `Ctrl+Shift+/` — there is no `:help`.

---

## 4. Window Splits

Oride supports dynamic split panes for viewing and editing multiple files side-by-side:

- **Vertical Split:** `Ctrl+Alt+V` (splits pane vertically).
- **Horizontal Split:** `Ctrl+Alt+H` (splits pane horizontally).
- **Cycle Active Pane:** `F6` (moves focus to next pane).
- **Close Active Pane:** `Ctrl+Alt+W`.

---

## 5. Search and Replace

### Within the Active Buffer
- `Ctrl+F`: Opens the search mini-modal at the bottom.
- `F3`: Jumps to the next match.
- `Ctrl+H`: Opens the Find and Replace dialog (press `Tab` to switch between Find and Replace fields).
- **Search Toggles:**
  - `Alt+C`: Toggle Case Sensitivity.
  - `Alt+A`: Toggle Accent Insensitivity (e.g., `funcao` matches `função`).
  - `Alt+W`: Toggle Whole Word matching.
  - `Alt+R`: Toggle **Regular Expression (Regex)** search.
- `Alt+Enter`: Replace current match.
- `Ctrl+Alt+Enter`: Replace all occurrences.

### Across the Entire Project
- `Ctrl+Shift+F`: Opens project-wide search. Leverages system `ripgrep` (`rg`) when available or falls back to an internal Rust engine. Supports glob pattern filtering (e.g. `*.rs, !target/*`).

---

## 6. Integrated Task Runner

Automate builds, tests, and formatting tasks directly within the editor via declarative tasks.

### Setting Up Tasks
Create a `tasks.toml` file in your project root or in `.oride/tasks.toml`:

```toml
[tasks.build]
label = "Cargo: Build"
command = "cargo build"
description = "Compile the workspace in debug mode"

[tasks.test]
label = "Cargo: Test Active Module"
command = "cargo test $FILE_NAME"
description = "Run tests for currently open module"

[tasks.run]
label = "Run Binary"
command = "cargo run"

[tasks.format]
label = "Format Code"
command = "cargo fmt --all"
```

### Magic Variables:
- `$FILE`: Full absolute path of active file.
- `$FILE_NAME`: File name with extension (e.g., `main.rs`).
- `$FILE_STEM`: File name without extension (e.g., `main`).
- `$FILE_DIR`: Parent directory of active file.
- `$WORKSPACE`: Root folder of open workspace.
- `$LINE`: Current cursor line (1-based).
- `$COL`: Current cursor column (1-based).

### Running Tasks:
Type `:tasks` in the command prompt or open the Command Palette (`Ctrl+Shift+P`) and choose *Run Task*.

---

## 7. System & LSP Health Diagnostics (`:health`)

To ensure your environment has all necessary CLI tools and language servers configured:

1. Type `:health` in the command prompt or select **Help → Health Check**.
2. The diagnostic modal checks:
   - System executables in `$PATH` (`git`, `rg`, `cargo`, compilers).
   - Connectivity and status of configured Language Servers (`rust-analyzer`, `clangd`, `ori-lsp`, etc.).
   - Terminal graphics protocol support and font rendering metrics.

---

## 8. In-Terminal Markdown Preview

Oride provides a rich Markdown preview rendered right inside your terminal:

- **Shortcut:** Press `Ctrl+Shift+V` or `Alt+P` on any `.md` file.
- **Capabilities:**
  - Unicode box-drawing tables with column alignments.
  - Syntax-highlighted code blocks (fenced code injections).
  - Interactive checklists (`[ ]` and `[x]`).
  - Terminal image rendering on supported terminals (Kitty, Ghostty, WezTerm) via `terminal_images = true` in config.

---

## 9. Themes and Visual Customization

Oride includes over 10 built-in themes (`tokyo-night`, `dracula`, `nord`, `one-dark`, `catppuccin-mocha`, `monokai`, `github-dark`, `solarized-dark`, `solarized-light`, etc.) and supports external themes without recompilation.

- **Switching Themes:** Open the Command Palette (`Ctrl+Shift+P`), type *Theme*, and scroll to see **Live Preview** in real time.
- **Creating Custom Themes:**
  Save your theme in `~/.config/oride/themes/my-theme.toml`. Oride discovers it automatically. See the [Theme Development Guide](themes.md) for theme structure details.

---

## 10. Configuration & Internationalization

### Configuration Layers
Configuration merges cleanly across three layers:
1. Built-in defaults.
2. User configuration: `~/.config/oride/config.toml`.
3. Workspace configuration: `.oride/config.toml`.

### Language and Locales
Oride supports dynamic UI localization. In your `config.toml`:

```toml
locale = "en-US" # or "pt-BR"
```

Custom translations can be added to `~/.config/oride/locales/<locale>.toml` without recompilation.
