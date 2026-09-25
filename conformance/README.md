# Conformance

Harness diferencial de paridade entre o **oráculo Rust** (`crates/`) e a
**implementação Go** (`internal/`), prevista no plano de migração.

O mesmo caso é executado nas duas implementações e o estado observável é
comparado sem tolerância. Enquanto o Rust existir, ele é a referência
comportamental — exceto onde o ledger de paridade diz o contrário.

## Rodando

O lado Rust precisa de um modo de entrada scriptada — ele não pode ser linkado
dentro de um teste Go — então o oráculo se dirige por subprocesso:

```bash
cargo build -p oride          # o oráculo
mkdir -p /tmp/oride-conf-ws
cargo run -p oride -- conformance run \
  --case conformance/cases/editor/typing-undo-groups.json \
  --workspace /tmp/oride-conf-ws \
  --json
```

Com `--json` o stdout é reservado ao JSON e todo diagnóstico vai para o stderr,
sem ANSI — a mesma convenção de saída de máquina do Prumo.

O lado Go é uma suíte de testes, porque pode rodar a implementação no processo:

```bash
go test ./conformance/...
```

**Armadilha:** `cargo test` não reescreve o binário em `target/debug`. Se você
editar o oráculo e só rodar `cargo test`, `go test ./conformance/...` compara a
implementação Go contra a revisão Rust anterior. Rode `cargo build -p oride`
antes — é o que o job `conformance` do CI faz.

## As duas garantias que sustentam o resto

**O oráculo é determinístico.** `TestOracleIsDeterministic` roda cada caso duas
vezes, em diretórios diferentes, e exige relatórios idênticos. Paridade
diferencial não significa nada se a referência responde coisas diferentes em
execuções diferentes. Foi esse teste que pegou o vazamento do basename do
workspace no nome da raiz da árvore — o único valor do dump que vinha do
ambiente em vez do produto.

**O comparador detecta divergência.** `TestComparisonDetectsDivergence` muta um
relatório e exige que a comparação perceba. Um comparador que sempre responde
"igual" faria a paridade subir a 100% sozinha — o único modo de falha deste
harness que se pareceria com sucesso.

Um comparador que não consegue comparar — schema diferente, contagem de quadros
diferente — devolve erro, não uma falha de paridade. Uma suíte quebrada não pode
se disfarçar de progresso.

## Formato do caso

```json
{
  "description": "o que este caso guarda, em uma frase",
  "files": [{ "path": "relativo/ao/workspace.txt", "content": "..." }],
  "open": "relativo/ao/workspace.txt",
  "config": { "mouse": false, "keys": { "ctrl+k": "help" } },
  "git": false,
  "steps": [
    { "kind": "action", "action": "move_doc_end" },
    { "kind": "text", "text": "abc" },
    { "kind": "chord", "chord": "ctrl+s" }
  ]
}
```

| Campo | Obrigatório | Efeito |
|---|---|---|
| `files` | não | Materializa arquivos antes de rodar. Caminhos relativos ao workspace; `..` que escape da raiz é erro |
| `open` | não | Arquivo aberto ao iniciar. Ausente ⇒ buffer vazio |
| `config` | não | Overrides sobre `Config::default()`. Campo ausente mantém o default do produto |
| `config.keys` | não | Overrides de `[keys]` (chord → id de ação), aplicados sobre os bindings default |
| `git` | não | Habilita consultas a git. Desligado por padrão: um caso que não testa git não deve depender de o binário existir |
| `steps` | **sim** | Sequência de passos. O estado é capturado após **cada** passo |

### Passos

| `kind` | Campos | Semântica |
|---|---|---|
| `action` | `action` | Dispara pelo id estável (`move_left_extend`), sem passar pelo keymap |
| `chord` | `chord` | Resolve pelo keymap efetivo. **Falha se o chord não tiver binding** — é assim que o caso detecta drift de keymap |
| `text` | `text` | Digita caractere a caractere, como o usuário |

## Por que o dump é determinístico

O modo conformance constrói o app sem PTY, sem watcher de disco, sem LSP, sem
plugins externos e sem ler a config do usuário. Um caso depende apenas dos
arquivos que ele mesmo declara. Caminhos são relativos ao workspace
(`<workspace>/...`) para que a fixture não dependa de onde o caso rodou.

Consequência deliberada: recursos que são fronteira de processo ou de ambiente
— PTY, LSP, `:health` — aparecem no dump apenas como contagem ou booleano
(`terminal_attached`, `diagnostics`, `hover.lines`). Eles são cobertos por testes
com servidor stub, não pelo dump de estado.

## Versionamento

`DUMP_SCHEMA_VERSION` versiona o shape do dump. Mudar qualquer campo exige
incrementar a constante: uma fixture antiga precisa falhar de forma explícita em
vez de comparar campos que já não existem.

`Action::ALL` e `Action::id()` em `crates/oride-keymap/src/action.rs` são a
tabela canônica de ações do produto. O teste
`id_roundtrips_through_parse_action` garante que id e variante não divergem, e é
essa tabela que a implementação Go copia.

## Organização dos casos

```
conformance/cases/
  editor/    buffer, seleção, multi-cursor, undo/redo
  keymap/    resolução de chord, rebinding, aliases
  find/      busca no buffer (literal, regex, acentos, palavra)
  tree/      árvore de projeto, expand/collapse, ordem
  tabs/      abrir/fechar/alternar
  config/    merge, .editorconfig, temas
  git/       status, SCM
```

O nome do arquivo descreve o invariante, não a feature: um caso que falha deve
dizer pelo nome o que quebrou.
