//! Testes automatizados de UI/UX de ponta a ponta (End-to-End Headless TUI).
//!
//! Utiliza `ratatui::backend::TestBackend` para renderizar em matriz 2D em memória,
//! simulando cliques de mouse, arrasto, atalhos de teclado, multi-cursor e modais,
//! inspecionando o buffer de tela resultante sem necessidade de interação manual.

use std::fs;
use std::path::PathBuf;

use crossterm::event::{
    KeyCode, KeyEvent, KeyEventKind, KeyEventState, KeyModifiers, MouseButton, MouseEvent,
    MouseEventKind,
};
use oride_app::{App, CompletionChoice, KeyCommand, Locale, Overlay};
use oride_core::ByteOffset;
use oride_keymap::Action;
use ratatui::backend::TestBackend;
use ratatui::Terminal;
use tempfile::tempdir;

/// Harness de automação de testes de UI/UX.
struct UiHarness {
    app: App,
    terminal: Terminal<TestBackend>,
}

impl UiHarness {
    /// Habilita o mouse no app sob teste.
    ///
    /// O produto vem com mouse desligado (a captura toma o terminal do usuário),
    /// mas este harness existe justamente para dirigir mouse e teclado — então
    /// ele opta explicitamente em vez de depender do default. O default do
    /// produto é verificado por `mouse_is_off_by_default`.
    fn opt_in_to_mouse(app: &mut App) {
        app.mouse_enabled = true;
    }

    fn new(width: u16, height: u16) -> Self {
        let backend = TestBackend::new(width, height);
        let terminal = Terminal::new(backend).expect("falha ao criar terminal virtual de teste");
        let workspace = std::env::temp_dir();
        let mut store = oride_core::DocumentStore::new();
        store.open_empty();
        let mut app =
            App::from_store_with_config(store, oride_config::Config::default(), workspace);
        Self::opt_in_to_mouse(&mut app);
        let mut harness = Self { app, terminal };
        harness.render();
        harness
    }

    fn with_workspace(dir: PathBuf, width: u16, height: u16) -> Self {
        let backend = TestBackend::new(width, height);
        let terminal = Terminal::new(backend).expect("falha ao criar terminal virtual");
        let mut app = App::open_workspace(dir).expect("falha ao abrir workspace");
        Self::opt_in_to_mouse(&mut app);
        let mut harness = Self { app, terminal };
        harness.render();
        harness
    }

    fn render(&mut self) {
        self.terminal
            .draw(|frame| self.app.draw(frame))
            .expect("falha ao desenhar frame no TestBackend");
    }

    fn screen_text(&self) -> String {
        let buffer = self.terminal.backend().buffer();
        let mut text = String::new();
        for row in 0..buffer.area.height {
            for col in 0..buffer.area.width {
                text.push_str(buffer[(col, row)].symbol());
            }
            text.push('\n');
        }
        text
    }

    fn click(&mut self, col: u16, row: u16) {
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Down(MouseButton::Left),
            column: col,
            row,
            modifiers: KeyModifiers::empty(),
        });
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Up(MouseButton::Left),
            column: col,
            row,
            modifiers: KeyModifiers::empty(),
        });
        self.render();
    }

    fn alt_click(&mut self, col: u16, row: u16) {
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Down(MouseButton::Left),
            column: col,
            row,
            modifiers: KeyModifiers::ALT,
        });
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Up(MouseButton::Left),
            column: col,
            row,
            modifiers: KeyModifiers::ALT,
        });
        self.render();
    }

    fn double_click(&mut self, col: u16, row: u16) {
        self.click(col, row);
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Down(MouseButton::Left),
            column: col,
            row,
            modifiers: KeyModifiers::empty(),
        });
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Up(MouseButton::Left),
            column: col,
            row,
            modifiers: KeyModifiers::empty(),
        });
        self.render();
    }

    fn triple_click(&mut self, col: u16, row: u16) {
        self.double_click(col, row);
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Down(MouseButton::Left),
            column: col,
            row,
            modifiers: KeyModifiers::empty(),
        });
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Up(MouseButton::Left),
            column: col,
            row,
            modifiers: KeyModifiers::empty(),
        });
        self.render();
    }

    fn drag(&mut self, start_col: u16, start_row: u16, end_col: u16, end_row: u16) {
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Down(MouseButton::Left),
            column: start_col,
            row: start_row,
            modifiers: KeyModifiers::empty(),
        });
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Drag(MouseButton::Left),
            column: end_col,
            row: end_row,
            modifiers: KeyModifiers::empty(),
        });
        self.app.handle_mouse(MouseEvent {
            kind: MouseEventKind::Up(MouseButton::Left),
            column: end_col,
            row: end_row,
            modifiers: KeyModifiers::empty(),
        });
        self.render();
    }

    fn scroll(&mut self, is_up: bool, col: u16, row: u16) {
        let kind = if is_up {
            MouseEventKind::ScrollUp
        } else {
            MouseEventKind::ScrollDown
        };
        self.app.handle_mouse(MouseEvent {
            kind,
            column: col,
            row,
            modifiers: KeyModifiers::empty(),
        });
        self.render();
    }

    fn key(&mut self, code: KeyCode) {
        self.app.handle_key(KeyEvent {
            code,
            modifiers: KeyModifiers::empty(),
            kind: KeyEventKind::Press,
            state: KeyEventState::empty(),
        });
        self.render();
    }

    fn key_ctrl(&mut self, code: KeyCode) {
        self.app.handle_key(KeyEvent {
            code,
            modifiers: KeyModifiers::CONTROL,
            kind: KeyEventKind::Press,
            state: KeyEventState::empty(),
        });
        self.render();
    }

    fn key_alt(&mut self, code: KeyCode) {
        self.app.handle_key(KeyEvent {
            code,
            modifiers: KeyModifiers::ALT,
            kind: KeyEventKind::Press,
            state: KeyEventState::empty(),
        });
        self.render();
    }

    fn key_ctrl_shift(&mut self, code: KeyCode) {
        self.app.handle_key(KeyEvent {
            code,
            modifiers: KeyModifiers::CONTROL | KeyModifiers::SHIFT,
            kind: KeyEventKind::Press,
            state: KeyEventState::empty(),
        });
        self.render();
    }

    fn type_text(&mut self, text: &str) {
        for character in text.chars() {
            if character == '\n' {
                self.key(KeyCode::Enter);
            } else {
                self.key(KeyCode::Char(character));
            }
        }
        self.render();
    }
}

#[test]
fn mouse_is_off_by_default() {
    // O harness opta por mouse para poder dirigi-lo; este teste garante que o
    // produto, sem opt-in, não captura o terminal do usuário.
    let mut store = oride_core::DocumentStore::new();
    store.open_empty();
    let app =
        App::from_store_with_config(store, oride_config::Config::default(), std::env::temp_dir());
    assert!(
        !app.mouse_enabled,
        "mouse precisa vir desligado: com captura ligada, seleção e cópia nativas do terminal param de funcionar"
    );
    assert!(!oride_config::Config::default().mouse);
}

#[test]
fn test_ui_menu_bar_mouse_interaction() {
    let mut harness = UiHarness::new(100, 30);
    harness.app.update_locale(Locale::EnUs);
    harness.render();
    let initial_screen = harness.screen_text();
    assert!(initial_screen.contains("File"));
    assert!(initial_screen.contains("Edit"));
    assert!(initial_screen.contains("View"));

    // 1. Clique na coordenada do menu "File" (coluna 2, linha 0)
    harness.click(2, 0);
    let menu_screen = harness.screen_text();
    // Dropdown do menu File deve estar visível na tela
    assert!(menu_screen.contains("New tab") || menu_screen.contains("Save"));

    // 2. Clique fora do dropdown fecha o menu
    harness.click(80, 20);
    let closed_screen = harness.screen_text();
    assert!(!closed_screen.contains("New tab"));

    // 3. Clique em "View" (coluna 14, linha 0)
    harness.click(14, 0);
    let view_menu_screen = harness.screen_text();
    assert!(
        view_menu_screen.contains("Toggle tree") || view_menu_screen.contains("Toggle terminal")
    );

    // 4. Abre o menu "File" (coluna 2, linha 0) e clica no item "New tab" (coluna 4, linha 2)
    let initial_tabs = harness.app.store.tab_summaries().len();
    harness.click(2, 0);
    assert!(harness.screen_text().contains("New tab"));
    // Linha 2, coluna 4 é dentro do item "New tab"
    harness.click(4, 2);
    // Menu deve ter fechado e nova aba deve ter sido aberta
    assert!(!harness.screen_text().contains("Save as…"));
    assert_eq!(
        harness.app.store.tab_summaries().len(),
        initial_tabs + 1,
        "Clicar em 'New tab' no submenu deve criar uma nova aba"
    );
}

#[test]
fn test_ui_tabs_mouse_switching() {
    let mut harness = UiHarness::new(100, 30);

    // Digita no primeiro documento
    harness.type_text("CONTEUDO_ABA_1");
    let active_doc1 = harness.app.store.active().unwrap().buffer().as_string();
    assert_eq!(active_doc1, "CONTEUDO_ABA_1");

    // Cria nova aba via atalho
    harness.app.apply(KeyCommand::Action(Action::NewTab));
    harness.render();
    harness.type_text("CONTEUDO_ABA_2");
    let active_doc2 = harness.app.store.active().unwrap().buffer().as_string();
    assert_eq!(active_doc2, "CONTEUDO_ABA_2");

    // Na barra de abas (linha 2), clica na primeira aba (aba 0 à esquerda)
    let tabs_x = harness.app.config.tree.width + 3;
    harness.click(tabs_x, 2);

    // Agora o documento ativo deve ter voltado para a Aba 1
    let switched_doc = harness.app.store.active().unwrap().buffer().as_string();
    assert_eq!(switched_doc, "CONTEUDO_ABA_1");

    // Clica na segunda aba (mais à direita, tabs_x + 18)
    harness.click(tabs_x + 18, 2);
    let switched_back = harness.app.store.active().unwrap().buffer().as_string();
    assert_eq!(switched_back, "CONTEUDO_ABA_2");
}

#[test]
fn test_ui_editor_mouse_clicks_and_selections() {
    let mut harness = UiHarness::new(100, 30);
    harness.type_text("hello world antigravity\nsecond line of text\n");

    let editor_x = harness.app.config.tree.width + 8;
    let editor_y = 3; // Linha de texto no editor

    // 1. Clique simples: posiciona cursor
    harness.click(editor_x, editor_y);
    let doc = harness.app.store.active().unwrap();
    assert!(doc.selection().is_empty());

    // 2. Duplo clique em "world": seleciona a palavra
    let world_x = editor_x + 6;
    harness.double_click(world_x, editor_y);
    let doc = harness.app.store.active().unwrap();
    let selected_text = doc.selected_text();
    assert_eq!(selected_text, "world");

    // 3. Triplo clique na linha: seleciona a linha inteira
    harness.triple_click(editor_x, editor_y);
    let doc = harness.app.store.active().unwrap();
    let selected_line = doc.selected_text();
    assert!(selected_line.contains("hello world antigravity"));

    // 4. Arrasto (Drag): seleciona intervalo
    harness.drag(editor_x, editor_y, editor_x + 5, editor_y);
    let doc = harness.app.store.active().unwrap();
    assert!(!doc.selection().is_empty());
}

#[test]
fn test_ui_multi_cursor_synchronized_movement_and_editing() {
    let mut harness = UiHarness::new(100, 30);
    harness.type_text("first line\nsecond line\nthird line");

    // Move cursor para o início do documento
    harness
        .app
        .apply(KeyCommand::Action(Action::MoveDocStart { extend: false }));
    harness.render();

    // 1. Adiciona cursor na segunda linha
    harness
        .app
        .apply(KeyCommand::Action(Action::AddCursorBelow));
    // 2. Adiciona cursor na terceira linha
    harness
        .app
        .apply(KeyCommand::Action(Action::AddCursorBelow));
    harness.render();

    {
        let doc = harness.app.store.active().unwrap();
        assert_eq!(doc.extra_carets().len(), 2);
        assert_eq!(doc.all_caret_offsets().len(), 3);
    }

    // 3. Movimenta com Seta Direita: TODOS os cursores devem avançar sincronizados!
    harness
        .app
        .apply(KeyCommand::Action(Action::MoveRight { extend: false }));
    harness.render();
    {
        let doc = harness.app.store.active().unwrap();
        assert_eq!(
            doc.extra_carets().len(),
            2,
            "os cursores secundários NÃO devem sumir ao mover"
        );
        // Verifica que todos avançaram para offset 1 de suas respectivas linhas
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(1), ByteOffset::new(12), ByteOffset::new(24)]
        );
    }

    // 4. Digitação simultânea nos 3 cursores
    harness.type_text("X");
    {
        let doc = harness.app.store.active().unwrap();
        assert_eq!(
            doc.buffer().as_string(),
            "fXirst line\nsXecond line\ntXhird line"
        );
        assert_eq!(doc.extra_carets().len(), 2);
    }

    // 5. Backspace simultâneo nos 3 cursores
    harness.app.apply(KeyCommand::Action(Action::Backspace));
    harness.render();
    {
        let doc = harness.app.store.active().unwrap();
        assert_eq!(
            doc.buffer().as_string(),
            "first line\nsecond line\nthird line"
        );
        assert_eq!(doc.extra_carets().len(), 2);
    }

    // 6. Move para o início da linha (Home)
    harness
        .app
        .apply(KeyCommand::Action(Action::MoveLineStart { extend: false }));
    harness.render();
    {
        let doc = harness.app.store.active().unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![ByteOffset::new(0), ByteOffset::new(11), ByteOffset::new(23)]
        );
    }

    // 7. Move para o fim da linha (End)
    harness
        .app
        .apply(KeyCommand::Action(Action::MoveLineEnd { extend: false }));
    harness.render();
    {
        let doc = harness.app.store.active().unwrap();
        assert_eq!(
            doc.all_caret_offsets(),
            vec![
                ByteOffset::new(10),
                ByteOffset::new(22),
                ByteOffset::new(33)
            ]
        );
    }

    // 8. Pressiona Esc: colapsa multi-cursores em um único cursor
    harness.key(KeyCode::Esc);
    {
        let doc = harness.app.store.active().unwrap();
        assert!(
            doc.extra_carets().is_empty(),
            "Esc deve colapsar cursores secundários"
        );
    }
}

#[test]
fn test_ui_multi_cursor_alt_click_and_mouse_collapse() {
    let mut harness = UiHarness::new(100, 30);
    harness.type_text("alpha item\nbeta item\ngamma item");

    let editor_x = harness.app.config.tree.width + 6;

    // Alt+Click na linha 1
    harness.alt_click(editor_x + 2, 4);
    // Alt+Click na linha 2
    harness.alt_click(editor_x + 2, 5);

    {
        let doc = harness.app.store.active().unwrap();
        assert!(
            !doc.extra_carets().is_empty(),
            "Alt+Click deve criar cursores secundários"
        );
    }

    // Clique normal sem Alt: deve limpar multi-cursores e posicionar um único cursor
    harness.click(editor_x + 4, 3);
    {
        let doc = harness.app.store.active().unwrap();
        assert!(
            doc.extra_carets().is_empty(),
            "clique simples sem Alt deve colapsar cursores"
        );
    }
}

#[test]
fn test_ui_command_palette_execution() {
    let mut harness = UiHarness::new(100, 30);

    // Abre Command Palette via atalho Ctrl+Shift+P
    harness.key_ctrl_shift(KeyCode::Char('p'));
    let palette_screen = harness.screen_text();
    assert!(palette_screen.contains("Command Palette") || palette_screen.contains(">"));

    // Digita filtro para alternar números de linha
    harness.type_text("line numbers");
    // Pressiona Enter para executar a ação filtrada
    harness.key(KeyCode::Enter);

    // O status bar deve confirmar a alteração e o modal deve fechar
    let screen_after = harness.screen_text();
    assert!(!screen_after.contains("Command Palette"));
}

#[test]
fn test_ui_find_and_replace_workflow() {
    let mut harness = UiHarness::new(100, 30);
    harness.type_text("apples and apples everywhere");

    // Abre Find modal via Ctrl+F
    harness.key_ctrl(KeyCode::Char('f'));
    let find_screen = harness.screen_text();
    assert!(find_screen.contains("Find / Replace"));

    // Digita busca "apples"
    harness.type_text("apples");
    let hits_screen = harness.screen_text();
    assert!(hits_screen.contains("match") || hits_screen.contains("/"));

    // Alterna para campo replace via Tab
    harness.key(KeyCode::Tab);
    harness.type_text("oranges");

    // Substitui tudo via Ctrl+Alt+Enter
    harness.app.handle_key(KeyEvent {
        code: KeyCode::Enter,
        modifiers: KeyModifiers::CONTROL | KeyModifiers::ALT,
        kind: KeyEventKind::Press,
        state: KeyEventState::empty(),
    });
    harness.render();

    // Fecha o modal via Esc
    harness.key(KeyCode::Esc);

    // Verifica que o texto foi substituído no buffer
    let doc = harness.app.store.active().unwrap();
    assert_eq!(doc.buffer().as_string(), "oranges and oranges everywhere");
}

#[test]
fn test_ui_split_panes_no_black_screen_on_close() {
    let mut harness = UiHarness::new(100, 30);
    harness.type_text("PANE_1_CONTENT");

    // Abre nova aba
    harness.app.apply(KeyCommand::Action(Action::NewTab));
    harness.render();
    harness.type_text("PANE_2_CONTENT");

    // Abre split vertical
    harness.app.apply(KeyCommand::Action(Action::SplitVertical));
    harness.render();

    let split_screen = harness.screen_text();
    // Ambos os painéis devem renderizar texto na tela
    assert!(split_screen.contains("PANE_2_CONTENT"));

    // Fecha a aba ativa com split aberto (duas vezes para confirmar descarte de aba modificada)
    harness.app.apply(KeyCommand::Action(Action::CloseTab));
    harness.app.apply(KeyCommand::Action(Action::CloseTab));
    harness.render();

    let after_close_screen = harness.screen_text();
    // A tela NÃO pode estar preta/vazia; deve exibir PANE_1_CONTENT sincronizado
    assert!(after_close_screen.contains("PANE_1_CONTENT"));
    assert!(!harness.app.should_quit);
}

#[test]
fn test_ui_file_tree_double_click_opens_file() {
    let temp_directory = tempdir().expect("tempdir");
    let file_path = temp_directory.path().join("sample_code.rs");
    fs::write(&file_path, "fn sample_function() -> i32 { 100 }\n").expect("write file");

    let mut harness = UiHarness::with_workspace(temp_directory.path().to_path_buf(), 100, 30);

    let screen = harness.screen_text();
    assert!(screen.contains("sample_code.rs"));

    // Clica duas vezes sobre a linha do arquivo na árvore (linha 1 da árvore = y=4, pois y=2 é borda e y=3 é a pasta raiz)
    harness.double_click(4, 4);

    // O arquivo deve ter sido aberto no editor
    let doc = harness.app.store.active().unwrap();
    assert_eq!(doc.tab_title(), "sample_code.rs");
    assert!(doc.buffer().as_string().contains("sample_function"));
}

#[test]
fn test_ui_scroll_mouse_wheel() {
    let mut harness = UiHarness::new(100, 30);
    for index in 0..50 {
        harness.type_text(&format!("Line {index}\n"));
    }

    let editor_x = harness.app.config.tree.width + 10;
    let editor_y = 10;

    // Rola para baixo com a rodinha
    harness.scroll(false, editor_x, editor_y);
    let scroll_down = harness.app.scroll_y;
    assert!(scroll_down > 0, "scroll down deve aumentar scroll_y");

    // Rola para cima com a rodinha
    harness.scroll(true, editor_x, editor_y);
    assert!(
        harness.app.scroll_y < scroll_down,
        "scroll up deve diminuir scroll_y"
    );
}

#[test]
fn test_ui_markdown_preview_and_interaction() {
    let temp_directory = tempdir().expect("tempdir");
    let file_path = temp_directory.path().join("doc.md");
    let md_content = "# Doc Title\n\n| ColA | ColB |\n| :--- | ---: |\n| Left | Right |\n\nSee [Manual](https://example.com) here.\n\n```rust\nfn test() {\n    let val = 10;\n}\n```\n";
    fs::write(&file_path, md_content).expect("write md");

    let mut harness = UiHarness::with_workspace(temp_directory.path().to_path_buf(), 120, 35);
    // Abre o arquivo doc.md dando duplo clique no item da árvore
    harness.double_click(4, 4);
    assert_eq!(harness.app.store.active().unwrap().tab_title(), "doc.md");

    // Ativa preview com ToggleMdPreview
    harness
        .app
        .apply(KeyCommand::Action(Action::ToggleMdPreview));
    harness.render();

    let screen = harness.screen_text();
    // Deve renderizar o painel de preview
    assert!(
        screen.contains("preview"),
        "deve conter cabeçalho de preview"
    );
    // Deve conter a tabela renderizada com Unicode box drawing
    assert!(
        screen.contains('┌') && screen.contains('┐'),
        "deve conter moldura superior da tabela"
    );
    assert!(
        screen.contains("ColA") && screen.contains("ColB"),
        "deve conter cabeçalhos da tabela"
    );
    assert!(
        screen.contains('└') && screen.contains('┘'),
        "deve conter moldura inferior da tabela"
    );
    // Deve conter link e código
    assert!(screen.contains("Manual"), "deve renderizar texto do link");
    assert!(
        screen.contains("fn test()"),
        "deve conter o bloco de código"
    );

    // Testa scroll na área do preview (coluna 90, linha 10)
    harness.scroll(false, 90, 10);
    harness.scroll(true, 90, 10);

    // Testa atalho Alt+Enter no preview
    harness.key_alt(KeyCode::Enter);
    let status = harness.app.status_message.as_deref().unwrap_or("");
    assert!(
        status.contains("Abrindo")
            || status.contains("https://example.com")
            || status.contains("link")
            || status.contains("Falha")
    );
}

#[test]
fn test_ui_file_tree_rename_delete_and_copy_path() {
    let temp_directory = tempdir().expect("tempdir");
    let file_path = temp_directory.path().join("original_name.txt");
    fs::write(&file_path, "sample content").expect("write file");

    let mut harness = UiHarness::with_workspace(temp_directory.path().to_path_buf(), 100, 30);
    harness.render();

    // Foca na árvore e seleciona o arquivo
    harness.app.apply(KeyCommand::Action(Action::FocusTree));
    harness.key(KeyCode::Down);
    harness.render();

    // Copia path (y)
    harness.key(KeyCode::Char('y'));
    let status = harness.app.status_message.as_deref().unwrap_or("");
    assert!(status.contains("copiado") && status.contains("original_name.txt"));

    // Renomeia (r)
    harness.key(KeyCode::Char('r'));
    harness.render();
    assert!(harness.screen_text().contains("renomear"));

    // Digita novo nome e pressiona Enter
    for _ in 0..30 {
        harness.key(KeyCode::Backspace);
    }
    harness.type_text("brand_new_name.txt");
    harness.key(KeyCode::Enter);
    harness.render();

    let new_path = temp_directory.path().join("brand_new_name.txt");
    assert!(
        new_path.exists(),
        "arquivo deve ter sido renomeado no disco"
    );
    assert!(!file_path.exists(), "arquivo antigo não deve mais existir");
    assert!(harness.screen_text().contains("brand_new_name.txt"));

    // Deleta (d)
    harness.key(KeyCode::Char('d'));
    harness.render();
    assert!(harness.screen_text().contains("confirmar exclusão"));

    harness.key(KeyCode::Enter);
    harness.render();

    assert!(
        !new_path.exists(),
        "arquivo deve ter sido excluído do disco"
    );
    let tree_rows = harness.app.tree.as_ref().unwrap().flat_rows();
    assert!(!tree_rows.iter().any(|r| r.name == "brand_new_name.txt"));
}

#[test]
fn test_ui_git_scm_stage_unstage_and_commit() {
    let temp_directory = tempdir().expect("tempdir");
    let path = temp_directory.path();
    let _ = std::process::Command::new("git")
        .args(["init"])
        .current_dir(path)
        .output()
        .unwrap();
    let _ = std::process::Command::new("git")
        .args(["config", "user.name", "Test"])
        .current_dir(path)
        .output()
        .unwrap();
    let _ = std::process::Command::new("git")
        .args(["config", "user.email", "test@test.com"])
        .current_dir(path)
        .output()
        .unwrap();

    let sample_file = path.join("git_tracked.rs");
    fs::write(&sample_file, "fn foo() {}\n").unwrap();

    let mut harness = UiHarness::with_workspace(path.to_path_buf(), 120, 35);
    harness.render();

    // Abre o painel SCM
    harness.app.apply(KeyCommand::Action(Action::FocusScm));
    harness.render();

    let screen = harness.screen_text();
    assert!(screen.contains("git_tracked.rs"));

    // Stage (s)
    harness.key(KeyCode::Char('s'));
    harness.render();
    let status = harness.app.status_message.as_deref().unwrap_or("");
    assert!(status.contains("staged") && status.contains("git_tracked.rs"));

    // Unstage (u)
    harness.key(KeyCode::Char('u'));
    harness.render();
    let status = harness.app.status_message.as_deref().unwrap_or("");
    assert!(status.contains("unstaged") && status.contains("git_tracked.rs"));

    // Stage novamente para commit
    harness.key(KeyCode::Char('s'));
    harness.render();

    // Abre diálogo de commit (c)
    harness.key(KeyCode::Char('c'));
    harness.render();
    assert!(harness.screen_text().contains("git commit"));

    // Digita mensagem e faz commit
    harness.type_text("first commit from test");
    harness.key(KeyCode::Enter);
    harness.render();

    let status = harness.app.status_message.as_deref().unwrap_or("");
    assert!(status.contains("commit concluído") || status.contains("first commit from test"));
}

#[test]
fn test_ui_project_search_and_replace() {
    let temp_directory = tempdir().expect("tempdir");
    let path = temp_directory.path();

    let file_a = path.join("file_a.txt");
    let file_b = path.join("file_b.rs");
    fs::write(&file_a, "hello OLD_PROJECT_TEXT world\n").unwrap();
    fs::write(&file_b, "fn run() { let x = \"OLD_PROJECT_TEXT\"; }\n").unwrap();

    let mut harness = UiHarness::with_workspace(path.to_path_buf(), 120, 35);
    harness.render();

    // Abre file_a.txt no editor para testar recarregamento automático do buffer
    harness.app.open_document_path(&file_a).unwrap();
    harness.render();

    // Dispara busca no projeto
    harness.app.apply(KeyCommand::Action(Action::ProjectFind));
    harness.render();
    assert!(harness.screen_text().contains("find in project"));

    // Digita termo de busca
    harness.type_text("OLD_PROJECT_TEXT");
    harness.render();

    // Pressiona Tab para alternar para o campo Replace
    harness.key(KeyCode::Tab);
    harness.render();
    assert!(harness.screen_text().contains("Repl:"));

    // Digita termo de substituição
    harness.type_text("NEW_PROJECT_TEXT");
    harness.render();

    // Pressiona Enter para aplicar substituição no projeto
    harness.key(KeyCode::Enter);
    harness.render();

    // Status deve confirmar substituições ou recarregamento
    let status = harness.app.status_message.as_deref().unwrap_or("");
    assert!(
        status.contains("Substituídas") || status.contains("recarregado"),
        "status={status}"
    );

    // Valida no disco
    let content_a = fs::read_to_string(&file_a).unwrap();
    let content_b = fs::read_to_string(&file_b).unwrap();
    assert!(content_a.contains("NEW_PROJECT_TEXT") && !content_a.contains("OLD_PROJECT_TEXT"));
    assert!(content_b.contains("NEW_PROJECT_TEXT") && !content_b.contains("OLD_PROJECT_TEXT"));

    // Valida que o buffer ativo em memória também refletiu a substituição
    let active_buffer_text = harness.app.store.active().unwrap().buffer().as_string();
    assert!(active_buffer_text.contains("NEW_PROJECT_TEXT"));
}

#[test]
fn test_ui_splitter_mouse_drag_and_pane_resize() {
    let temp_directory = tempdir().expect("tempdir");
    let path = temp_directory.path();
    let file = path.join("code.rs");
    fs::write(&file, "fn main() {\n    println!(\"hello split\");\n}\n").unwrap();

    let mut harness = UiHarness::with_workspace(path.to_path_buf(), 120, 35);
    harness.app.mouse_enabled = true;
    harness.app.open_document_path(&file).unwrap();
    harness.render();

    // 1. Redimensionamento do divisor da Árvore via Mouse Drag
    assert!(harness.app.show_tree);
    let initial_tree_width = harness.app.tree_width;
    let tree_splitter = harness
        .app
        .hit_regions
        .tree_splitter
        .expect("divisor de árvore deve existir");

    // Arrasta divisor da árvore 8 colunas para a direita
    harness.drag(
        tree_splitter.x,
        tree_splitter.y + 2,
        tree_splitter.x + 8,
        tree_splitter.y + 2,
    );
    assert_eq!(harness.app.tree_width, initial_tree_width + 8);

    // Ajuste fino por teclado (Action::TreeShrink / Action::TreeGrow)
    harness.app.apply(KeyCommand::Action(Action::TreeShrink));
    assert_eq!(harness.app.tree_width, initial_tree_width + 6);
    harness.app.apply(KeyCommand::Action(Action::TreeGrow));
    assert_eq!(harness.app.tree_width, initial_tree_width + 8);

    // 2. Divisão de editor (split vertical) e redimensionamento via Mouse Drag
    harness.app.apply(KeyCommand::Action(Action::SplitVertical));
    harness.render();
    assert!(harness.app.split.is_split());
    assert_eq!(harness.app.split.ratio_percent, 50);

    let editor_splitter = harness
        .app
        .hit_regions
        .editor_splitter
        .expect("divisor de editor deve existir");

    // Arrasta divisor do split para a direita
    harness.drag(
        editor_splitter.x,
        editor_splitter.y + 2,
        editor_splitter.x + 12,
        editor_splitter.y + 2,
    );
    assert!(
        harness.app.split.ratio_percent > 50,
        "ratio deve ter aumentado com drag para direita"
    );

    // Ajuste fino de split via teclado
    let prev_ratio = harness.app.split.ratio_percent;
    harness
        .app
        .apply(KeyCommand::Action(Action::ResizePaneShrink));
    assert_eq!(harness.app.split.ratio_percent, prev_ratio - 5);
    harness
        .app
        .apply(KeyCommand::Action(Action::ResizePaneGrow));
    assert_eq!(harness.app.split.ratio_percent, prev_ratio);
}

#[test]
fn test_ui_session_persistence_roundtrip() {
    let dir = tempdir().unwrap();
    let workspace = dir.path().to_path_buf();
    let file1 = workspace.join("file1.rs");
    let file2 = workspace.join("file2.rs");
    fs::write(&file1, "fn one() {}\n".repeat(40)).unwrap();
    fs::write(&file2, "fn two() {}\n".repeat(40)).unwrap();

    // Cria diretório .oride para salvar session local
    let session_dir = workspace.join(".oride");
    fs::create_dir_all(&session_dir).unwrap();

    let mut harness = UiHarness::with_workspace(workspace.clone(), 120, 40);
    harness.app.open_document_path(&file1).unwrap();
    harness.app.open_document_path(&file2).unwrap();
    harness.app.tree_width = 34;
    harness.app.show_scm = true;
    harness.app.apply(KeyCommand::Action(Action::SplitVertical));
    harness.app.split.ratio_percent = 65;
    harness.app.scroll_y = 15;

    // Persiste sessão
    harness.app.persist_session();

    // Verifica que session.toml foi gravado
    let session_file = session_dir.join("session.toml");
    assert!(session_file.exists(), "session.toml deve existir em .oride");

    let saved_session =
        oride_app::Session::load_for_workspace(&workspace).expect("sessão deve ser carregada");
    assert_eq!(saved_session.tree_width, Some(34));
    assert_eq!(saved_session.show_scm, Some(true));
    assert_eq!(saved_session.scroll_y, 15);
    let split_info = saved_session.split.clone().expect("split deve estar salvo");
    assert_eq!(split_info.orientation, "vertical");
    assert_eq!(split_info.ratio_percent, 65);

    // Restaura nova instância de App a partir da sessão salva
    let restored_app = App::from_session(saved_session).expect("deve restaurar App da sessão");
    assert_eq!(restored_app.tree_width, 34);
    assert!(restored_app.show_scm);
    assert_eq!(restored_app.scroll_y, 15);
    assert!(restored_app.split.is_split());
    assert_eq!(restored_app.split.ratio_percent, 65);
}

#[test]
fn test_ui_project_search_with_glob_filter() {
    let dir = tempdir().unwrap();
    let workspace = dir.path().to_path_buf();
    let file_rs = workspace.join("source.rs");
    let file_txt = workspace.join("notes.txt");
    fs::write(&file_rs, "pub fn unique_search_target() {}\n").unwrap();
    fs::write(&file_txt, "unique_search_target in plain text\n").unwrap();

    let mut harness = UiHarness::with_workspace(workspace, 120, 40);

    // Abre ProjectFind
    harness.app.apply(KeyCommand::Action(Action::ProjectFind));
    harness.render();

    // Digita "unique_search_target"
    for ch in "unique_search_target".chars() {
        harness.key(KeyCode::Char(ch));
    }
    harness.render();

    // Verifica que encontrou ambos os arquivos
    if let oride_app::Overlay::ProjectFind { hits, .. } = &harness.app.overlay {
        assert_eq!(
            hits.len(),
            2,
            "deve encontrar hits em source.rs e notes.txt"
        );
    } else {
        panic!("deve estar no modal ProjectFind");
    }

    // Ativa filtro glob via Alt+G
    harness.key_alt(KeyCode::Char('g'));

    // Digita "*.rs" no campo Glob
    for ch in "*.rs".chars() {
        harness.key(KeyCode::Char(ch));
    }
    harness.render();

    // Verifica que o filtro glob restringiu para apenas source.rs
    if let oride_app::Overlay::ProjectFind {
        hits, file_glob, ..
    } = &harness.app.overlay
    {
        assert_eq!(file_glob.as_deref(), Some("*.rs"));
        assert_eq!(
            hits.len(),
            1,
            "apenas source.rs deve corresponder ao glob *.rs"
        );
        assert!(hits[0].path.to_string_lossy().ends_with("source.rs"));
    } else {
        panic!("deve estar no modal ProjectFind");
    }

    let screen = harness.screen_text();
    assert!(screen.contains("Glob: *.rs"));
    assert!(screen.contains("source.rs"));
}

#[test]
fn test_ui_markdown_image_metadata_and_graphics_protocol() {
    let dir = tempdir().unwrap();
    let workspace = dir.path().to_path_buf();
    let doc_path = workspace.join("doc.md");
    let png_path = workspace.join("sample.png");

    fs::write(&doc_path, "# Header\n\n![Arquitetura](./sample.png)\n").unwrap();

    // Cria PNG válido (1024x768)
    let mut png_data = vec![0x89, b'P', b'N', b'G', b'\r', b'\n', 0x1a, b'\n'];
    png_data.extend_from_slice(&[0, 0, 0, 13]); // chunk len
    png_data.extend_from_slice(b"IHDR");
    png_data.extend_from_slice(&1024u32.to_be_bytes());
    png_data.extend_from_slice(&768u32.to_be_bytes());
    png_data.extend_from_slice(&[8, 6, 0, 0, 0]);
    fs::write(&png_path, &png_data).unwrap();

    let mut harness = UiHarness::with_workspace(workspace, 120, 40);
    harness.app.open_document_path(&doc_path).unwrap();
    harness
        .app
        .apply(KeyCommand::Action(Action::ToggleMdPreview));
    harness.render();

    // Valida que o card contém resolução e formato inspecionados
    let screen = harness.screen_text();
    assert!(screen.contains("PNG"), "card deve conter o formato PNG");
    assert!(
        screen.contains("1024x768 px"),
        "card deve conter as dimensões 1024x768 px"
    );
    assert!(
        screen.contains("arquivo local encontrado"),
        "card deve indicar arquivo local encontrado"
    );

    // Ativa terminal_images na configuração e simula terminal Kitty
    harness.app.config.markdown.terminal_images = true;
    unsafe {
        std::env::set_var("KITTY_WINDOW_ID", "1");
    }
    // Força re-render do preview
    harness
        .app
        .apply(KeyCommand::Action(Action::ToggleMdPreview)); // desativa
    harness
        .app
        .apply(KeyCommand::Action(Action::ToggleMdPreview)); // reativa com nova config
    harness.render();

    let screen_kitty = harness.screen_text();
    assert!(
        screen_kitty.contains("Kitty Graphics"),
        "deve detectar protocolo Kitty com terminal_images ativo"
    );

    unsafe {
        std::env::remove_var("KITTY_WINDOW_ID");
    }
}

#[test]
fn test_ui_git_ahead_behind_and_sync_actions() {
    let mut harness = UiHarness::new(120, 30);
    harness.app.git_branch = Some("main".to_string());
    harness.app.git_ahead_behind = Some((2, 1));
    harness.render();

    let screen = harness.screen_text();
    assert!(
        screen.contains("git:main ↑2 ↓1"),
        "A barra de status deve exibir a branch com indicador ahead/behind. Tela:\n{screen}"
    );

    // Testa atalhos de pull e push no painel SCM
    harness.app.apply(KeyCommand::Action(Action::FocusScm));
    assert_eq!(harness.app.focus, oride_app::Focus::Scm);

    // Tecla 'p' no SCM aciona git pull (fail-closed seguro em ambiente de teste sem remote)
    harness.key(KeyCode::Char('p'));
    assert!(
        harness
            .app
            .status_message
            .as_deref()
            .unwrap_or("")
            .contains("git pull"),
        "Status deve reportar tentativa de git pull"
    );

    // Tecla 'P' (Shift+P) no SCM aciona git push
    harness.key(KeyCode::Char('P'));
    assert!(
        harness
            .app
            .status_message
            .as_deref()
            .unwrap_or("")
            .contains("git push"),
        "Status deve reportar tentativa de git push"
    );

    // Ações via Command Palette
    harness.app.apply(KeyCommand::Action(Action::GitPull));
    harness.app.apply(KeyCommand::Action(Action::GitPush));
}

#[test]
fn test_ui_compact_completion_dropdown_anchored_to_cursor() {
    let mut harness = UiHarness::new(100, 30);
    // Insere texto para deslocar o cursor
    harness.app.apply(KeyCommand::InsertChar('f'));
    harness.app.apply(KeyCommand::InsertChar('n'));
    harness.app.apply(KeyCommand::InsertChar(' '));
    harness.app.apply(KeyCommand::InsertChar('m'));
    harness.render();

    // Ativa overlay de autocompletar
    harness.app.overlay = Overlay::Completion {
        items: vec![
            CompletionChoice {
                display: "main() -> () — fn".to_string(),
                insert_text: "main()".to_string(),
            },
            CompletionChoice {
                display: "match — keyword".to_string(),
                insert_text: "match".to_string(),
            },
            CompletionChoice {
                display: "map(|x| ...) — method".to_string(),
                insert_text: "map".to_string(),
            },
        ],
        selected: 0,
        replace_start: 3,
    };
    harness.render();

    let screen = harness.screen_text();
    // Deve conter o título do popup compacto
    assert!(
        screen.contains("Suggest"),
        "O dropdown compacto deve exibir o bloco 'Suggest'. Tela:\n{screen}"
    );
    assert!(
        screen.contains("main() -> ()"),
        "Deve exibir os itens de autocompletar. Tela:\n{screen}"
    );
    // Não deve conter 'completions' (que era o título da paleta modal antiga de 80% de largura)
    assert!(
        !screen.contains("completions"),
        "Não deve utilizar a paleta modal antiga para completion"
    );

    // O dropdown deve ser compacto: a linha do 'Suggest' não deve ocupar 80 colunas
    let buffer = harness.terminal.backend().buffer();
    let mut found_suggest = false;
    for y in 0..buffer.area.height {
        let mut line_str = String::new();
        for x in 0..buffer.area.width {
            line_str.push_str(buffer[(x, y)].symbol());
        }
        if line_str.contains("Suggest") {
            found_suggest = true;
            let chars: Vec<char> = line_str.chars().collect();
            let suggest_idx = line_str.find("Suggest").unwrap();
            let char_offset = line_str[..suggest_idx].chars().count();
            let start_box = chars[..char_offset]
                .iter()
                .rposition(|&c| c == '┌')
                .expect("deve ter canto superior esquerdo antes de Suggest");
            let end_box = char_offset
                + chars[char_offset..]
                    .iter()
                    .position(|&c| c == '┐')
                    .expect("deve ter canto superior direito após de Suggest");
            let box_width = end_box - start_box + 1;
            assert!(
                box_width <= 45 && box_width >= 24,
                "Largura da caixa deve ser compacta (24..=45), mas foi {box_width}. Linha: '{line_str}'"
            );
        }
    }
    assert!(found_suggest, "Deveria ter encontrado a caixa 'Suggest'");

    // Navega para baixo no popup
    harness.key(KeyCode::Down);
    assert!(matches!(
        harness.app.overlay,
        Overlay::Completion { selected: 1, .. }
    ));
    harness.render();

    // Pressiona Esc para fechar o popup
    harness.key(KeyCode::Esc);
    assert_eq!(harness.app.overlay, Overlay::None);
}
