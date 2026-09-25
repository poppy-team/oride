# cmd/

Pontos de entrada executáveis. Cada subpasta é um binário do produto.

Um `main` aqui é **composition root**: monta as dependências e cede o controle.
Não contém regra de edição, de layout nem de protocolo. Se um binário precisar de
lógica, ela pertence a um pacote em `internal/`.

| Binário | Papel |
|---|---|
| `oride` | editor: abre a TUI, expõe `conformance run` como oráculo |
