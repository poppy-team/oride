# internal/config

Configuração efetiva, temas e `.editorconfig`.

**Papel:** resolver, em camadas, o que o editor deve usar: defaults → usuário
(`~/.config/oride/`) → projeto (`.oride/`). Depois os overlays.

**Resolve `.editorconfig` com a semântica de "mais próximo vence"** — a
referência Rust tinha o defeito inverso (ledger item B22).

**API para a view:** `Config.UI.*` (cores e largura de gutter) e
`Config.Syntax.*` (20 papéis de token) são dados puros; mapeá-los para o tipo de
estilo do Charm é trabalho da view, não daqui.

**Não importa `keymap` nem `action`** de propósito: `Config.Keys` é
`map[string]string` de ids crus, e `keymap.FromBindings` faz o parse. Isso evita
o ciclo `config ↔ keymap`.

**Gerado em parte.** `keybindings.go` (98 pares) e `themes.go` (10 temas) vêm de
`scripts/gen-tables.py`.
