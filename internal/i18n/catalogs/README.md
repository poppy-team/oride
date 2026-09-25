# internal/i18n/catalogs

Os catálogos embutidos.

Um arquivo TOML por locale, compilado no binário por `embed`. O catálogo padrão
é a raiz da herança de fallback: todo campo que outro locale não declarar é
resolvido aqui.

Este arquivo e o do lado Rust apontam para o **mesmo** conteúdo — não duplique
um catálogo aqui e no oráculo.
