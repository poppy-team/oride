# internal/search

Busca no buffer e no projeto.

**Duas metades, com purezas diferentes:**
- `find.go` — o matcher **puro** sobre texto: dobra de caixa e de acento,
  fronteira de palavra, regex, offsets em **bytes no texto original**.
- `project.go` — a busca no projeto, que é fronteira de I/O.

**Provider preferido: `rg --json`**, não `path:line:text` — um nome de arquivo
pode conter dois-pontos e a divisão ingênua reporta caminho sem sentido. O
fallback em Go cobre a máquina sem ripgrep.

**Os dois backends não ignoram os mesmos arquivos** (ledger B25): `rg` pula só o
que o git ignora, e só honra um `.gitignore` dentro de um repositório; o fallback
pula saída de build incondicionalmente. Por isso a busca **devolve qual backend
respondeu** — um resultado ausente precisa de explicação.
