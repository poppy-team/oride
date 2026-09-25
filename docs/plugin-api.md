# Plugin API (P8)

Crate: **`oride-plugin`**. Há **dois** mecanismos distintos, e eles estão em
estados diferentes — esta seção existe para não confundir os dois.

| Mecanismo | Como a extensão é escrita | Estado |
|-----------|---------------------------|--------|
| **Embutido** | trait `Plugin` em Rust, compilado no binário | **ligado** — registrado em `PluginHost` no boot |
| **Externo** | `plugin.toml` declarando um executável | **implementado, não ligado** — nenhum caminho do app chama `discover_external_plugins` |

Nenhum dos dois carrega código de terceiros no processo: não há `Lua`, `WASM`
nem `dlopen`. Um plugin externo é um **executável** que o Oride invoca, o que
mantém o isolamento de processo — ver a nota de decisão abaixo.

Também não há, ainda, API estável versionada para terceiros.

## LanguageProvider

Metadados de linguagem (comentário, soft wrap, dica de LSP). O highlight
continua em `oride-syntax`.

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

Providers built-in: plain, oriscript, ori, markdown, mdx, html, css, javascript,
typescript/tsx, rust, python, nim e ruby.

Uso no app: `plugin_host.language(lang)` em toggle comment, soft wrap, servidor
LSP default e sugestões offline.

## Plugin + PluginCtx

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

Hooks: `OnOpen` (após defaults de linguagem ao abrir), `OnSave` (após save ok).

### Built-ins atuais

| Plugin | Comandos | Hooks |
|--------|----------|--------|
| `word-count` | **Plugin: word count** (palette) | — |
| `show-path` | **Plugin: show file path** (palette) | — |
| `lifecycle` | — | silencioso (pode anunciar em testes) |

## Command palette

`Ctrl+Shift+P` lista actions nativas **e** labels de plugins. Enter em
`Plugin: word count` executa o comando.

## Host

```rust
let host = oride_plugin::builtin_host();
host.palette_commands();
host.run_command("word_count", &mut ctx);
host.dispatch_hook(PluginHook::OnOpen, &mut ctx);
```

## Plugins externos (implementado, não ligado)

Um plugin externo é uma pasta com um `plugin.toml`:

```toml
[plugin]
name = "meu-plugin"
version = "0.1.0"
description = "exemplo"

[[commands]]
id = "saudacao"
label = "Plugin: saudação"
executable = "echo"
args = ["olá"]

[hooks.on_save]
executable = "echo"
args = ["salvo"]
```

Descoberta: `discover_external_plugins(&[dir])` procura em `dir/plugins/` e no
próprio `dir`, carregando cada `plugin.toml` encontrado. `ExternalPlugin::load_file`
faz o parse e guarda o diretório raiz do plugin.

**Estado real:** o crate `oride-plugin` reexporta essas duas funções, e o app não
as chama. Um plugin externo declarado hoje não executa nada — a capacidade existe
e não está ligada. Isto é uma divergência conhecida entre o código e o produto, e
não uma funcionalidade disponível.

O executável e seus argumentos são passados **separadamente**, sem shell
intermediário, para que um manifesto não possa injetar comandos.

## Decisão: por que executável, e não Lua/WASM

Um host de script dentro do processo daria a um plugin acesso à memória do editor.
Invocar um executável mantém o plugin isolado por processo, e é a mesma costura de
que a fase de cliente do harness precisa. `Lua`/`WASM` continuam fora de escopo.

Ver também: `docs/planning/post-0.1-roadmap.md` § P8.
