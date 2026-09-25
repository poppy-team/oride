# internal/session

Persistência entre execuções.

**Papel:** guardar e restaurar o que o usuário espera reencontrar: abas abertas,
largura da árvore, visibilidade de painéis, scroll e split.

**Digest FNV-1a**, não `DefaultHasher`: o hash da referência Rust não era
estável entre versões, e o nome do arquivo de sessão mudava sozinho. Os três
digests golden são **idênticos aos do Rust**, byte a byte.

**Fronteira de I/O.** Um arquivo de sessão corrompido é descartado, não fatal —
perder o layout é aceitável, não abrir o editor não é.
