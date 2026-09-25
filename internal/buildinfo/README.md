# internal/buildinfo

Identidade da build.

**Papel:** dar à versão, ao commit e à data um alvo estável para `-ldflags -X`,
para que o pipeline não precise conhecer nomes de variável espalhados.

**Puro e trivial:** só variáveis, sem dependências e sem inicialização.
