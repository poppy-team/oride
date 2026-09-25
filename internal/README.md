# internal/

O produto. Nenhum pacote daqui pode ser importado de fora do módulo.

## Direção das dependências

```
cmd/oride → internal/tui → internal/app → domínio puro → adaptadores
```

Aponta sempre **para dentro**, na direção das regras de edição
(`docs/architecture/clean-code-contract.md` §2). `app` nunca importa `tui`; as
superfícies de `tui` nunca importam `app` — só o composition root monta as views.

## Camadas

| Camada | Pacotes | Natureza |
|---|---|---|
| Domínio puro | `buffer`, `action`, `editor`, `keymap`(parcial), `search`(matcher), `term`(chunk) | sem I/O, sem TTY |
| Modelo | `app` | estado + despacho; sem apresentação |
| Apresentação | `tui` | renderiza e emite mensagens |
| Adaptadores | `fs`, `git`, `lsp`, `plugin`, `session`, `term`, `osutil` | fronteiras de processo e disco |
| Contratos executáveis | `architecture` | regras do contrato como teste |

## Regras

As cláusulas de `docs/architecture/clean-code-contract.md` §3 são verificadas por
`internal/architecture`, não por convenção: ciclos, importação do consumidor,
nomes proibidos e README por pasta falham o build.

Toda pasta aqui tem o seu próprio `README.md` — é o §3 do contrato, e é checado.
