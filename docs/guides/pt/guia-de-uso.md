# Manual do Usuário — Oride

Bem-vindo ao **Oride** (Ori + IDE), um editor e mini-IDE de terminal leve, modular e extensível construído em Rust. Este guia prático ensina a utilizar todos os recursos do editor, desde as operações básicas até os fluxos avançados de edição modal, task runner e diagnósticos de ambiente.

---

## 1. Primeiros Passos & Instalação

### Compilação e Instalação
O Oride é compilado com o ecossistema padrão do Rust (Cargo):

```bash
# Compilar versão de release otimizada
cargo build --release

# Instalar no sistema (~/.local/bin/oride)
./scripts/install.sh
```

### Inicializando o Editor
Você pode abrir o Oride em diferentes modos a partir da linha de comando:

```bash
# Abrir no diretório atual (define o CWD como raiz do workspace)
oride .

# Abrir um arquivo específico
oride src/main.rs

# Abrir um arquivo em uma pasta qualquer como workspace
oride /caminho/para/projeto/arquivo.c

# Verificar a versão instalada
oride --version
```

---

## 2. Layout da Interface

![Interface do Oride](../../assets/oride-interface.png)

A interface do Oride foi desenhada para eficiência máxima no terminal, aproveitando a biblioteca Ratatui:

- **Barra Superior de Menus:** Acessível via teclas `Alt+F` (File), `Alt+E` (Edit), `Alt+V` (View), `Alt+G` (Go), `Alt+I` (Git) e `Alt+H` (Help).
- **Abas de Buffers (Tabs):** Exibe todos os arquivos abertos, indicando o ativo com fundo destacado e alterações não salvas com um indicador `*`.
- **Árvore de Arquivos (Project Tree):** Painel lateral esquerdo com navegação recursiva, ícones Nerd Fonts e status Git (`M`, `A`, `?`).
- **Área de Edição de Texto:** Suporta números de linha (gutter), múltiplos cursores, quebra de linha suave (`Alt+Z`) e realce sintático Tree-Sitter.
- **Painel Git / SCM:** Acessível via `Ctrl+Shift+G` no lado direito, listando arquivos modificados na working tree.
- **Terminal Embutido (PTY):** Terminal interativo colapsável na parte inferior (`Ctrl+\`` ou `Ctrl+"`).
- **Linha de Status:** Rodapé exibindo posição do cursor (linha, coluna), codificação, linguagem detectada, modo de edição ativo e mensagens do sistema.

---

## 3. Modos de Edição

O Oride oferece suporte a dois paradigmas de edição que convivem de forma integrada:

### Modo Padrão (Estilo Micro)
Por padrão, o editor funciona de forma intuitiva como a maioria dos editores de terminal modernos:
- Setas para mover o cursor, `Home`/`End`, `PageUp`/`PageDown`.
- Seleção direta com `Shift + Setas` ou com o mouse (quando `mouse = true` nas configs).
- Atalhos universais: `Ctrl+S` para salvar, `Ctrl+C` para copiar, `Ctrl+V` para colar, `Ctrl+Z` para desfazer e `Ctrl+Y` para refazer.

### Modo Modal (Estilo Vim)
Para desenvolvedores habituados à edição modal sem tirar as mãos da linha base:

1. **Ativação:** Defina `modal_mode = true` no seu `config.toml`, ou abra a Command Palette (`Ctrl+Shift+P`) e escolha **Toggle modal mode (Vim / CUA)** para alternar sem sair do editor.
2. **Modos Disponíveis:**
   - **NORMAL:** Modo padrão para navegação e operações em texto. O cursor é exibido em bloco.
   - **INSERT (`i`, `a`, `o`, `I`, `A`, `O`):** Modo de inserção direta de texto. Pressione `Esc` para retornar ao modo NORMAL.
   - **VISUAL (`v`):** Seleção de texto por caracteres.
   - **VISUAL LINE (`V`):** Seleção por linhas inteiras.
   - **COMMAND (`:`):** Linha de comando no rodapé para executar ações administrativas do editor.

3. **Movimentos Essenciais no Modo Normal:**
   - `h`, `j`, `k`, `l`: Esquerda, baixo, cima, direita.
   - `w`, `b`: Próxima palavra, início da palavra anterior.
   - `0`, `$`: Início da linha, fim da linha.
   - `gg`, `G`: Início do buffer, fim do buffer.
   - `x`: Deletar caractere sob o cursor.
   - `u`: Desfazer alteração.

4. **Comandos de Linha de Comando (`:`):**
   - `:w` — Salva o arquivo ativo.
   - `:q` — Fecha o editor ou a aba atual.
   - `:q!` — Força o fechamento sem salvar alterações pendentes.
   - `:wq` — Salva o arquivo e sai do editor.
   - `:e <caminho>` — Abre um novo arquivo no buffer.
   - `:tasks` — Abre a janela do Task Runner.
   - `:health` — Abre o painel de diagnóstico de saúde do sistema e LSPs.
   - `:theme <nome>` — Altera o tema visual (ex: `:theme tokyo-night`).
   - `:lang <código>` — Troca o idioma da interface (ex: `:lang en-US`).
   - `:noh` — Limpa a seleção da busca atual.
   - `:run <tarefa>` — Executa uma tarefa do runner pelo nome.

   Para ver todos os atalhos, use `F1`, `Ctrl+G` ou `Ctrl+Shift+/` — não há `:help`.

---

## 4. Divisão de Janelas (Splits)

O Oride permite dividir o espaço de trabalho em múltiplos painéis de edição simultâneos:

- **Dividir Verticalmente:** `Ctrl+Alt+V` (abre um novo painel lado a lado).
- **Dividir Horizontalmente:** `Ctrl+Alt+H` (abre um painel acima/abaixo).
- **Navegar entre Painéis:** `F6` (avança o foco para o próximo split).
- **Fechar Painel Ativo:** `Ctrl+Alt+W`.

---

## 5. Localização e Substituição Avançada

### No Arquivo Ativo
- `Ctrl+F`: Abre o mini-modal de busca no rodapé.
- `F3`: Salta para a próxima ocorrência encontrada.
- `Ctrl+H`: Abre o painel de busca e substituição (tecle `Tab` para alternar entre os campos de busca e substituição).
- **Modificadores de Busca (Toggle):**
  - `Alt+C`: Diferencia maiúsculas de minúsculas (Case Sensitive).
  - `Alt+A`: Ignora acentos (ex: busca por `funcao` encontra `função`).
  - `Alt+W`: Palavra inteira.
  - `Alt+R`: Ativa suporte a **Expressões Regulares (Regex)** completas.
- `Alt+Enter`: Substitui a ocorrência atual.
- `Ctrl+Alt+Enter`: Substitui todas as ocorrências de uma vez.

### No Projeto Inteiro
- `Ctrl+Shift+F`: Abre a ferramenta de busca no workspace. Utiliza o `ripgrep` (`rg`) do sistema caso disponível ou o motor nativo em Rust como fallback. Permite aplicar filtros de glob (ex: `*.rs, !target/*`).

---

## 6. Task Runner Integrado

O Oride possui um executor de tarefas embutido para automatizar compilações, testes e scripts sem precisar sair do editor.

### Como Funciona
Basta criar um arquivo `tasks.toml` na raiz do seu projeto ou em `.oride/tasks.toml`:

```toml
[tasks.build]
label = "Cargo: Build"
command = "cargo build"
description = "Compila o workspace em modo debug"

[tasks.test]
label = "Cargo: Test Active Module"
command = "cargo test $FILE_NAME"
description = "Executa os testes do módulo aberto"

[tasks.run]
label = "Executar Binário"
command = "cargo run"

[tasks.format]
label = "Formatar Código"
command = "cargo fmt --all"
```

### Variáveis Mágicas Disponíveis:
- `$FILE`: Caminho completo absoluto do arquivo ativo.
- `$FILE_NAME`: Nome do arquivo ativo (ex: `main.rs`).
- `$FILE_STEM`: Nome do arquivo sem extensão (ex: `main`).
- `$FILE_DIR`: Diretório pai do arquivo ativo.
- `$WORKSPACE`: Raiz do workspace/projeto aberto.
- `$LINE`: Linha atual do cursor (1-based).
- `$COL`: Coluna atual do cursor (1-based).

### Execução:
Digite `:tasks` na linha de comando ou pressione `Ctrl+Shift+P` para abrir a paleta e selecionar a tarefa desejada.

---

## 7. Diagnóstico e Saúde do Sistema (`:health`)

Para garantir que todas as ferramentas do seu ambiente de desenvolvimento estejam funcionando perfeitamente, o Oride oferece o comando `:health`.

Ao digitar `:health` ou acionar a opção no menu *Help*:
- O editor inspeciona a presença e versões de ferramentas essenciais no seu `PATH`: `git`, `rg` (ripgrep), `cargo`, `gcc`/`clang`, servidores LSP.
- Exibe o status de conectividade com os servidores LSP configurados (`rust-analyzer`, `clangd`, `ori-lsp`, etc.).
- Informa o estado de renderização de fontes e protocolo gráfico do terminal.

---

## 8. Pré-visualização de Markdown no Terminal

O Oride inclui um motor avançado de renderização de Markdown no próprio terminal:

- **Atalho:** Pressione `Ctrl+Shift+V` ou `Alt+P` em qualquer arquivo `.md`.
- **Recursos Suportados:**
  - Tabelas desenhadas com linhas Unicode elegantes e alinhamento de colunas.
  - Blocos de código com realce sintático embutido (fenced code injection).
  - Listas de tarefas interativas (`[ ]` e `[x]`).
  - Suporte a gráficos de terminal para imagens nos protocolos Kitty, Ghostty e WezTerm (ative `terminal_images = true` no seu `config.toml`).

---

## 9. Temas e Customização Visual

O Oride vem com mais de 10 temas integrados de alta qualidade (`tokyo-night`, `dracula`, `nord`, `one-dark`, `catppuccin-mocha`, `monokai`, `github-dark`, etc.) e suporta temas de terceiros sem necessidade de recompilação.

- **Trocar de Tema:** Abra a Command Palette (`Ctrl+Shift+P`), busque por *Theme* e navegue com as setas para ver o **Live Preview** imediato.
- **Criar Tema Personalizado:**
  Coloque um arquivo `.toml` em `~/.config/oride/themes/meu-tema.toml`. O editor detecta o arquivo automaticamente na próxima inicialização. Consulte o [Guia de Temas](themes.md) para a especificação completa de paleta e tokens de syntax.

---

## 10. Configuração & Internacionalização

### Arquivos de Configuração
O Oride carrega a configuração mesclando em camadas:
1. Padrões embutidos no binário.
2. Configuração global do usuário: `~/.config/oride/config.toml`.
3. Configuração local do workspace: `.oride/config.toml`.

### Idiomas Disponíveis
O editor suporta tradução dinâmica de interface. No seu `config.toml`:

```toml
locale = "pt-BR" # ou "en-US"
```

Novos idiomas podem ser adicionados criando arquivos TOML em `~/.config/oride/locales/<código>.toml` sem recompilar o executável.
