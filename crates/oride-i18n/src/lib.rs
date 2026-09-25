//! Catálogo de internacionalização (i18n) do editor Oride.
//! Suporta `pt-BR` (default), `en-US` e carregamento de `.toml` dinâmicos em disco.

use std::collections::BTreeMap;
use std::path::Path;
use std::sync::{Arc, OnceLock, RwLock};

use serde::{Deserialize, Serialize};

static DEFAULT_PT_BR_TOML: &str = include_str!("../../../internal/i18n/catalogs/pt-BR.toml");
static DEFAULT_EN_US_TOML: &str = include_str!("../../../internal/i18n/catalogs/en-US.toml");

#[derive(Debug, Clone, PartialEq, Eq, Hash, Default, Serialize, Deserialize)]
pub enum Locale {
    #[default]
    PtBr,
    EnUs,
    Custom(String),
}

impl Locale {
    #[must_use]
    pub fn from_str_loose(s: &str) -> Self {
        let trimmed = s.trim().to_ascii_lowercase().replace('_', "-");
        match trimmed.as_str() {
            "pt" | "pt-br" | "pt-pt" | "portugues" | "português" => Self::PtBr,
            "en" | "en-us" | "en-gb" | "english" | "ingles" | "inglês" => Self::EnUs,
            other => {
                let reg = global_registry();
                if let Some(matched) = reg.find_matching(other) {
                    matched
                } else {
                    Self::Custom(s.trim().to_string())
                }
            }
        }
    }

    #[must_use]
    pub fn as_str(&self) -> &str {
        match self {
            Self::PtBr => "pt-BR",
            Self::EnUs => "en-US",
            Self::Custom(id) => id.as_str(),
        }
    }

    #[must_use]
    pub fn display_name(&self) -> String {
        match self {
            Self::PtBr => "Português (Brasil)".to_string(),
            Self::EnUs => "English (US)".to_string(),
            Self::Custom(id) => {
                let reg = global_registry();
                if let Some(def) = reg.get(id) {
                    def.name.clone()
                } else {
                    id.clone()
                }
            }
        }
    }

    #[must_use]
    pub fn menu_label_for(&self) -> String {
        format!("{} [{}]", self.display_name(), self.as_str())
    }

    #[must_use]
    pub fn messages(&self) -> Messages {
        let reg = global_registry();
        let def = reg.get(self.as_str()).cloned().unwrap_or_else(|| {
            if self.as_str().to_ascii_lowercase().starts_with("en") {
                reg.get("en-US").cloned().expect("embedded en-US locale")
            } else {
                reg.get("pt-BR").cloned().expect("embedded pt-BR locale")
            }
        });
        Messages { def }
    }
}

pub static AVAILABLE_LOCALES: &[Locale] = &[Locale::PtBr, Locale::EnUs];

#[must_use]
pub fn available_locales() -> Vec<Locale> {
    let reg = global_registry();
    reg.list_locales()
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct MenuMessages {
    pub file: String,
    pub edit: String,
    pub view: String,
    pub go: String,
    pub git: String,
    pub help: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct FileMenuMessages {
    pub new_tab: String,
    pub open_file: String,
    pub open_folder: String,
    pub save: String,
    pub save_as: String,
    pub save_all: String,
    pub reload_file: String,
    pub quit: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct EditMenuMessages {
    pub undo: String,
    pub redo: String,
    pub cut: String,
    pub copy: String,
    pub paste: String,
    pub select_all: String,
    pub find: String,
    pub find_in_project: String,
    pub replace: String,
    pub toggle_comment: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct ViewMenuMessages {
    pub command_palette: String,
    pub toggle_tree: String,
    pub toggle_terminal: String,
    pub toggle_scm: String,
    pub color_theme: String,
    pub display_language: String,
    pub toggle_mouse: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct PaletteMessages {
    pub theme_picker_title: String,
    pub theme_picker_hint: String,
    pub locale_picker_title: String,
    pub locale_picker_hint: String,
    pub palette_hint: String,
    pub theme_applied: String,
    pub theme_cancelled: String,
    pub locale_applied: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct StatusMessages {
    pub mouse_on: String,
    pub mouse_off: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct LocaleDefinition {
    pub id: String,
    pub name: String,
    pub menu: MenuMessages,
    pub file_menu: FileMenuMessages,
    pub edit_menu: EditMenuMessages,
    pub view_menu: ViewMenuMessages,
    pub palette: PaletteMessages,
    pub status: StatusMessages,
}

#[derive(Debug, Clone)]
pub struct Messages {
    def: Arc<LocaleDefinition>,
}

impl Messages {
    #[must_use]
    pub fn file(&self) -> &str {
        if self.def.menu.file.is_empty() {
            "Arquivo"
        } else {
            &self.def.menu.file
        }
    }

    #[must_use]
    pub fn edit(&self) -> &str {
        if self.def.menu.edit.is_empty() {
            "Editar"
        } else {
            &self.def.menu.edit
        }
    }

    #[must_use]
    pub fn view(&self) -> &str {
        if self.def.menu.view.is_empty() {
            "Exibir"
        } else {
            &self.def.menu.view
        }
    }

    #[must_use]
    pub fn go(&self) -> &str {
        if self.def.menu.go.is_empty() {
            "Ir"
        } else {
            &self.def.menu.go
        }
    }

    #[must_use]
    pub fn git(&self) -> &str {
        if self.def.menu.git.is_empty() {
            "Git"
        } else {
            &self.def.menu.git
        }
    }

    #[must_use]
    pub fn help(&self) -> &str {
        if self.def.menu.help.is_empty() {
            "Ajuda"
        } else {
            &self.def.menu.help
        }
    }

    // Menu File
    #[must_use]
    pub fn new_tab(&self) -> &str {
        if self.def.file_menu.new_tab.is_empty() {
            "Nova aba"
        } else {
            &self.def.file_menu.new_tab
        }
    }

    #[must_use]
    pub fn open_file(&self) -> &str {
        if self.def.file_menu.open_file.is_empty() {
            "Abrir arquivo…"
        } else {
            &self.def.file_menu.open_file
        }
    }

    #[must_use]
    pub fn open_folder(&self) -> &str {
        if self.def.file_menu.open_folder.is_empty() {
            "Abrir pasta…"
        } else {
            &self.def.file_menu.open_folder
        }
    }

    #[must_use]
    pub fn save(&self) -> &str {
        if self.def.file_menu.save.is_empty() {
            "Salvar"
        } else {
            &self.def.file_menu.save
        }
    }

    #[must_use]
    pub fn save_as(&self) -> &str {
        if self.def.file_menu.save_as.is_empty() {
            "Salvar como…"
        } else {
            &self.def.file_menu.save_as
        }
    }

    #[must_use]
    pub fn save_all(&self) -> &str {
        if self.def.file_menu.save_all.is_empty() {
            "Salvar todos"
        } else {
            &self.def.file_menu.save_all
        }
    }

    #[must_use]
    pub fn reload_file(&self) -> &str {
        if self.def.file_menu.reload_file.is_empty() {
            "Recarregar arquivo"
        } else {
            &self.def.file_menu.reload_file
        }
    }

    #[must_use]
    pub fn quit(&self) -> &str {
        if self.def.file_menu.quit.is_empty() {
            "Sair"
        } else {
            &self.def.file_menu.quit
        }
    }

    // Menu Edit
    #[must_use]
    pub fn undo(&self) -> &str {
        if self.def.edit_menu.undo.is_empty() {
            "Desfazer"
        } else {
            &self.def.edit_menu.undo
        }
    }

    #[must_use]
    pub fn redo(&self) -> &str {
        if self.def.edit_menu.redo.is_empty() {
            "Refazer"
        } else {
            &self.def.edit_menu.redo
        }
    }

    #[must_use]
    pub fn cut(&self) -> &str {
        if self.def.edit_menu.cut.is_empty() {
            "Recortar"
        } else {
            &self.def.edit_menu.cut
        }
    }

    #[must_use]
    pub fn copy(&self) -> &str {
        if self.def.edit_menu.copy.is_empty() {
            "Copiar"
        } else {
            &self.def.edit_menu.copy
        }
    }

    #[must_use]
    pub fn paste(&self) -> &str {
        if self.def.edit_menu.paste.is_empty() {
            "Colar"
        } else {
            &self.def.edit_menu.paste
        }
    }

    #[must_use]
    pub fn select_all(&self) -> &str {
        if self.def.edit_menu.select_all.is_empty() {
            "Selecionar tudo"
        } else {
            &self.def.edit_menu.select_all
        }
    }

    #[must_use]
    pub fn find(&self) -> &str {
        if self.def.edit_menu.find.is_empty() {
            "Localizar…"
        } else {
            &self.def.edit_menu.find
        }
    }

    #[must_use]
    pub fn find_in_project(&self) -> &str {
        if self.def.edit_menu.find_in_project.is_empty() {
            "Localizar no projeto…"
        } else {
            &self.def.edit_menu.find_in_project
        }
    }

    #[must_use]
    pub fn replace(&self) -> &str {
        if self.def.edit_menu.replace.is_empty() {
            "Substituir…"
        } else {
            &self.def.edit_menu.replace
        }
    }

    #[must_use]
    pub fn toggle_comment(&self) -> &str {
        if self.def.edit_menu.toggle_comment.is_empty() {
            "Alternar comentário"
        } else {
            &self.def.edit_menu.toggle_comment
        }
    }

    // Menu View
    #[must_use]
    pub fn command_palette(&self) -> &str {
        if self.def.view_menu.command_palette.is_empty() {
            "Paleta de comandos…"
        } else {
            &self.def.view_menu.command_palette
        }
    }

    #[must_use]
    pub fn toggle_tree(&self) -> &str {
        if self.def.view_menu.toggle_tree.is_empty() {
            "Alternar árvore"
        } else {
            &self.def.view_menu.toggle_tree
        }
    }

    #[must_use]
    pub fn toggle_terminal(&self) -> &str {
        if self.def.view_menu.toggle_terminal.is_empty() {
            "Alternar terminal"
        } else {
            &self.def.view_menu.toggle_terminal
        }
    }

    #[must_use]
    pub fn toggle_scm(&self) -> &str {
        if self.def.view_menu.toggle_scm.is_empty() {
            "Painel Git/SCM"
        } else {
            &self.def.view_menu.toggle_scm
        }
    }

    #[must_use]
    pub fn color_theme(&self) -> &str {
        if self.def.view_menu.color_theme.is_empty() {
            "Tema de cores…"
        } else {
            &self.def.view_menu.color_theme
        }
    }

    #[must_use]
    pub fn display_language(&self) -> &str {
        if self.def.view_menu.display_language.is_empty() {
            "Idioma da interface (Language)…"
        } else {
            &self.def.view_menu.display_language
        }
    }

    #[must_use]
    pub fn toggle_mouse(&self) -> &str {
        if self.def.view_menu.toggle_mouse.is_empty() {
            "Ativar / desativar mouse"
        } else {
            &self.def.view_menu.toggle_mouse
        }
    }

    // Palette & Overlays
    #[must_use]
    pub fn theme_picker_title(&self) -> &str {
        if self.def.palette.theme_picker_title.is_empty() {
            "Preferências: Tema de Cores (Live Preview)"
        } else {
            &self.def.palette.theme_picker_title
        }
    }

    #[must_use]
    pub fn theme_picker_hint(&self) -> &str {
        if self.def.palette.theme_picker_hint.is_empty() {
            "↑↓ Preview · Enter Aplica · Esc Cancela"
        } else {
            &self.def.palette.theme_picker_hint
        }
    }

    #[must_use]
    pub fn locale_picker_title(&self) -> &str {
        if self.def.palette.locale_picker_title.is_empty() {
            "Preferências: Idioma da Interface"
        } else {
            &self.def.palette.locale_picker_title
        }
    }

    #[must_use]
    pub fn locale_picker_hint(&self) -> &str {
        if self.def.palette.locale_picker_hint.is_empty() {
            "↑↓ Seleciona · Enter Aplica · Esc Cancela"
        } else {
            &self.def.palette.locale_picker_hint
        }
    }

    #[must_use]
    pub fn palette_hint(&self) -> &str {
        if self.def.palette.palette_hint.is_empty() {
            "↑↓ · Enter executa · Esc"
        } else {
            &self.def.palette.palette_hint
        }
    }

    #[must_use]
    pub fn theme_applied(&self, theme: &str) -> String {
        let tmpl = if self.def.palette.theme_applied.is_empty() {
            "tema aplicado e salvo: {name}"
        } else {
            &self.def.palette.theme_applied
        };
        tmpl.replace("{name}", theme)
    }

    #[must_use]
    pub fn theme_cancelled(&self) -> &str {
        if self.def.palette.theme_cancelled.is_empty() {
            "troca de tema cancelada"
        } else {
            &self.def.palette.theme_cancelled
        }
    }

    #[must_use]
    pub fn locale_applied(&self, locale: &str) -> String {
        let tmpl = if self.def.palette.locale_applied.is_empty() {
            "idioma alterado e salvo: {locale}"
        } else {
            &self.def.palette.locale_applied
        };
        tmpl.replace("{locale}", locale)
    }

    #[must_use]
    pub fn mouse_status(&self, enabled: bool) -> &str {
        if enabled {
            if self.def.status.mouse_on.is_empty() {
                "mouse: ON · clique/drag/scroll · desligar: Exibir → Mouse"
            } else {
                &self.def.status.mouse_on
            }
        } else if self.def.status.mouse_off.is_empty() {
            "mouse: OFF · ligar: Exibir → Mouse"
        } else {
            &self.def.status.mouse_off
        }
    }
}

pub fn normalize_locale_id(id: &str) -> String {
    id.trim().to_lowercase().replace(['_', ' '], "-")
}

#[derive(Debug, Clone)]
pub struct LocaleRegistry {
    locales: BTreeMap<String, Arc<LocaleDefinition>>,
}

impl Default for LocaleRegistry {
    fn default() -> Self {
        Self::new()
    }
}

impl LocaleRegistry {
    #[must_use]
    pub fn new() -> Self {
        let mut reg = Self {
            locales: BTreeMap::new(),
        };
        if let Ok(pt) = toml::from_str::<LocaleDefinition>(DEFAULT_PT_BR_TOML) {
            reg.register(pt);
        }
        if let Ok(en) = toml::from_str::<LocaleDefinition>(DEFAULT_EN_US_TOML) {
            reg.register(en);
        }
        reg
    }

    #[must_use]
    pub fn load_with_paths(workspace_hint: Option<&Path>) -> Self {
        let mut reg = Self::new();
        if let Some(user_dir) = dirs::config_dir().map(|d| d.join("oride").join("locales")) {
            reg.load_from_dir(&user_dir);
        }
        if let Some(ws) = workspace_hint {
            reg.load_from_dir(&ws.join(".oride").join("locales"));
        }
        reg
    }

    pub fn register(&mut self, def: LocaleDefinition) {
        let key = normalize_locale_id(&def.id);
        self.locales.insert(key, Arc::new(def));
    }

    #[must_use]
    pub fn get(&self, id: &str) -> Option<&Arc<LocaleDefinition>> {
        let key = normalize_locale_id(id);
        self.locales.get(&key)
    }

    #[must_use]
    pub fn list_locales(&self) -> Vec<Locale> {
        self.locales
            .values()
            .map(|def| {
                let id = def.id.as_str();
                if id.eq_ignore_ascii_case("pt-BR") {
                    Locale::PtBr
                } else if id.eq_ignore_ascii_case("en-US") {
                    Locale::EnUs
                } else {
                    Locale::Custom(def.id.clone())
                }
            })
            .collect()
    }

    pub fn load_from_dir(&mut self, dir: &Path) {
        if !dir.is_dir() {
            return;
        }
        let Ok(entries) = std::fs::read_dir(dir) else {
            return;
        };
        for entry in entries.flatten() {
            let path = entry.path();
            if path.extension().and_then(|e| e.to_str()) == Some("toml") {
                if let Ok(content) = std::fs::read_to_string(&path) {
                    if let Ok(def) = toml::from_str::<LocaleDefinition>(&content) {
                        self.register(def);
                    }
                }
            }
        }
    }

    #[must_use]
    pub fn find_matching(&self, s: &str) -> Option<Locale> {
        let query = normalize_locale_id(s);
        for def in self.locales.values() {
            if normalize_locale_id(&def.id) == query || def.name.to_lowercase().contains(&query) {
                return Some(if def.id.eq_ignore_ascii_case("pt-BR") {
                    Locale::PtBr
                } else if def.id.eq_ignore_ascii_case("en-US") {
                    Locale::EnUs
                } else {
                    Locale::Custom(def.id.clone())
                });
            }
        }
        None
    }
}

static REGISTRY: OnceLock<RwLock<LocaleRegistry>> = OnceLock::new();

pub fn global_registry() -> std::sync::RwLockReadGuard<'static, LocaleRegistry> {
    REGISTRY
        .get_or_init(|| RwLock::new(LocaleRegistry::load_with_paths(None)))
        .read()
        .expect("locale registry lock")
}

pub fn reload_locales(workspace_hint: Option<&Path>) {
    let reg = LocaleRegistry::load_with_paths(workspace_hint);
    if let Ok(mut lock) = REGISTRY
        .get_or_init(|| RwLock::new(LocaleRegistry::new()))
        .write()
    {
        *lock = reg;
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn locale_from_str_loose() {
        assert_eq!(Locale::from_str_loose("pt-BR"), Locale::PtBr);
        assert_eq!(Locale::from_str_loose("pt_BR"), Locale::PtBr);
        assert_eq!(Locale::from_str_loose("pt"), Locale::PtBr);
        assert_eq!(Locale::from_str_loose("en-US"), Locale::EnUs);
        assert_eq!(Locale::from_str_loose("en"), Locale::EnUs);
        assert_eq!(Locale::from_str_loose("english"), Locale::EnUs);
    }

    #[test]
    fn messages_translation() {
        let pt = Locale::PtBr.messages();
        let en = Locale::EnUs.messages();
        assert_eq!(pt.file(), "Arquivo");
        assert_eq!(en.file(), "File");
        assert_eq!(pt.save(), "Salvar");
        assert_eq!(en.save(), "Save");
    }

    #[test]
    fn loads_custom_locale_from_toml() {
        let toml_data = r#"
id = "es-ES"
name = "Español"

[menu]
file = "Archivo"
edit = "Editar"
view = "Ver"
go = "Ir"
help = "Ayuda"

[file_menu]
save = "Guardar"
quit = "Salir"
"#;
        let def: LocaleDefinition = toml::from_str(toml_data).expect("parse toml");
        assert_eq!(def.id, "es-ES");
        assert_eq!(def.menu.file, "Archivo");

        let mut reg = LocaleRegistry::new();
        reg.register(def);
        let es = reg.get("es-ES").expect("found es-ES");
        assert_eq!(es.menu.file, "Archivo");
        let msg = Messages { def: es.clone() };
        assert_eq!(msg.file(), "Archivo");
        assert_eq!(msg.save(), "Guardar");
        // Fallback para campos omitidos
        assert_eq!(msg.git(), "Git");
    }

    #[test]
    fn discovers_and_loads_locales_from_directory() {
        let tmp = tempfile::tempdir().expect("tempdir");
        let fr_file = tmp.path().join("fr-FR.toml");
        std::fs::write(
            &fr_file,
            r#"
id = "fr-FR"
name = "Français (France)"

[menu]
file = "Fichier"
edit = "Édition"
view = "Affichage"
go = "Aller"
git = "Git"
help = "Aide"

[file_menu]
new_tab = "Nouvel onglet"
save = "Enregistrer"
quit = "Quitter"
"#,
        )
        .expect("write fr-FR.toml");

        let mut reg = LocaleRegistry::new();
        reg.load_from_dir(tmp.path());

        let fr = reg.get("fr-FR").expect("found fr-FR");
        assert_eq!(fr.name, "Français (France)");
        let msg = Messages { def: fr.clone() };
        assert_eq!(msg.file(), "Fichier");
        assert_eq!(msg.edit(), "Édition");
        assert_eq!(msg.save(), "Enregistrer");
        assert_eq!(msg.new_tab(), "Nouvel onglet");

        // Checa se aparece na listagem
        let all = reg.list_locales();
        assert!(all.iter().any(|l| l.as_str() == "fr-FR"));
    }
}
