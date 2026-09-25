//! Diagnóstico proativo de saúde do sistema e servidores LSP (:health).

use std::path::PathBuf;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct LspHealthItem {
    pub language: &'static str,
    pub command: &'static str,
    pub installed: bool,
    pub path: Option<PathBuf>,
    pub install_hint: &'static str,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct HealthReport {
    pub lsp_items: Vec<LspHealthItem>,
    pub static_grammars: Vec<&'static str>,
    pub dynamic_grammars_dir: Option<PathBuf>,
    pub dynamic_grammar_count: usize,
    pub active_theme: String,
    pub active_locale: String,
    pub modal_mode: bool,
}

/// Procura um executável no `$PATH` do sistema operacional.
#[must_use]
pub fn find_in_path(cmd: &str) -> Option<PathBuf> {
    oride_osutil::find_in_path(cmd)
}

/// Realiza uma varredura preventiva de saúde do ambiente.
#[must_use]
pub fn check_health(active_theme: &str, active_locale: &str, modal_mode: bool) -> HealthReport {
    let servers = [
        (
            "Rust",
            "rust-analyzer",
            "rustup component add rust-analyzer",
        ),
        ("C / C++", "clangd", "sudo apt install clangd"),
        (
            "Bash / Shell",
            "bash-language-server",
            "npm install -g bash-language-server",
        ),
        (
            "OriScript",
            "oriscript",
            "instale oriscript no PATH (oriscript lsp)",
        ),
        (
            "Python",
            "pylsp",
            "pip install python-lsp-server (ou pyright)",
        ),
        (
            "TypeScript / JS",
            "typescript-language-server",
            "npm install -g typescript-language-server",
        ),
        ("Lua", "lua-language-server", "instale lua-language-server"),
        ("Nim", "nimlsp", "nimble install nimlsp"),
        ("D", "serve-d", "dub fetch serve-d"),
    ];

    let mut lsp_items = Vec::new();
    for (language, command, install_hint) in servers {
        let path = find_in_path(command);
        let installed = path.is_some();
        lsp_items.push(LspHealthItem {
            language,
            command,
            installed,
            path,
            install_hint,
        });
    }

    let static_grammars = vec!["Rust", "C", "Bash", "Markdown", "OriScript"];

    let dynamic_grammars_dir = dirs::config_dir().map(|d| d.join("oride").join("grammars"));
    let mut dynamic_grammar_count = 0;
    if let Some(dir) = &dynamic_grammars_dir {
        if let Ok(entries) = std::fs::read_dir(dir) {
            dynamic_grammar_count = entries
                .flatten()
                .filter(|e| {
                    e.path()
                        .extension()
                        .and_then(|s| s.to_str())
                        .is_some_and(|ext| ext == "so" || ext == "dll" || ext == "dylib")
                })
                .count();
        }
    }

    HealthReport {
        lsp_items,
        static_grammars,
        dynamic_grammars_dir,
        dynamic_grammar_count,
        active_theme: active_theme.to_string(),
        active_locale: active_locale.to_string(),
        modal_mode,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_check_health_runs_without_panic() {
        let report = check_health("Tokyo Night", "pt-BR", true);
        assert!(!report.lsp_items.is_empty());
        assert_eq!(report.active_theme, "Tokyo Night");
        assert_eq!(report.active_locale, "pt-BR");
        assert!(report.modal_mode);
        assert_eq!(report.static_grammars.len(), 5);
    }
}
