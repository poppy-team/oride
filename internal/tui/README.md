# internal/tui

A camada de apresentação, em Bubble Tea v2.

**Papel:** transformar o estado de `internal/app` em tela, e transformar entrada em
mensagens. Nada mais. Nenhuma regra de edição, de layout de texto ou de protocolo
mora aqui — se uma decisão precisa existir quando não há terminal, ela pertence ao
modelo.

**O composition root é o `model.go`.** É o **único** arquivo deste pacote que pode
importar `internal/app`; as superfícies recebem o que precisam como dado. A regra
não é convenção: `R3-superficies-nao-importam-app` em `internal/architecture`
falha o build, e é o que faz uma superfície ser descartável — apagar a pasta e a
linha que a compõe, sem alcançar o modelo.

**Disciplina do runtime** (skill `bubbletea-v2`):
- `View()` é puro e barato: sem leitura de disco, sem subprocesso, sem rede. Ele
  roda a cada frame, e o que for lento pertence a um `Cmd`.
- `Cmd`/`Msg` é a **única** fronteira assíncrona. Uma goroutine alcança o programa
  por `Program.Send`, nunca mutando estado compartilhado.
- `Update` nunca bloqueia.
- Capacidade de terminal é **declarada no view** — alt screen, modo de mouse,
  keyboard enhancements, cursor. Alternar isso imperativamente a cada frame é como
  um terminal fica quebrado depois de um encerramento abrupto.

## Subpacotes

| Pacote | Papel | Importa `app`? |
|---|---|---|
| `layout` | geometria pura: onde cada superfície cai, e como o texto cabe nela | não |
| `component` | primitivas compartilhadas, quando a duplicação aparecer medida | não |
| `editor` `tree` `tabs` `statusbar` `menubar` `palette` `whichkey` `help` `modal` | uma superfície cada | não |

`component/` ainda **não existe**: o contrato §1.6 proíbe abstração especulativa, e
ele nasce na fatia em que a duplicação for medida, não antes.
