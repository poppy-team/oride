# internal/docdrift

Contratos de documentação, verificados contra o código.

**Papel:** impedir que um documento descreva um código que não existe mais.

**Por que existe:** a tabela de teclas copiada para um guia está correta no dia em
que é escrita e errada na primeira vez que uma binding muda. Em vez de repetir a
tabela, o documento passa a ser **derivado**: o teste regenera com `-update` e
compara nos dois sentidos sem ele.

**Como funciona:** o conteúdo gerado vive entre `<!-- GERADO: keymap -->` e
`<!-- FIM: keymap -->`. Fora dos marcadores é prosa, escrita à mão. Um documento
sem os marcadores é recusado — acrescentar em vez de substituir faria a checagem
passar enquanto o arquivo ganhava uma tabela duplicada.

**Regenerar:**
```bash
go test ./internal/docdrift/ -update
```

**Direção do fluxo:** este pacote importa `config` e `keymap`; nenhum dos dois o
importa. É checagem, não dependência de produto.

**Alcance:** além do keymap, é onde entram as checagens do grafo de foco, do
vocabulário de estados e das regras de degradação — todos contratos entre um
documento e o código.
