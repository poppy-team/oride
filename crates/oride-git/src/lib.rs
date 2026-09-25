//! Status git via `git status --porcelain` (sem libgit2).

use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::process::Command;

/// Status simplificado para a árvore.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum GitFileStatus {
    Modified,
    Added,
    Deleted,
    Untracked,
    Renamed,
    Conflict,
}

impl GitFileStatus {
    #[must_use]
    pub fn badge(self) -> char {
        match self {
            Self::Modified => 'M',
            Self::Added => 'A',
            Self::Deleted => 'D',
            Self::Untracked => '?',
            Self::Renamed => 'R',
            Self::Conflict => 'U',
        }
    }
}

/// Mapa path relativo ao root do repo → status (pior status se múltiplos).
pub fn status_map(cwd: &Path) -> HashMap<PathBuf, GitFileStatus> {
    let output = Command::new("git")
        .args(["status", "--porcelain", "-z"])
        .current_dir(cwd)
        .output();

    let Ok(output) = output else {
        return HashMap::new();
    };
    if !output.status.success() {
        return HashMap::new();
    }

    parse_porcelain_z(&output.stdout)
}

pub fn parse_porcelain_z(stdout: &[u8]) -> HashMap<PathBuf, GitFileStatus> {
    let mut map = HashMap::new();
    // -z: records separated by NUL; each "XY path" or rename "XY new_path\0old_path\0"
    let mut iter = stdout.split(|b| *b == 0);
    while let Some(rec) = iter.next() {
        if rec.len() < 3 {
            continue;
        }
        let xy = &rec[..2];
        let path_bytes = &rec[3..]; // skip "XY "
        let path_str = String::from_utf8_lossy(path_bytes);
        let path = PathBuf::from(path_str.trim());
        let status = classify(xy);
        if xy[0] == b'R' || xy[1] == b'R' || xy[0] == b'C' || xy[1] == b'C' {
            // Em rename/copy com -z, o próximo elemento no stream é o orig_path
            let _ = iter.next();
        }
        if path.as_os_str().is_empty() {
            continue;
        }
        map.entry(path)
            .and_modify(|s| *s = worse(*s, status))
            .or_insert(status);
    }
    map
}

/// Blame de uma linha (1-based). Formato curto para status.
pub fn blame_line(cwd: &Path, file: &Path, line_1based: usize) -> Option<String> {
    let rel = file.strip_prefix(cwd).unwrap_or(file);
    let output = Command::new("git")
        .args([
            "blame",
            "-L",
            &format!("{line_1based},{line_1based}"),
            "--porcelain",
            "--",
            &rel.to_string_lossy(),
        ])
        .current_dir(cwd)
        .output()
        .ok()?;
    if !output.status.success() {
        return None;
    }
    let text = String::from_utf8_lossy(&output.stdout);
    let mut author = None;
    let mut summary = None;
    for line in text.lines() {
        if let Some(a) = line.strip_prefix("author ") {
            author = Some(a.to_string());
        }
        if let Some(s) = line.strip_prefix("summary ") {
            summary = Some(s.to_string());
        }
    }
    match (author, summary) {
        (Some(a), Some(s)) => {
            let s: String = s.chars().take(40).collect();
            Some(format!("blame:{a} — {s}"))
        }
        (Some(a), None) => Some(format!("blame:{a}")),
        _ => None,
    }
}

/// Diff unstaged de um path (texto).
pub fn diff_file(cwd: &Path, file: &Path) -> Option<String> {
    let rel = file.strip_prefix(cwd).unwrap_or(file);
    let output = Command::new("git")
        .args(["diff", "--", &rel.to_string_lossy()])
        .current_dir(cwd)
        .output()
        .ok()?;
    // untracked: empty diff — try as new file
    let mut text = String::from_utf8_lossy(&output.stdout).to_string();
    if text.trim().is_empty() {
        let out2 = Command::new("git")
            .args([
                "diff",
                "--no-index",
                oride_osutil::null_device(),
                &rel.to_string_lossy(),
            ])
            .current_dir(cwd)
            .output()
            .ok()?;
        text = String::from_utf8_lossy(&out2.stdout).to_string();
    }
    if text.trim().is_empty() {
        None
    } else {
        Some(text)
    }
}

/// Lista ordenada (badge, path) para painel SCM.
pub fn scm_entries(cwd: &Path) -> Vec<(GitFileStatus, PathBuf)> {
    let map = status_map(cwd);
    let mut v: Vec<_> = map.into_iter().map(|(p, s)| (s, p)).collect();
    v.sort_by(|a, b| a.1.cmp(&b.1));
    v
}

/// Branch atual ou `None`.
pub fn current_branch(cwd: &Path) -> Option<String> {
    let output = Command::new("git")
        .args(["rev-parse", "--abbrev-ref", "HEAD"])
        .current_dir(cwd)
        .output()
        .ok()?;
    if !output.status.success() {
        return None;
    }
    let s = String::from_utf8_lossy(&output.stdout).trim().to_string();
    if s.is_empty() {
        None
    } else {
        Some(s)
    }
}

fn classify(xy: &[u8]) -> GitFileStatus {
    let x = xy[0] as char;
    let y = xy[1] as char;
    if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
        return GitFileStatus::Conflict;
    }
    if x == '?' || y == '?' {
        return GitFileStatus::Untracked;
    }
    if x == 'R' || y == 'R' {
        return GitFileStatus::Renamed;
    }
    if x == 'A' || y == 'A' {
        return GitFileStatus::Added;
    }
    if x == 'D' || y == 'D' {
        return GitFileStatus::Deleted;
    }
    if x == 'M' || y == 'M' {
        return GitFileStatus::Modified;
    }
    GitFileStatus::Modified
}

fn worse(a: GitFileStatus, b: GitFileStatus) -> GitFileStatus {
    use GitFileStatus::*;
    let rank = |s: GitFileStatus| match s {
        Conflict => 5,
        Deleted => 4,
        Modified => 3,
        Renamed => 2,
        Added => 1,
        Untracked => 0,
    };
    if rank(b) > rank(a) {
        b
    } else {
        a
    }
}

/// Erros em operações interativas do git.
#[derive(Debug, thiserror::Error)]
pub enum GitError {
    #[error("git falhou ({code:?}): {message}")]
    CommandFailed { code: Option<i32>, message: String },
    #[error("erro de I/O no git: {0}")]
    Io(#[from] std::io::Error),
    #[error("mensagem de commit não pode ser vazia")]
    EmptyCommitMessage,
}

/// Adiciona arquivo ao stage (`git add -- <path>`).
pub fn stage_path(cwd: &Path, path: &Path) -> Result<(), GitError> {
    let rel = path.strip_prefix(cwd).unwrap_or(path);
    let output = Command::new("git")
        .args(["add", "--", &rel.to_string_lossy()])
        .current_dir(cwd)
        .output()?;
    if !output.status.success() {
        return Err(GitError::CommandFailed {
            code: output.status.code(),
            message: String::from_utf8_lossy(&output.stderr).trim().to_string(),
        });
    }
    Ok(())
}

/// Remove arquivo do stage (`git restore --staged -- <path>` ou `git rm --cached`).
pub fn unstage_path(cwd: &Path, path: &Path) -> Result<(), GitError> {
    let rel = path.strip_prefix(cwd).unwrap_or(path);
    let output = Command::new("git")
        .args(["restore", "--staged", "--", &rel.to_string_lossy()])
        .current_dir(cwd)
        .output()?;
    if !output.status.success() {
        // Se HEAD não puder ser resolvido (repo vazio), faz fallback para git rm --cached
        let fallback = Command::new("git")
            .args(["rm", "--cached", "--", &rel.to_string_lossy()])
            .current_dir(cwd)
            .output()?;
        if !fallback.status.success() {
            return Err(GitError::CommandFailed {
                code: output.status.code(),
                message: String::from_utf8_lossy(&output.stderr).trim().to_string(),
            });
        }
    }
    Ok(())
}

/// Executa commit com a mensagem fornecida (`git commit -m <msg>`).
pub fn commit(cwd: &Path, message: &str) -> Result<String, GitError> {
    let trimmed = message.trim();
    if trimmed.is_empty() {
        return Err(GitError::EmptyCommitMessage);
    }
    let output = Command::new("git")
        .args(["commit", "-m", trimmed])
        .current_dir(cwd)
        .output()?;
    if !output.status.success() {
        return Err(GitError::CommandFailed {
            code: output.status.code(),
            message: String::from_utf8_lossy(&output.stderr).trim().to_string(),
        });
    }
    Ok(String::from_utf8_lossy(&output.stdout).trim().to_string())
}

/// Calcula commits à frente e atrás em relação à branch de rastreamento upstream (`ahead`, `behind`).
/// Executa `git rev-list --left-right --count HEAD...@{upstream}`.
pub fn ahead_behind(cwd: &Path) -> Option<(usize, usize)> {
    let output = Command::new("git")
        .args(["rev-list", "--left-right", "--count", "HEAD...@{upstream}"])
        .current_dir(cwd)
        .output()
        .ok()?;
    if !output.status.success() {
        return None;
    }
    let text = String::from_utf8_lossy(&output.stdout);
    parse_ahead_behind(text.trim())
}

#[must_use]
pub fn parse_ahead_behind(raw: &str) -> Option<(usize, usize)> {
    let mut parts = raw.split_whitespace();
    let ahead = parts.next()?.parse::<usize>().ok()?;
    let behind = parts.next()?.parse::<usize>().ok()?;
    Some((ahead, behind))
}

/// Formata a contagem de ahead/behind para exibição compacta (ex.: `↑1 ↓2`, `↑3`, `↓1`).
#[must_use]
pub fn format_ahead_behind(ahead: usize, behind: usize) -> Option<String> {
    match (ahead, behind) {
        (0, 0) => None,
        (a, 0) => Some(format!("↑{a}")),
        (0, b) => Some(format!("↓{b}")),
        (a, b) => Some(format!("↑{a} ↓{b}")),
    }
}

/// Executa pull da branch remota (`git pull`).
pub fn git_pull(cwd: &Path) -> Result<String, GitError> {
    let output = Command::new("git")
        .args(["pull"])
        .current_dir(cwd)
        .output()?;
    if !output.status.success() {
        let err = String::from_utf8_lossy(&output.stderr).trim().to_string();
        let message = if err.is_empty() {
            String::from_utf8_lossy(&output.stdout).trim().to_string()
        } else {
            err
        };
        return Err(GitError::CommandFailed {
            code: output.status.code(),
            message,
        });
    }
    let stdout = String::from_utf8_lossy(&output.stdout).trim().to_string();
    if stdout.is_empty() {
        Ok("Already up to date.".to_string())
    } else {
        Ok(stdout)
    }
}

/// Executa push para a branch remota (`git push`).
pub fn git_push(cwd: &Path) -> Result<String, GitError> {
    let output = Command::new("git")
        .args(["push"])
        .current_dir(cwd)
        .output()?;
    if !output.status.success() {
        let err = String::from_utf8_lossy(&output.stderr).trim().to_string();
        let message = if err.is_empty() {
            String::from_utf8_lossy(&output.stdout).trim().to_string()
        } else {
            err
        };
        return Err(GitError::CommandFailed {
            code: output.status.code(),
            message,
        });
    }
    let stdout = String::from_utf8_lossy(&output.stdout).trim().to_string();
    let stderr = String::from_utf8_lossy(&output.stderr).trim().to_string();
    if !stdout.is_empty() {
        Ok(stdout)
    } else if !stderr.is_empty() {
        Ok(stderr)
    } else {
        Ok("pushed successfully".to_string())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn classify_modified() {
        assert_eq!(classify(b" M"), GitFileStatus::Modified);
        assert_eq!(classify(b"??"), GitFileStatus::Untracked);
        assert_eq!(classify(b"A "), GitFileStatus::Added);
    }

    #[test]
    fn parses_renamed_file_without_phantom_entry() {
        // "R  new_name.rs\0old_name.rs\0 M modified.rs\0"
        let data = b"R  new_name.rs\0old_name.rs\0 M modified.rs\0";
        let map = parse_porcelain_z(data);
        assert_eq!(map.len(), 2);
        assert_eq!(
            map.get(&PathBuf::from("new_name.rs")),
            Some(&GitFileStatus::Renamed)
        );
        assert_eq!(
            map.get(&PathBuf::from("modified.rs")),
            Some(&GitFileStatus::Modified)
        );
        assert_eq!(map.get(&PathBuf::from("old_name.rs")), None);
    }

    #[test]
    fn stage_unstage_and_commit_flow() {
        let temp_dir = tempfile::tempdir().unwrap();
        let path = temp_dir.path();
        // Inicializa repo git temporário
        let _ = Command::new("git")
            .args(["init"])
            .current_dir(path)
            .output()
            .unwrap();
        let _ = Command::new("git")
            .args(["config", "user.name", "Test"])
            .current_dir(path)
            .output()
            .unwrap();
        let _ = Command::new("git")
            .args(["config", "user.email", "test@test.com"])
            .current_dir(path)
            .output()
            .unwrap();

        let file = path.join("file.txt");
        std::fs::write(&file, "hello").unwrap();

        // Stage
        stage_path(path, &PathBuf::from("file.txt")).unwrap();
        let entries = scm_entries(path);
        assert!(entries
            .iter()
            .any(|(s, p)| *s == GitFileStatus::Added && p == Path::new("file.txt")));

        // Unstage
        unstage_path(path, &PathBuf::from("file.txt")).unwrap();
        let entries = scm_entries(path);
        assert!(entries
            .iter()
            .any(|(s, p)| *s == GitFileStatus::Untracked && p == Path::new("file.txt")));

        // Stage novamente e commit
        stage_path(path, &PathBuf::from("file.txt")).unwrap();
        let commit_res = commit(path, "initial commit").unwrap();
        assert!(commit_res.contains("initial commit") || !commit_res.is_empty());
        assert!(scm_entries(path).is_empty());
    }

    #[test]
    fn parse_and_format_ahead_behind() {
        assert_eq!(parse_ahead_behind("1\t2"), Some((1, 2)));
        assert_eq!(parse_ahead_behind("0 0"), Some((0, 0)));
        assert_eq!(parse_ahead_behind("4 0"), Some((4, 0)));
        assert_eq!(parse_ahead_behind("invalid"), None);

        assert_eq!(format_ahead_behind(1, 2), Some("↑1 ↓2".into()));
        assert_eq!(format_ahead_behind(3, 0), Some("↑3".into()));
        assert_eq!(format_ahead_behind(0, 5), Some("↓5".into()));
        assert_eq!(format_ahead_behind(0, 0), None);
    }
}
