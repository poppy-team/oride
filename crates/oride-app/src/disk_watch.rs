//! Watch de arquivos abertos (reload se mudarem no disco).

use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::sync::mpsc::{self, Receiver};
use std::time::SystemTime;

use notify::{EventKind, RecommendedWatcher, RecursiveMode, Watcher};

/// Eventos de mudança em caminho observado.
#[derive(Debug, Clone)]
pub struct DiskChange {
    pub path: PathBuf,
}

pub struct DiskWatch {
    _watcher: Option<RecommendedWatcher>,
    receiver: Option<Receiver<DiskChange>>,
    /// Último mtime que nós mesmos gravamos (evita auto-reload do próprio save).
    ignore_mtime: HashMap<PathBuf, SystemTime>,
}

fn is_ignored_path(path: &Path) -> bool {
    path.components().any(|component| {
        let component_name = component.as_os_str().to_string_lossy();
        component_name == "target"
            || component_name == ".git"
            || component_name == "node_modules"
            || component_name == ".oride"
    })
}

impl DiskWatch {
    /// Sem watcher: `poll()` sempre devolve vazio.
    ///
    /// Usado pelo modo conformance, onde observar o disco tornaria a execução
    /// dependente de timing e do ambiente.
    #[must_use]
    pub fn disabled() -> Self {
        Self {
            _watcher: None,
            receiver: None,
            ignore_mtime: HashMap::new(),
        }
    }

    pub fn start(workspace: &Path) -> Self {
        let (sender, receiver) = mpsc::channel::<DiskChange>();
        let mut watcher =
            match notify::recommended_watcher(move |result: Result<notify::Event, _>| {
                if let Ok(event) = result {
                    if matches!(
                        event.kind,
                        EventKind::Modify(_) | EventKind::Create(_) | EventKind::Remove(_)
                    ) {
                        for path in event.paths {
                            if !is_ignored_path(&path) {
                                let _ = sender.send(DiskChange { path });
                            }
                        }
                    }
                }
            }) {
                Ok(watcher) => watcher,
                Err(_) => {
                    return Self {
                        _watcher: None,
                        receiver: None,
                        ignore_mtime: HashMap::new(),
                    };
                }
            };
        let _ = watcher.watch(workspace, RecursiveMode::Recursive);
        Self {
            _watcher: Some(watcher),
            receiver: Some(receiver),
            ignore_mtime: HashMap::new(),
        }
    }

    pub fn mark_saved(&mut self, path: &Path) {
        if let Ok(metadata) = std::fs::metadata(path) {
            if let Ok(modified_time) = metadata.modified() {
                self.ignore_mtime.insert(path.to_path_buf(), modified_time);
            }
        }
    }

    pub fn poll(&mut self) -> Vec<PathBuf> {
        let Some(receiver) = &self.receiver else {
            return Vec::new();
        };
        let mut changed_paths = Vec::new();
        while let Ok(change) = receiver.try_recv() {
            if is_ignored_path(&change.path) {
                continue;
            }
            if let Ok(metadata) = std::fs::metadata(&change.path) {
                if let Ok(modified_time) = metadata.modified() {
                    if self.ignore_mtime.get(&change.path) == Some(&modified_time) {
                        continue;
                    }
                }
            }
            if change.path.is_file() && !changed_paths.contains(&change.path) {
                changed_paths.push(change.path);
            }
        }
        changed_paths
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn ignores_target_and_git_directories() {
        assert!(is_ignored_path(Path::new("target/debug/build")));
        assert!(is_ignored_path(Path::new(".git/objects/abc")));
        assert!(is_ignored_path(Path::new("project/node_modules/pkg")));
        assert!(is_ignored_path(Path::new(".oride/workspace.toml")));
        assert!(!is_ignored_path(Path::new("src/main.rs")));
        assert!(!is_ignored_path(Path::new("crates/oride-app/src/app.rs")));
    }
}
