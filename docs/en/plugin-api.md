# Plugin Architecture & Extensibility API

Crate: **`oride-plugin`**. There are **two** distinct mechanisms, and they are in
different states — this section exists so the two are not confused.

| Mechanism | How the extension is written | State |
|-----------|------------------------------|-------|
| **Built-in** | a `Plugin` trait in Rust, compiled into the binary | **wired** — registered with `PluginHost` at boot |
| **External** | a `plugin.toml` declaring an executable | **implemented, not wired** — no application path calls `discover_external_plugins` |

Neither loads third-party code into the process: there is no `Lua`, `WASM` or
`dlopen`. An external plugin is an **executable** Oride invokes, which keeps the
process boundary intact.

There is also no stable versioned API for third parties yet.

---

## LanguageProvider Trait

Provides metadata per language (comment syntax, soft wrap preferences, default LSP command, and offline keyword suggestions). Highlighting queries remain encapsulated in `oride-syntax`.

```rust
pub trait LanguageProvider: Send + Sync {
    fn id(&self) -> &'static str;
    fn language_id(&self) -> LanguageId;
    fn extensions(&self) -> &'static [&'static str];
    fn comment_open(&self) -> Option<&'static str>;
    fn comment_close(&self) -> Option<&'static str>;
    fn lsp_command(&self) -> Option<&'static [&'static str]>;
    fn completion_words(&self) -> &'static [&'static str];
    fn default_soft_wrap(&self) -> bool;
}
```

**Built-in Providers:** `rust`, `c`, `bash`, `markdown`, `ori`, `python`, `javascript`, `typescript`/`tsx`, `html`, `css`, `nim`, `ruby`, `d`, `lua`, and `plain`.

**Usage:** `plugin_host.language(lang)` supplies data for comment toggling, automatic soft wrapping, fallback word completions, and default LSP process commands.

---

## Plugin Trait & PluginCtx

```rust
pub trait PluginCtx {
    fn set_status(&mut self, msg: &str);
    fn workspace_root(&self) -> &Path;
    fn active_path(&self) -> Option<PathBuf>;
    fn active_buffer_text(&self) -> String;
    fn active_is_dirty(&self) -> bool;
}

pub trait Plugin: Send + Sync {
    fn name(&self) -> &'static str;
    fn commands(&self) -> &'static [CommandMeta];
    fn on_hook(&self, hook: PluginHook, ctx: &mut dyn PluginCtx);
    fn run_command(&self, id: &str, ctx: &mut dyn PluginCtx) -> PluginResult;
}
```

### Lifecycle Hooks
- `OnOpen`: Dispatched immediately after a document buffer is loaded into memory.
- `OnSave`: Dispatched immediately after a buffer is successfully written to disk.

### Built-in Plugins

| Plugin | Commands | Hooks | Description |
|---|---|---|---|
| `word-count` | **Plugin: word count** | — | Counts words, characters, and lines in active buffer |
| `show-path` | **Plugin: show file path** | — | Displays the active document's path in status line |
| `lifecycle` | — | `OnOpen`, `OnSave` | Verifies lifecycle hook dispatch in test suites |

---

## Command Palette Integration

The Command Palette (`Ctrl+Shift+P`) dynamically presents built-in editor actions alongside registered plugin commands. Selecting a command executes `host.run_command(id, &mut ctx)`.

---

## External plugins (implemented, not wired)

An external plugin is a directory containing a `plugin.toml`:

```toml
[plugin]
name = "my-plugin"
version = "0.1.0"
description = "example"

[[commands]]
id = "greet"
label = "Plugin: greet"
executable = "echo"
args = ["hello"]

[hooks.on_save]
executable = "echo"
args = ["saved"]
```

Discovery: `discover_external_plugins(&[dir])` looks in `dir/plugins/` and in `dir`
itself, loading every `plugin.toml` it finds. `ExternalPlugin::load_file` parses it
and keeps the plugin's root directory.

**Actual state:** `oride-plugin` re-exports both functions and the application
never calls them. An external plugin declared today runs nothing — the capability
exists and is not wired. This is a known divergence between the code and the
product, not an available feature.

The executable and its arguments are passed **separately**, with no shell in
between, so a manifest cannot inject commands.

## Why an executable, and not Lua/WASM

An in-process script host would give a plugin access to the editor's memory.
Invoking an executable keeps the plugin isolated by process, and it is the same
seam the harness client phase needs. `Lua`/`WASM` remain out of scope.

## Verification

```bash
cargo test -p oride-plugin
```
