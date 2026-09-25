//! Modo conformance: execução headless determinística + dump canônico de estado.
//!
//! Existe para dar ao porte Go um **oráculo diferencial**: o mesmo caso é
//! executado aqui e na implementação Go, e os dumps são comparados sem
//! tolerância. Por isso tudo neste módulo é determinístico — sem PTY, sem
//! watcher de disco, sem LSP, sem plugins externos e sem ler a config do
//! usuário. Um caso só depende dos arquivos que ele mesmo declara.
//!
//! O formato do dump é versionado por [`DUMP_SCHEMA_VERSION`]: mudar o shape
//! exige incrementar a constante, para que uma fixture antiga falhe de forma
//! explícita em vez de comparar campos que já não existem.

use std::collections::{BTreeMap, HashMap};
use std::fmt;
use std::path::{Component, Path, PathBuf};

use oride_config::Config;
use oride_core::{Document, DocumentError, DocumentId, DocumentStore};
use oride_fs::{list_files_recursive, ProjectTree};
use oride_git::{scm_entries, status_map};
use oride_i18n::Locale;
use oride_keymap::{parse_action, parse_chord, Keymap};
use oride_plugin::builtin_host;
use oride_syntax::HighlightEngine;
use oride_ui::UiTheme;
use serde::{Deserialize, Serialize};

use crate::app::state::{App, Focus, KeyCommand, Overlay};
use crate::browser::BrowseMode;
use crate::disk_watch::DiskWatch;
use crate::find::FindState;
use crate::jump_list::JumpList;
use crate::mouse::HitRegions;
use crate::session::{Session, SplitSession};
use crate::split::{SplitOrientation, SplitState};

/// Versão do shape de [`StateDump`]. Incrementar ao mudar qualquer campo.
pub const DUMP_SCHEMA_VERSION: u32 = 1;

/// Marcador que substitui o caminho absoluto do workspace no dump.
const WORKSPACE_PLACEHOLDER: &str = "<workspace>";

/// Erros do modo conformance.
#[derive(Debug)]
pub enum ConformanceError {
    Document(DocumentError),
    Json(serde_json::Error),
    Io {
        path: PathBuf,
        source: std::io::Error,
    },
    Keymap(String),
    /// O chord do caso não está ligado a nenhuma ação na config efetiva.
    UnboundChord(String),
    /// O id de ação do caso não existe.
    UnknownAction(String),
    /// O caminho declarado pelo caso escapa do workspace.
    PathEscape(String),
}

impl fmt::Display for ConformanceError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Document(error) => write!(formatter, "documento: {error}"),
            Self::Json(error) => write!(formatter, "json: {error}"),
            Self::Io { path, source } => write!(formatter, "io em `{}`: {source}", path.display()),
            Self::Keymap(message) => write!(formatter, "keymap: {message}"),
            Self::UnboundChord(chord) => write!(formatter, "chord sem binding: `{chord}`"),
            Self::UnknownAction(id) => write!(formatter, "ação desconhecida: `{id}`"),
            Self::PathEscape(path) => write!(formatter, "caminho fora do workspace: `{path}`"),
        }
    }
}

impl std::error::Error for ConformanceError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Document(error) => Some(error),
            Self::Json(error) => Some(error),
            Self::Io { source, .. } => Some(source),
            _ => None,
        }
    }
}

impl From<DocumentError> for ConformanceError {
    fn from(error: DocumentError) -> Self {
        Self::Document(error)
    }
}

impl From<serde_json::Error> for ConformanceError {
    fn from(error: serde_json::Error) -> Self {
        Self::Json(error)
    }
}

// ---------------------------------------------------------------------------
// Caso
// ---------------------------------------------------------------------------

/// Um caso de conformance: arquivos, config e passos.
#[derive(Debug, Clone, Deserialize)]
pub struct ConformanceCase {
    #[serde(default)]
    pub description: Option<String>,
    /// Arquivos materializados no workspace antes da execução.
    #[serde(default)]
    pub files: Vec<CaseFile>,
    /// Arquivo aberto ao iniciar (relativo ao workspace). Ausente ⇒ buffer vazio.
    #[serde(default)]
    pub open: Option<String>,
    /// Overrides sobre `Config::default()`.
    #[serde(default)]
    pub config: CaseConfig,
    /// Habilita as consultas a git (rótulo de status, cache SCM). Desligado por
    /// padrão porque cada consulta é um subprocesso: um caso que não testa git
    /// não deve depender de o binário existir.
    #[serde(default)]
    pub git: bool,
    /// Sequência de passos. O estado é capturado após cada um.
    pub steps: Vec<Step>,
}

/// Arquivo declarado pelo caso.
#[derive(Debug, Clone, Deserialize)]
pub struct CaseFile {
    pub path: String,
    pub content: String,
}

/// Overrides de configuração. Campo ausente mantém o default do produto.
#[derive(Debug, Clone, Default, Deserialize)]
pub struct CaseConfig {
    pub theme: Option<String>,
    pub locale: Option<String>,
    pub show_line_numbers: Option<bool>,
    pub soft_wrap: Option<bool>,
    pub mouse: Option<bool>,
    pub modal_mode: Option<bool>,
    pub tab_size: Option<u8>,
    pub insert_spaces: Option<bool>,
    pub tree_width: Option<u16>,
    pub show_tree: Option<bool>,
    pub show_scm: Option<bool>,
    pub show_hidden: Option<bool>,
    /// Overrides de `[keys]` (chord → id de ação) sobre os bindings default.
    #[serde(default)]
    pub keys: BTreeMap<String, String>,
}

impl CaseConfig {
    fn apply_to(&self, config: &mut Config) {
        if let Some(value) = &self.theme {
            config.theme = value.clone();
        }
        if let Some(value) = &self.locale {
            config.locale = value.clone();
        }
        if let Some(value) = self.show_line_numbers {
            config.show_line_numbers = value;
        }
        if let Some(value) = self.soft_wrap {
            config.soft_wrap = value;
        }
        if let Some(value) = self.mouse {
            config.mouse = value;
        }
        if let Some(value) = self.modal_mode {
            config.editor.modal_mode = value;
        }
        if let Some(value) = self.tab_size {
            config.editor.tab_size = value;
        }
        if let Some(value) = self.insert_spaces {
            config.editor.insert_spaces = value;
        }
        if let Some(value) = self.tree_width {
            config.tree.width = value;
        }
        if let Some(value) = self.show_hidden {
            config.tree.show_hidden = value;
        }
        for (chord, action) in &self.keys {
            config.keys.insert(chord.clone(), action.clone());
        }
    }
}

/// Um passo da execução.
#[derive(Debug, Clone, Deserialize)]
#[serde(tag = "kind", rename_all = "snake_case")]
pub enum Step {
    /// Tecla resolvida pelo keymap da config efetiva. Falha se não houver
    /// binding — é assim que o caso detecta drift de keymap.
    Chord { chord: String },
    /// Texto digitado caractere a caractere.
    Text { text: String },
    /// Ação disparada pelo id estável, sem passar pelo keymap.
    Action { action: String },
    /// Ajusta o estado de busca no buffer e recalcula.
    ///
    /// Um driver scriptado não consegue digitar dentro de um overlay, e o que o
    /// caso quer exercitar é o motor de busca — dobra de acentos, fronteira de
    /// palavra, regex e offsets — não a sobreposição visual.
    Search {
        query: String,
        #[serde(default)]
        replace: Option<String>,
        #[serde(default)]
        case_sensitive: bool,
        #[serde(default)]
        ignore_accents: bool,
        #[serde(default)]
        whole_word: bool,
        #[serde(default)]
        use_regex: bool,
    },
}

impl ConformanceCase {
    pub fn from_json(source: &str) -> Result<Self, ConformanceError> {
        Ok(serde_json::from_str(source)?)
    }

    pub fn from_path(path: &Path) -> Result<Self, ConformanceError> {
        let source = std::fs::read_to_string(path).map_err(|source| ConformanceError::Io {
            path: path.to_path_buf(),
            source,
        })?;
        Self::from_json(&source)
    }
}

// ---------------------------------------------------------------------------
// Resultado
// ---------------------------------------------------------------------------

/// Resultado de um caso: o estado após cada passo.
#[derive(Debug, Clone, Serialize)]
pub struct CaseReport {
    pub schema: u32,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    pub steps: usize,
    /// `frames[i]` é o estado imediatamente após o passo `frames[i].after_step`.
    pub frames: Vec<Frame>,
}

/// Estado observável em um ponto da execução.
#[derive(Debug, Clone, Serialize)]
pub struct Frame {
    /// Índice 1-based do passo que produziu este estado.
    pub after_step: usize,
    /// Descrição do passo: `chord ctrl+s`, `text "ab"`, `action save`.
    pub step: String,
    pub state: StateDump,
}

impl CaseReport {
    pub fn to_json(&self) -> Result<String, ConformanceError> {
        Ok(serde_json::to_string_pretty(self)?)
    }
}

// ---------------------------------------------------------------------------
// Dump de estado
// ---------------------------------------------------------------------------

/// Fotografia determinística e comparável do estado da aplicação.
///
/// Todos os caminhos são relativos ao workspace para que a fixture não dependa
/// de onde o caso foi executado.
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct StateDump {
    pub schema: u32,
    pub focus: &'static str,
    pub overlay: OverlayDump,
    pub status: Option<String>,
    pub tabs: Vec<TabDump>,
    pub active_tab: Option<u64>,
    pub dirty_count: usize,
    pub document: Option<DocumentDump>,
    pub find: FindDump,
    pub split: SplitDump,
    pub tree: Option<TreeDump>,
    pub show_tree: bool,
    pub show_scm: bool,
    pub scm: Vec<ScmDump>,
    pub config: ConfigDump,
    pub vim: Option<String>,
    pub diagnostics: usize,
    pub lsp_failures: Vec<String>,
    pub terminal_attached: bool,
    pub should_quit: bool,
}

/// Uma aba aberta.
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct TabDump {
    pub id: u64,
    pub title: String,
    pub path: Option<String>,
    pub dirty: bool,
    pub active: bool,
}

/// Estado do documento ativo.
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct DocumentDump {
    pub id: u64,
    pub path: Option<String>,
    pub title: String,
    pub dirty: bool,
    /// Contador de versão do documento — guarda a coalescência de edições.
    pub version: u64,
    pub lines: usize,
    pub bytes: usize,
    /// Nº de escalares Unicode: guarda a semântica de índice byte↔rune.
    pub chars: usize,
    /// Buffer completo. É o principal lastro de paridade byte a byte.
    pub text: String,
    pub caret: Option<CaretDump>,
    pub selection: SelectionDump,
    pub extra_carets: Vec<usize>,
    pub selected_text: String,
    /// Rótulos do histórico de undo, do mais antigo ao mais recente.
    pub undo_labels: Vec<String>,
    pub redo_labels: Vec<String>,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct CaretDump {
    pub line: usize,
    /// Coluna em escalares Unicode (0-based), como `oride_core::Caret`.
    pub column: usize,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct SelectionDump {
    pub anchor: usize,
    pub head: usize,
    pub empty: bool,
    pub start: usize,
    pub end: usize,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct FindDump {
    pub query: String,
    pub replace: String,
    pub matches: Vec<MatchDump>,
    pub current: usize,
    pub case_sensitive: bool,
    pub ignore_accents: bool,
    pub whole_word: bool,
    pub use_regex: bool,
    pub regex_error: Option<String>,
    pub show_replace: bool,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct MatchDump {
    pub start: usize,
    pub end: usize,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct SplitDump {
    pub panes: usize,
    pub orientation: &'static str,
    pub ratio_percent: u16,
    pub focused: usize,
    pub pane_docs: Vec<u64>,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct TreeDump {
    pub rows: usize,
    pub selected: usize,
    /// Ordem exata das linhas visíveis, com profundidade e estado de expansão.
    pub visible: Vec<TreeRowDump>,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct TreeRowDump {
    pub depth: usize,
    pub name: String,
    pub path: String,
    pub is_dir: bool,
    pub expanded: bool,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct ScmDump {
    pub badge: char,
    pub status: String,
    pub path: String,
}

/// Configuração efetiva — o que o app realmente usou, não o que foi pedido.
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct ConfigDump {
    pub theme: String,
    pub locale: String,
    pub show_line_numbers: bool,
    pub soft_wrap: bool,
    pub mouse: bool,
    pub modal_mode: bool,
    pub tab_size: u8,
    pub insert_spaces: bool,
    pub tree_width: u16,
    pub tree_show_hidden: bool,
    /// Bindings resolvidos, ordenados por chord.
    pub keys: Vec<KeyBindingDump>,
}

#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
pub struct KeyBindingDump {
    pub chord: String,
    pub action: String,
}

/// Sobreposição ativa.
///
/// Cada variante carrega apenas campos determinísticos: texto vindo de LSP ou
/// de `:health` é contado, não copiado, porque depende do ambiente e não do
/// produto.
#[derive(Debug, Clone, Serialize, PartialEq, Eq)]
#[serde(tag = "overlay", rename_all = "snake_case")]
pub enum OverlayDump {
    None,
    CommandPalette {
        query: String,
        selected: usize,
    },
    Browse {
        mode: String,
        cwd: Option<String>,
        filter: String,
        selected: usize,
        entries: usize,
    },
    Prompt {
        kind: String,
        buffer: String,
    },
    Help {
        query: String,
        selected: usize,
    },
    Find,
    ProjectFind {
        query: String,
        replace_query: Option<String>,
        file_glob: Option<String>,
        selected: usize,
        case_sensitive: bool,
        use_regex: bool,
        hits: usize,
        status: String,
        focus_field: u8,
    },
    Diagnostics {
        selected: usize,
    },
    Completion {
        items: Vec<String>,
        selected: usize,
        replace_start: usize,
    },
    Hover {
        lines: usize,
    },
    ReloadConfirm {
        path: Option<String>,
    },
    BufferPicker {
        query: String,
        selected: usize,
    },
    Diff {
        path: Option<String>,
        lines: usize,
        scroll: usize,
    },
    SurroundPick,
    MultiPicker {
        query: String,
        selected: usize,
    },
    UndoTree {
        selected: usize,
    },
    ThemePicker {
        query: String,
        selected: usize,
        initial_theme: String,
        themes: Vec<String>,
    },
    LocalePicker {
        query: String,
        selected: usize,
        locales: Vec<String>,
    },
    VimCommand {
        buffer: String,
    },
    HealthCheck {
        scroll: usize,
    },
}

// ---------------------------------------------------------------------------
// Execução
// ---------------------------------------------------------------------------

/// Executa um caso em um diretório isolado e devolve o relatório.
///
/// O chamador é dono de `workspace` e deve entregá-lo vazio: o caso o
/// materializa.
pub fn run_case(case: &ConformanceCase, workspace: &Path) -> Result<CaseReport, ConformanceError> {
    materialize(case, workspace)?;

    // Canonicaliza depois de materializar: `ProjectTree::open` canonicaliza os
    // caminhos que guarda, e comparar um caminho canonicalizado com um que não é
    // falharia em diretórios atrás de symlink (macOS `/var` → `/private/var`),
    // fazendo a normalização da raiz da árvore depender da plataforma.
    let workspace = workspace
        .canonicalize()
        .map_err(|source| ConformanceError::Io {
            path: workspace.to_path_buf(),
            source,
        })?;

    let mut app = build_app(case, &workspace)?;

    let mut frames = Vec::with_capacity(case.steps.len());
    for (index, step) in case.steps.iter().enumerate() {
        apply_step(&mut app, step)?;
        frames.push(Frame {
            after_step: index + 1,
            step: describe_step(step),
            state: dump_state(&app),
        });
    }

    Ok(CaseReport {
        schema: DUMP_SCHEMA_VERSION,
        description: case.description.clone(),
        steps: case.steps.len(),
        frames,
    })
}

/// Cria os arquivos declarados. Recusa caminhos que escapam do workspace.
fn materialize(case: &ConformanceCase, workspace: &Path) -> Result<(), ConformanceError> {
    std::fs::create_dir_all(workspace).map_err(|source| ConformanceError::Io {
        path: workspace.to_path_buf(),
        source,
    })?;
    for file in &case.files {
        let target = resolve_inside(workspace, &file.path)?;
        if let Some(parent) = target.parent() {
            std::fs::create_dir_all(parent).map_err(|source| ConformanceError::Io {
                path: parent.to_path_buf(),
                source,
            })?;
        }
        std::fs::write(&target, &file.content).map_err(|source| ConformanceError::Io {
            path: target,
            source,
        })?;
    }
    Ok(())
}

/// Resolve `relative` sob `workspace`, rejeitando `..` que saia da raiz.
fn resolve_inside(workspace: &Path, relative: &str) -> Result<PathBuf, ConformanceError> {
    let escape = || ConformanceError::PathEscape(relative.to_string());
    let mut resolved = workspace.to_path_buf();
    let mut depth = workspace.components().count();
    for component in Path::new(relative).components() {
        match component {
            Component::Normal(part) => {
                resolved.push(part);
                depth += 1;
            }
            Component::CurDir => {}
            Component::ParentDir => {
                if depth <= workspace.components().count() {
                    return Err(escape());
                }
                resolved.pop();
                depth -= 1;
            }
            Component::RootDir | Component::Prefix(_) => return Err(escape()),
        }
    }
    if !resolved.starts_with(workspace) {
        return Err(escape());
    }
    Ok(resolved)
}

/// Constrói o `App` de forma determinística.
///
/// Diferenças deliberadas em relação a `App::open_path`: sem PTY, sem watcher,
/// sem LSP e sem plugins externos ou config do usuário. O que resta é o produto
/// decidindo sobre as entradas do caso.
fn build_app(case: &ConformanceCase, workspace: &Path) -> Result<App, ConformanceError> {
    let mut config = Config::default();
    case.config.apply_to(&mut config);

    let theme = UiTheme::from_config_parts(&config.theme_ui, &config.syntax)
        .unwrap_or_else(|_| UiTheme::default());
    let keymap = Keymap::from_string_map(
        config
            .keys
            .iter()
            .map(|(chord, action)| (chord.as_str(), action.as_str())),
    )
    .map_err(|error| ConformanceError::Keymap(error.to_string()))?;

    let show_hidden = config.tree.show_hidden;
    let tree = ProjectTree::open(workspace, show_hidden).ok();
    let file_index = list_files_recursive(workspace, show_hidden).unwrap_or_default();

    let git_enabled = case.git && config.tree.git_status;
    let git_status = if git_enabled {
        status_map(workspace)
    } else {
        HashMap::new()
    };
    let scm_cache = if git_enabled {
        scm_entries(workspace)
    } else {
        Vec::new()
    };

    let locale = Locale::from_str_loose(&config.locale);
    let menus = crate::menus::menus_for_locale(&locale);
    let vim = if config.editor.modal_mode {
        Some(crate::modal::VimState::default())
    } else {
        None
    };

    let mut store = DocumentStore::new();
    match &case.open {
        Some(relative) => {
            let path = resolve_inside(workspace, relative)?;
            store.open_path(&path)?;
        }
        None => {
            store.open_empty();
        }
    }
    let active_document_id = store.active_id().unwrap_or_else(|| DocumentId::from_raw(0));

    Ok(App {
        store,
        scroll_y: 0,
        should_quit: false,
        status_message: None,
        message_expires: None,
        quit_confirm_pending: false,
        close_tab_confirm: None,
        theme,
        locale,
        show_line_numbers: config.show_line_numbers,
        keymap,
        config: config.clone(),
        last_editor_height: 20,
        last_editor_text_width: 80,
        focus: Focus::Editor,
        show_tree: case.config.show_tree.unwrap_or(true),
        tree,
        tree_scroll: 0,
        workspace: workspace.to_path_buf(),
        git_status,
        git_branch: None,
        git_ahead_behind: None,
        blame_target: None,
        blame_cache: None,
        blame_refresh_at: None,
        use_nerd_icons: true,
        tree_width: config.tree.width.max(8),
        // Sem PTY: o terminal é uma fronteira de processo, não de produto.
        terminal: None,
        overlay: Overlay::None,
        file_index,
        highlight: HighlightEngine::new(),
        soft_wrap: config.soft_wrap,
        show_md_preview: false,
        preview_scroll: 0,
        find: FindState::default(),
        disk_watch: DiskWatch::disabled(),
        lsp_clients: HashMap::new(),
        lsp_failures: HashMap::new(),
        diagnostics: Vec::new(),
        show_diagnostics: false,
        lsp_doc_version: 1,
        pending_reload: None,
        plugin_host: builtin_host(),
        split: SplitState::single(active_document_id),
        show_scm: case.config.show_scm.unwrap_or(false),
        scm_selected: 0,
        scm_width: 28,
        scm_cache,
        menu_open: None,
        show_which_key: false,
        show_welcome: false,
        jump_list: JumpList::default(),
        menus,
        hit_regions: HitRegions::default(),
        mouse_enabled: config.mouse,
        mouse_drag_anchor: None,
        splitter_drag: None,
        last_click: None,
        click_count: 0,
        surround_pending: false,
        cached_preview: None,
        focused_cursor_pos: None,
        vim,
    })
}

fn apply_step(app: &mut App, step: &Step) -> Result<(), ConformanceError> {
    match step {
        Step::Chord { chord } => {
            let parsed = parse_chord(chord)
                .map_err(|error| ConformanceError::Keymap(format!("chord `{chord}`: {error}")))?;
            let action = app
                .keymap
                .resolve_chord(parsed)
                .ok_or_else(|| ConformanceError::UnboundChord(chord.clone()))?;
            app.apply(KeyCommand::Action(action));
        }
        Step::Text { text } => {
            for character in text.chars() {
                app.apply(KeyCommand::InsertChar(character));
            }
        }
        Step::Action { action } => {
            let parsed = parse_action(action)
                .map_err(|_| ConformanceError::UnknownAction(action.clone()))?;
            app.apply(KeyCommand::Action(parsed));
        }
        Step::Search {
            query,
            replace,
            case_sensitive,
            ignore_accents,
            whole_word,
            use_regex,
        } => {
            let haystack = app.store.active()?.buffer().as_string();
            app.find.query = query.clone();
            if let Some(value) = replace {
                app.find.replace = value.clone();
            }
            app.find.case_sensitive = *case_sensitive;
            app.find.ignore_accents = *ignore_accents;
            app.find.whole_word = *whole_word;
            app.find.use_regex = *use_regex;
            app.find.recompute(&haystack);
        }
    }
    Ok(())
}

fn describe_step(step: &Step) -> String {
    match step {
        Step::Chord { chord } => format!("chord {chord}"),
        Step::Text { text } => format!("text {text:?}"),
        Step::Action { action } => format!("action {action}"),
        Step::Search {
            query,
            case_sensitive,
            whole_word,
            use_regex,
            ..
        } => format!("search {query:?} case={case_sensitive} word={whole_word} re={use_regex}"),
    }
}

// ---------------------------------------------------------------------------
// Serialização do estado
// ---------------------------------------------------------------------------

/// Captura o estado observável. Nunca lê o disco nem o relógio.
#[must_use]
pub fn dump_state(app: &App) -> StateDump {
    let active_id = app.store.active_id();
    let active = active_id.and_then(|id| app.store.get(id));

    StateDump {
        schema: DUMP_SCHEMA_VERSION,
        focus: focus_name(app.focus),
        overlay: overlay_dump(&app.overlay, &app.workspace),
        status: app
            .status_message
            .as_deref()
            .map(|message| normalize_paths(message, &app.workspace)),
        tabs: app
            .store
            .tab_ids()
            .iter()
            .map(|id| {
                let document = app.store.get(*id);
                TabDump {
                    id: id.as_u64(),
                    title: document.map_or_else(String::new, Document::tab_title),
                    path: document
                        .and_then(Document::path)
                        .map(|path| relative_path(path, &app.workspace)),
                    dirty: document.is_some_and(Document::is_dirty),
                    active: Some(*id) == active_id,
                }
            })
            .collect(),
        active_tab: active_id.map(DocumentId::as_u64),
        dirty_count: app.store.dirty_count(),
        document: active.map(|document| document_dump(document, &app.workspace)),
        find: find_dump(&app.find),
        split: SplitDump {
            panes: app.split.panes.len(),
            orientation: match app.split.orientation {
                SplitOrientation::Vertical => "vertical",
                SplitOrientation::Horizontal => "horizontal",
            },
            ratio_percent: app.split.ratio_percent,
            focused: app.split.focused,
            pane_docs: app
                .split
                .panes
                .iter()
                .map(|pane| pane.doc_id.as_u64())
                .collect(),
        },
        tree: app.tree.as_ref().map(|tree| TreeDump {
            rows: tree.count_visible_rows(),
            selected: tree.selected_index(),
            visible: tree
                .flat_rows()
                .iter()
                .map(|row| TreeRowDump {
                    depth: row.depth,
                    name: tree_row_name(row, &app.workspace),
                    path: relative_path(&row.path, &app.workspace),
                    is_dir: row.is_dir,
                    expanded: row.expanded,
                })
                .collect(),
        }),
        show_tree: app.show_tree,
        show_scm: app.show_scm,
        scm: app
            .scm_cache
            .iter()
            .map(|(status, path)| ScmDump {
                badge: status.badge(),
                status: format!("{status:?}"),
                path: relative_path(path, &app.workspace),
            })
            .collect(),
        config: ConfigDump {
            theme: app.config.theme.clone(),
            locale: app.config.locale.clone(),
            show_line_numbers: app.config.show_line_numbers,
            soft_wrap: app.config.soft_wrap,
            mouse: app.config.mouse,
            modal_mode: app.config.editor.modal_mode,
            tab_size: app.config.editor.tab_size,
            insert_spaces: app.config.editor.insert_spaces,
            tree_width: app.tree_width,
            tree_show_hidden: app.config.tree.show_hidden,
            keys: app
                .keymap
                .list_bindings()
                .iter()
                .map(|(chord, action)| KeyBindingDump {
                    chord: chord.clone(),
                    action: action.id().to_string(),
                })
                .collect(),
        },
        vim: app.vim.as_ref().map(|state| format!("{:?}", state.mode)),
        diagnostics: app.diagnostics.len(),
        lsp_failures: {
            let mut failures: Vec<String> = app
                .lsp_failures
                .iter()
                .map(|(language, message)| format!("{}: {message}", language.as_str()))
                .collect();
            failures.sort();
            failures
        },
        terminal_attached: app.terminal.is_some(),
        should_quit: app.should_quit,
    }
}

fn document_dump(document: &Document, workspace: &Path) -> DocumentDump {
    let buffer = document.buffer();
    let selection = document.selection();
    let text = buffer.as_string();
    DocumentDump {
        id: document.id().as_u64(),
        path: document.path().map(|path| relative_path(path, workspace)),
        title: document.tab_title(),
        dirty: document.is_dirty(),
        version: document.version(),
        lines: buffer.line_count(),
        bytes: buffer.len_bytes(),
        chars: text.chars().count(),
        text,
        caret: document.caret().ok().map(|caret| CaretDump {
            line: caret.line,
            column: caret.column,
        }),
        selection: SelectionDump {
            anchor: selection.anchor.as_usize(),
            head: selection.head.as_usize(),
            empty: selection.is_empty(),
            start: selection.start().as_usize(),
            end: selection.end().as_usize(),
        },
        extra_carets: document
            .extra_carets()
            .iter()
            .map(|offset| offset.as_usize())
            .collect(),
        selected_text: document.selected_text(),
        undo_labels: document.undo_history_labels(),
        redo_labels: document.redo_history_labels(),
    }
}

fn find_dump(find: &FindState) -> FindDump {
    FindDump {
        query: find.query.clone(),
        replace: find.replace.clone(),
        matches: find
            .matches
            .iter()
            .map(|range| MatchDump {
                start: range.start,
                end: range.end,
            })
            .collect(),
        current: find.current,
        case_sensitive: find.case_sensitive,
        ignore_accents: find.ignore_accents,
        whole_word: find.whole_word,
        use_regex: find.use_regex,
        regex_error: find.regex_error.clone(),
        show_replace: find.show_replace,
    }
}

fn overlay_dump(overlay: &Overlay, workspace: &Path) -> OverlayDump {
    match overlay {
        Overlay::None => OverlayDump::None,
        Overlay::CommandPalette { query, selected } => OverlayDump::CommandPalette {
            query: query.clone(),
            selected: *selected,
        },
        Overlay::Browse(browser) => OverlayDump::Browse {
            mode: match browser.mode {
                BrowseMode::Folder => "folder",
                BrowseMode::File => "file",
                BrowseMode::SaveAs => "save_as",
            }
            .to_string(),
            cwd: Some(relative_path(&browser.cwd, workspace)),
            filter: browser.filter.clone(),
            selected: browser.selected,
            entries: browser.visible_entries().len(),
        },
        Overlay::Prompt { kind, buffer } => OverlayDump::Prompt {
            kind: format!("{kind:?}"),
            buffer: buffer.clone(),
        },
        Overlay::Help { query, selected } => OverlayDump::Help {
            query: query.clone(),
            selected: *selected,
        },
        Overlay::Find => OverlayDump::Find,
        Overlay::ProjectFind {
            query,
            selected,
            case_sensitive,
            use_regex,
            hits,
            status,
            replace_query,
            file_glob,
            focus_field,
        } => OverlayDump::ProjectFind {
            query: query.clone(),
            replace_query: replace_query.clone(),
            file_glob: file_glob.clone(),
            selected: *selected,
            case_sensitive: *case_sensitive,
            use_regex: *use_regex,
            hits: hits.len(),
            status: status.clone(),
            focus_field: *focus_field,
        },
        Overlay::Diagnostics { selected } => OverlayDump::Diagnostics {
            selected: *selected,
        },
        Overlay::Completion {
            items,
            selected,
            replace_start,
        } => OverlayDump::Completion {
            items: items.iter().map(|item| item.display.clone()).collect(),
            selected: *selected,
            replace_start: *replace_start,
        },
        Overlay::Hover { text } => OverlayDump::Hover {
            lines: text.lines().count(),
        },
        Overlay::ReloadConfirm { path } => OverlayDump::ReloadConfirm {
            path: Some(relative_path(path, workspace)),
        },
        Overlay::BufferPicker { query, selected } => OverlayDump::BufferPicker {
            query: query.clone(),
            selected: *selected,
        },
        Overlay::Diff {
            path,
            lines,
            scroll,
        } => OverlayDump::Diff {
            path: Some(relative_path(path, workspace)),
            lines: lines.len(),
            scroll: *scroll,
        },
        Overlay::SurroundPick => OverlayDump::SurroundPick,
        Overlay::MultiPicker { query, selected } => OverlayDump::MultiPicker {
            query: query.clone(),
            selected: *selected,
        },
        Overlay::UndoTree { selected } => OverlayDump::UndoTree {
            selected: *selected,
        },
        Overlay::ThemePicker {
            query,
            selected,
            initial_theme,
            themes,
        } => OverlayDump::ThemePicker {
            query: query.clone(),
            selected: *selected,
            initial_theme: initial_theme.clone(),
            themes: themes.clone(),
        },
        Overlay::LocalePicker {
            query,
            selected,
            locales,
        } => OverlayDump::LocalePicker {
            query: query.clone(),
            selected: *selected,
            locales: locales.clone(),
        },
        Overlay::VimCommand { buffer } => OverlayDump::VimCommand {
            buffer: buffer.clone(),
        },
        Overlay::HealthCheck { scroll, .. } => OverlayDump::HealthCheck { scroll: *scroll },
    }
}

fn focus_name(focus: Focus) -> &'static str {
    match focus {
        Focus::Editor => "editor",
        Focus::Tree => "tree",
        Focus::Terminal => "terminal",
        Focus::Scm => "scm",
    }
}

/// Nome de exibição de uma linha da árvore.
///
/// A linha raiz mostra o basename do diretório do workspace — o único valor do
/// dump que viria do ambiente em vez de vir do produto. Sem normalizar, uma
/// fixture gravada em `/tmp/a` falha em `/tmp/b`, e a suíte passaria a depender
/// de onde rodou.
fn tree_row_name(row: &oride_fs::TreeRow, workspace: &Path) -> String {
    if row.path == workspace {
        WORKSPACE_PLACEHOLDER.to_string()
    } else {
        row.name.clone()
    }
}

fn relative_path(path: &Path, workspace: &Path) -> String {
    match path.strip_prefix(workspace) {
        Ok(relative) => format!(
            "{WORKSPACE_PLACEHOLDER}/{}",
            relative.to_string_lossy().replace('\\', "/")
        ),
        Err(_) => path.to_string_lossy().replace('\\', "/"),
    }
}

/// Troca ocorrências do caminho do workspace pelo marcador, para que mensagens
/// de status não carreguem o diretório temporário do caso.
fn normalize_paths(message: &str, workspace: &Path) -> String {
    let workspace_text = workspace.to_string_lossy();
    message.replace(workspace_text.as_ref(), WORKSPACE_PLACEHOLDER)
}

// ---------------------------------------------------------------------------
// Sessão
// ---------------------------------------------------------------------------

/// Sessão derivada do estado atual — usada pelos casos de persistência.
#[must_use]
pub fn session_dump(app: &App) -> Session {
    let open_files = app.store.open_paths();
    let active_index = app
        .store
        .active()
        .ok()
        .and_then(Document::path)
        .and_then(|path| open_files.iter().position(|file| file == path))
        .unwrap_or(0);
    let split = if app.split.is_split() {
        Some(SplitSession {
            orientation: match app.split.orientation {
                SplitOrientation::Vertical => "vertical".to_string(),
                SplitOrientation::Horizontal => "horizontal".to_string(),
            },
            secondary_file: app.split.panes.get(1).and_then(|pane| {
                app.store
                    .get(pane.doc_id)
                    .and_then(Document::path)
                    .map(Path::to_path_buf)
            }),
            ratio_percent: app.split.ratio_percent,
        })
    } else {
        None
    };
    Session {
        workspace: app.workspace.clone(),
        files: open_files,
        active_index,
        scroll_y: app.scroll_y,
        tree_width: Some(app.tree_width),
        show_tree: Some(app.show_tree),
        show_scm: Some(app.show_scm),
        split,
    }
}
