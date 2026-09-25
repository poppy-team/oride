# internal/architecture

O contrato de arquitetura como teste.

**Papel:** transformar as cláusulas de `docs/architecture/clean-code-contract.md`
§3 em regras que falham o build, em vez de prosa que decai.

**Regras** (R1–R7 em `Rules()`): ausência de ciclos, não importar o consumidor,
superfícies de `tui` não importarem o modelo, folhas puras permanecerem folhas,
nomes de pasta proibidos, README por pasta, e o harness não virar dependência de
produto.

**Desenho:** cada regra é um `Rule` com nome, justificativa e função de checagem.
O teste que as executa é um laço — acrescentar regra não significa escrever outro
teste.

**Contraprova:** `TestRulesCanFail` dá a cada regra um grafo que a viola (deve
acusar) e um que a satisfaz (deve silenciar). Uma regra que nunca falha não está
medindo nada.

**Só análise.** Lê imports com `go/parser` e o disco; não importa o produto, o que
mantém a checagem independente do código que ela julga.
