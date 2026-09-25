# internal/fs

Árvore de projeto e ícones.

**Papel:** leitura preguiçosa do diretório de trabalho, e o mapa de glifo por
extensão.

**Ordem e filtro:** pastas antes de arquivos, ambos por nome; ocultos filtrados
conforme a config; `target` e `node_modules` descartados incondicionalmente.

**Enter revela em vez de alternar** — a referência alternava, e o resultado era
uma árvore que se fechava sozinha ao navegar.

**Fronteira de I/O:** `os.ReadDir`/`Stat`. É adaptador, não domínio.
