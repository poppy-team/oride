# internal/plugin

Extensão por manifesto.

**Papel:** descobrir plugins externos declarados em `plugin.toml` e executá-los.

**Isolamento por processo:** um plugin é um **executável**. Não há Lua, WASM nem
`dlopen` — um host de script dentro do processo daria ao plugin acesso à memória
do editor.

**Executável e argumentos são passados separadamente**, nunca por shell: um
manifesto não pode injetar comando. Há teste com `;echo INJETADO` que passa.

**Falha fechada:** executável ausente falha com status; manifesto malformado é
ignorado em vez de fatal; descoberta ordenada por nome para ser determinística.
