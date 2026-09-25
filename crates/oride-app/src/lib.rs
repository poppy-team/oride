//! Composição do app TUI: estado, teclas e loop principal.

mod app;
mod browser;
mod clipboard;
pub mod component;
pub mod conformance;
mod disk_watch;
mod find;
pub mod health;
mod jump_list;
mod menus;
pub mod modal;
mod mouse;
mod run;
mod session;
mod split;
pub mod tasks;
mod terminal_guard;

pub use app::{App, CompletionChoice, Focus, KeyCommand, Overlay};
pub use component::{
    Component, ComponentId, ComponentRegistry, EditorComponent, MenuBarComponent, ScmComponent,
    StatusBarComponent, TerminalComponent, TreeComponent,
};
pub use conformance::{
    dump_state, run_case, ConformanceCase, ConformanceError, StateDump, DUMP_SCHEMA_VERSION,
};
pub use mouse::{HitRegions, HitTarget};
pub use oride_i18n::Locale;
pub use run::run;
pub use session::{Session, SplitSession};
pub use split::{SplitOrientation, SplitState};
