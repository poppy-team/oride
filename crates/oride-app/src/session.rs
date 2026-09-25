//! Sessão leve: workspace + lista de arquivos abertos.

use std::fs;
use std::path::{Path, PathBuf};

use serde::{Deserialize, Serialize};

/// FNV-1a 64-bit. Offset basis e prime do padrão.
const FNV_OFFSET_BASIS: u64 = 0xcbf2_9ce4_8422_2325;
const FNV_PRIME: u64 = 0x0000_0100_0000_01b3;

/// Digest determinístico de um caminho canônico.
///
/// Não usamos `DefaultHasher`: ele é explicitamente instável entre versões do
/// Rust, e uma sessão salva por uma versão deixava de ser encontrada pela
/// seguinte — o usuário via a sessão "desaparecer" depois de um update.
/// FNV-1a é estável, não tem dependência e é reproduzível em qualquer
/// linguagem: o porte em Go precisa produzir exatamente este mesmo digest.
fn workspace_digest(canonical_workspace: &Path) -> u64 {
    let mut hash = FNV_OFFSET_BASIS;
    for byte in canonical_workspace.to_string_lossy().as_bytes() {
        hash ^= u64::from(*byte);
        hash = hash.wrapping_mul(FNV_PRIME);
    }
    hash
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub struct SplitSession {
    pub orientation: String,
    pub secondary_file: Option<PathBuf>,
    pub ratio_percent: u16,
}

#[derive(Debug, Clone, Serialize, Deserialize, Default, PartialEq, Eq)]
pub struct Session {
    pub workspace: PathBuf,
    pub files: Vec<PathBuf>,
    /// Índice na lista `files` (não DocumentId).
    pub active_index: usize,
    #[serde(default)]
    pub scroll_y: usize,
    #[serde(default)]
    pub tree_width: Option<u16>,
    #[serde(default)]
    pub show_tree: Option<bool>,
    #[serde(default)]
    pub show_scm: Option<bool>,
    #[serde(default)]
    pub split: Option<SplitSession>,
}

impl Session {
    #[must_use]
    pub fn path_for_workspace(workspace: &Path) -> Option<PathBuf> {
        let canonical_workspace =
            fs::canonicalize(workspace).unwrap_or_else(|_| workspace.to_path_buf());
        let local_dir = canonical_workspace.join(".oride");
        let local_session = local_dir.join("session.toml");
        if local_session.exists() || local_dir.is_dir() {
            return Some(local_session);
        }

        let base_directory = dirs::data_local_dir()?.join("oride").join("sessions");
        let filename = format!("{:016x}.toml", workspace_digest(&canonical_workspace));
        Some(base_directory.join(filename))
    }

    #[must_use]
    pub fn global_path() -> Option<PathBuf> {
        dirs::data_local_dir().map(|directory| directory.join("oride").join("session.toml"))
    }

    pub fn load_for_workspace(workspace: &Path) -> Option<Self> {
        let path = Self::path_for_workspace(workspace)?;
        let content = fs::read_to_string(path).ok()?;
        toml::from_str(&content).ok()
    }

    pub fn load() -> Option<Self> {
        if let Ok(current_dir) = std::env::current_dir() {
            if let Some(session) = Self::load_for_workspace(&current_dir) {
                return Some(session);
            }
        }
        let global_path = Self::global_path()?;
        let content = fs::read_to_string(global_path).ok()?;
        toml::from_str(&content).ok()
    }

    pub fn save(&self) -> std::io::Result<()> {
        let target_path = Self::path_for_workspace(&self.workspace).or_else(Self::global_path);
        let Some(path) = target_path else {
            return Ok(());
        };
        if let Some(parent_directory) = path.parent() {
            fs::create_dir_all(parent_directory)?;
        }
        let toml_text = toml::to_string_pretty(self)
            .map_err(|error| std::io::Error::new(std::io::ErrorKind::InvalidData, error))?;
        fs::write(path, toml_text)
    }

    #[must_use]
    pub fn from_workspace(workspace: &Path, files: Vec<PathBuf>, active_index: usize) -> Self {
        Self {
            workspace: workspace.to_path_buf(),
            files,
            active_index,
            ..Default::default()
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn roundtrip_toml() {
        let session = Session {
            workspace: PathBuf::from("/tmp/proj"),
            files: vec![PathBuf::from("/tmp/proj/a.md")],
            active_index: 0,
            scroll_y: 12,
            tree_width: Some(30),
            show_tree: Some(true),
            show_scm: Some(false),
            split: Some(SplitSession {
                orientation: "vertical".into(),
                secondary_file: Some(PathBuf::from("/tmp/proj/b.rs")),
                ratio_percent: 60,
            }),
        };
        let toml_string = toml::to_string(&session).unwrap();
        let parsed_session: Session = toml::from_str(&toml_string).unwrap();
        assert_eq!(parsed_session.workspace, session.workspace);
        assert_eq!(parsed_session.files.len(), 1);
        assert_eq!(parsed_session.scroll_y, 12);
        assert_eq!(parsed_session.tree_width, Some(30));
        assert_eq!(parsed_session.split, session.split);
    }

    #[test]
    fn deterministic_workspace_session_path() {
        let path_alpha = Session::path_for_workspace(Path::new("/tmp/test_dir_alpha"));
        let path_beta = Session::path_for_workspace(Path::new("/tmp/test_dir_beta"));
        assert_ne!(path_alpha, path_beta);
    }

    #[test]
    fn workspace_digest_is_pinned_for_cross_language_parity() {
        // Golden: a implementação Go precisa produzir exatamente estes digests
        // para os mesmos caminhos. Mudar o algoritmo é uma mudança de formato de
        // arquivo — invalida toda sessão já salva, então quebrar este teste é
        // intencional e precisa de migração.
        assert_eq!(
            workspace_digest(Path::new("/tmp/oride")),
            0xD667_B877_1CBE_1AA3
        );
        // Entrada vazia devolve o offset basis: prova que o algoritmo é FNV-1a
        // puro e não alguma variante parecida.
        assert_eq!(workspace_digest(Path::new("")), 0xCBF2_9CE4_8422_2325);
        assert_eq!(
            workspace_digest(Path::new("/tmp/test_dir_alpha")),
            0x980F_B79D_966D_0797
        );
    }

    #[test]
    fn workspace_digest_distinguishes_paths_and_is_stable() {
        let first = workspace_digest(Path::new("/tmp/projeto-a"));
        let second = workspace_digest(Path::new("/tmp/projeto-b"));
        assert_ne!(first, second);
        // Estabilidade: o mesmo caminho sempre produz o mesmo digest.
        assert_eq!(first, workspace_digest(Path::new("/tmp/projeto-a")));
    }
}
