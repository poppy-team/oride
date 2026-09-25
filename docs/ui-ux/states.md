# Vocabulário de estados

Os nomes que o estado observável usa. O dump de conformidade
(`internal/app`, `SchemaVersion` 1) é comparado byte a byte contra o oráculo Rust,
então um nome novo aqui é uma mudança de contrato entre as duas implementações.

## Foco

`editor` · `tree` · `terminal` · `scm` — ver `docs/ui-ux/focus-graph.md`.

## Sobreposição

`none` quando não há nenhuma. As demais variantes são nomeadas pela superfície:
`command_palette`, `help`, `find`, `project_find`, `diagnostics`, `completion`,
`hover`, `buffer_picker`, `diff`, `theme_picker`, `locale_picker`,
`health_check`, entre outras.

**Só estados determinísticos entram no dump.** Um estado cujo conteúdo depende do
ambiente aparece como contagem ou booleano — `hover` conta linhas, `project_find`
conta ocorrências, `diff` conta linhas. Um dump que carregasse o conteúdo real
seria uma comparação que falha por motivo alheio ao produto.

## Status

- vazio: sem mensagem;
- texto: a mensagem transitória corrente.

Mensagem transitória **expira por tempo** e não entra no dump — o dump é
determinístico, e um relógio não é.

## Degradação

| Estado | Significa |
|---|---|
| `terminal_attached` | há um PTY ligado |
| `diagnostics` | quantos diagnósticos o servidor publicou |
| `lsp_failures` | falhas de servidor, ordenadas |
| `vim` | o modo modal, quando ligado |

Fronteiras de processo aparecem **só assim**: nunca como conteúdo, sempre como
contagem, booleano ou lista ordenada.
