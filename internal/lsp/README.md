# internal/lsp

Cliente Language Server Protocol sobre stdio.

**Papel:** falar JSON-RPC com um servidor de linguagem durante a edição.

**Decisão que torna isto testável:** o **transporte é um par de streams,
separado do processo**. `NewWithTransport` permite dirigir o cliente inteiro por
um `io.Pipe`, sem servidor instalado; `Spawn` é só o caso em que o outro lado é
um processo. Doze testes exercitam o protocolo completo sem LSP na máquina.

**Escrita por fila limitada.** Escrever direto do chamador bloquearia enquanto o
servidor recusasse ler, e nenhum deadline de contexto interrompe uma escrita
bloqueada. A fila transforma isso em erro imediato (`ErrNotReading`).

**Conversão escalar ↔ UTF-16 recusa posição que parte um par substituto.** A
posição nomeia meio caractere; arredondar para qualquer lado moveria o cursor
para onde o servidor não quis dizer.

**Falha é valor.** Servidor ausente, lento ou quebrado vira linha de status.
