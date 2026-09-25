# internal/action

A tabela canônica de ações.

**Papel:** dar identidade estável a cada operação que o editor sabe executar.

**Decisão central:** `type Action string` — **a string é a identidade**. Um id de
keybinding na config, uma entrada da palette e uma asserção de conformidade
nomeiam o mesmo objeto por construção, sem tabela de tradução para manter.

**API:** `Parse(string) (Action, error)`, `Action.Valid()`, `Action.String()`,
`All()`, `Palette()`. Alias map para nomes históricos (`checkhealth` →
`health_check`).

**Puro.** Não importa nenhum pacote interno (verificado por R4 em
`internal/architecture`): é folha por invariante, não por acaso.

**Gerado.** `action.go` vem de `scripts/gen-tables.py`, que lê o `Action` do Rust
e valida os ids contra esta tabela. Não edite à mão — o `--check` do gerador
falha o CI.
