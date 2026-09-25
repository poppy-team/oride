# internal/app

O modelo. Estado e despacho, sem apresentação.

**Papel:** ser a **única** implementação do comportamento do editor. A TUI
renderiza daqui; o harness de conformidade compara daqui. Não existem dois
caminhos, então uma correção aparece nos dois.

**Contrato com quem consome**
- `New(...)` monta o estado; `ApplyKey`/`ApplyText`/`Apply`/`ApplySearch` mudam.
- `DumpState()` é o estado observável serializado — `SchemaVersion` **congelado
  em 1**, porque é comparado byte a byte contra o oráculo Rust.
- Nenhuma decisão de cor, geometria ou widget mora aqui.

**Não importa `internal/tui` nem `conformance`** (R2). O core desconhece a camada
visual — `clean-code-contract` §2.

**O que a TUI exige e ainda não existe aqui:** overlay e split como estado real,
viewport/scroll, tamanho de janela e mouse. Ver o plano de M3.
