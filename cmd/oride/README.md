# cmd/oride

O binário do editor.

**Papel:** interpretar a linha de comando, montar o modelo (`internal/app`) e a
view (`internal/tui`), e entregar o controle ao runtime.

**Responsabilidades**
- `--version` / `--help`, com a identidade vinda de `internal/buildinfo`.
- Subcomando `conformance run --case F --workspace D [--json]`: o **oráculo**
  congelado, que o harness diferencial executa como subprocesso. stdout é
  reservado ao JSON; diagnósticos vão para stderr; sem ANSI.
- Abrir a TUI quando não há subcomando.

**Não faz:** edição, layout, despacho de ação. Nada aqui deve precisar de teste
de comportamento — se precisar, a lógica está no lugar errado.
