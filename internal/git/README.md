# internal/git

Estado do repositório.

**Papel:** responder o que a árvore e o painel SCM precisam mostrar.

**Decisão:** shell-out para `git status --porcelain -z`, sem libgit2. É o mesmo
caminho da referência Rust, e a saída por NUL não quebra com nome de arquivo que
contém espaço ou acento.

**Renomeação emite dois registros**, e o segundo — o caminho original — é
consumido, não exibido como arquivo separado.

**Fora de um repositório devolve mapa vazio**, não erro: não ser um repo é um
estado normal do editor.

**Fronteira de processo.** Nenhuma operação de escrita aqui pode rodar sem que a
camada acima tenha decidido — falha é valor, não exceção.
