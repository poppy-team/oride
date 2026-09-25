# Prumo — Router de Documentação (Oride)

PRUMO é um router de intenção. Leia apenas o documento necessário para a tarefa
atual; não carregue a árvore de documentação inteira.

## Linha de implementação atual

- **Go** é a implementação alvo (`cmd/`, `internal/`) — em migração.
- **Rust** (`crates/`) está congelado como **oráculo diferencial** da migração.
  É arquivado em tag quando a paridade for atingida.
- Formatos canônicos: Markdown, JSON, TOML e Git.
- Estado derivado: caches, índices, contexto de execução, adaptadores compilados.

## Estado do projeto

- [Estado do projeto](../PROJECT_STATE.md) — fase ativa, meta ativa, próximas ações
- [`prumo.json`](../prumo.json) — manifesto canônico do projeto (Protocolo v3)
- [`AGENTS.md`](../AGENTS.md) — invariantes do produto e comandos de validação
- [Changelog](../CHANGELOG.md) — histórico de releases
- [Roadmap](../ROADMAP.md) — direção da fundação ao v1.0

## Migração Rust → Go

- [Ledger de paridade](migration/parity-ledger.md) — **toda** divergência entre o
  oráculo e a implementação Go, e toda contradição entre código e documentação.
  É a fonte única de exceções; um caso que divirja sem entrada aqui é regressão
- [Harness de conformance](../conformance/README.md) — formato do caso, as duas
  garantias que sustentam a paridade, e como rodar os dois lados
- Casos: `conformance/cases/<feature>/*.json`

**Regra:** paridade comportamental é comparada sem tolerância. Onde o ledger diz
*spec vence*, o oráculo está errado e o Go implementa a spec corrigida.

## Quero usar o produto

- [Guia de uso (Português)](guides/pt/guia-de-uso.md) — manual, modo modal Vim, splits, task runner
- [User Guide (English)](guides/en/user-guide.md)
- [Referência de configuração](guides/pt/config.md) — TOML, keymaps, mouse, terminal
- [Guia de temas](guides/pt/themes.md) — criação de temas com preview ao vivo
- [Sintaxe e linguagens](guides/pt/syntax.md) — detecção, tree-sitter, preview Markdown
- [Markdown](guides/pt/markdown.md) — extensões, pipeline de highlight, imagens

## Quero desenvolver

- [Design e arquitetura](design.md) — camadas, invariantes, fases
- [Plugin API](plugin-api.md) — manifesto, hooks e ferramentas externas
- [Polimento e releases](polish.md) — padrões de qualidade e histórico
- [Padrões de código](development/coding-standards.md)
- [Estratégia de testes](development/testing-strategy.md)
- [Arquitetura](architecture/overview.md) · [contrato de Clean Code](architecture/clean-code-contract.md) · [ADR 001](architecture/adr/001-architecture-baseline.md)
- [Governança de repositório](governance/repository-governance.md)

## Quero operar

- [Deploy e release](operations/deployment.md) · [Observabilidade](operations/observability.md)
- Validação: `cargo test --workspace` (oráculo) · `go test ./...` (implementação)
- Diagnóstico de ambiente e LSP: `:health` dentro do Oride

## Sou um agente

1. Leia `AGENTS.md`, `prumo.json` e `PROJECT_STATE.md` — e pare aí se bastar.
2. Use Lean Progressive Context: o menor contexto suficiente, expandindo só
   quando a evidência for insuficiente.
3. Prefira ponteiros (símbolo, seção, caminho) a despejar arquivos inteiros.
4. **O core não tem UI**: nunca misture lógica de terminal em `internal/buffer`,
   `editor`, `config` ou `keymap`. Esses pacotes são testáveis sem TTY.
5. Ao mudar comportamento, registre a divergência no ledger de paridade **e**
   adicione o caso de conformance que a reproduz — os dois no mesmo commit.
6. Mantenha documentação e código sincronizados; uma afirmação não verificada é
   uma dívida, não uma descrição.

## Decisões e specs

- [Design Document](design.md)
- [Especificação de configuração](guides/pt/config.md)
- [Arquitetura de plugins](plugin-api.md)
- [ADR 001 — linha de base arquitetural](architecture/adr/001-architecture-baseline.md)

## Metas

Metas vivem sob `.prumo/` e definem conclusão mensurável. A fase ativa está em
`prumo.json` (`goals.active_phase`).

## Inteligência durável

Inteligência compacta de projeto e tarefas vive em
`.prumo/history/project-intelligence.json`. Medições observadas e estimativas são
distinguidas, nunca somadas.
