# internal/i18n

Catálogos de mensagens.

**Papel:** dar rótulo traduzido a menu, palette e status, sem que a view conheça
strings soltas.

**Herança de fallback:** um campo ausente no catálogo corrente cai no catálogo
padrão, em vez de aparecer vazio. É a melhoria B23 do ledger.

**Fontes:** catálogos embutidos (`catalogs/`, via `embed`) e catálogos em disco
que os substituem. Os caminhos embutidos usam `path.Join`, nunca
`filepath.Join` — o separador do Windows não existe dentro de um FS embutido.

**Trocado entre PT e EN, o arquivo é o mesmo** que o Rust lê, por `include_str!`
nos dois lados: duas cópias divergiriam.
