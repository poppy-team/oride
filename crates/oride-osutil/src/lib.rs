//! Costuras de plataforma.
//!
//! Cada função aqui existe porque o mesmo conceito tem nome diferente em cada
//! sistema operacional. Espalhar `#[cfg]` pelos crates consumidores transforma uma
//! mudança de plataforma em uma caça; concentrar num crate a transforma numa
//! edição só.
//!
//! Tudo falha com valor, nunca com panic: o editor continua utilizável num sistema
//! onde um recurso não existe.

use std::path::{Path, PathBuf};
use std::process::Command;

/// O dispositivo nulo do sistema (`/dev/null`, `NUL`).
///
/// Um literal `/dev/null` passado a um subprocesso no Windows é um caminho
/// relativo de nome estranho, e a ferramenta falha por um motivo que não aparece
/// na mensagem de erro.
#[must_use]
pub fn null_device() -> &'static str {
    if cfg!(windows) {
        "NUL"
    } else {
        "/dev/null"
    }
}

/// O shell do sistema, com os argumentos que o fazem executar um comando.
///
/// `sh -c` não existe no Windows, onde o interpretador é `cmd /C`. Devolver a
/// dupla (programa, argumentos) mantém a construção do comando fora do `cfg`.
#[must_use]
pub fn shell_for_command() -> (&'static str, &'static str) {
    if cfg!(windows) {
        ("cmd", "/C")
    } else {
        ("sh", "-c")
    }
}

/// Monta um `Command` que executa `command` no shell do sistema.
///
/// O texto vai como **um** argumento, nunca interpolado numa linha de shell
/// construída à mão — assim um comando com aspas ou `;` não vira outra coisa.
#[must_use]
pub fn shell_command(command: &str) -> Command {
    let (program, flag) = shell_for_command();
    let mut cmd = Command::new(program);
    cmd.arg(flag).arg(command);
    cmd
}

/// O shell padrão quando a configuração e `$SHELL` estão vazios.
#[must_use]
pub fn default_shell() -> &'static str {
    if cfg!(windows) {
        "cmd"
    } else {
        "/bin/sh"
    }
}

/// Abre um alvo (URL ou caminho) com o aplicativo padrão do sistema.
///
/// O processo é destacado e suas saídas descartadas: um navegador que imprime no
/// terminal do editor corromperia a tela.
pub fn open_target(target: &str) -> std::io::Result<()> {
    #[cfg(target_os = "macos")]
    let mut cmd = Command::new("open");
    #[cfg(target_os = "windows")]
    let mut cmd = {
        let mut c = Command::new("cmd");
        // O primeiro argumento vazio é o título da janela; sem ele, `start`
        // trataria o alvo como título e não abriria nada.
        c.args(["/C", "start", ""]);
        c
    };
    #[cfg(not(any(target_os = "macos", target_os = "windows")))]
    let mut cmd = Command::new("xdg-open");

    cmd.arg(target);
    cmd.stdin(std::process::Stdio::null());
    cmd.stdout(std::process::Stdio::null());
    cmd.stderr(std::process::Stdio::null());
    cmd.spawn()?;
    Ok(())
}

/// As extensões que tornam um arquivo executável neste sistema.
///
/// No Unix a permissão já responde; no Windows um executável tem sufixo, e
/// procurar por `rg` sem `.exe` não encontra nada.
#[must_use]
pub fn executable_extensions() -> &'static [&'static str] {
    if cfg!(windows) {
        &["", ".exe", ".cmd", ".bat", ".com"]
    } else {
        &[""]
    }
}

/// Procura um executável no `$PATH`.
///
/// Resolve o sufixo por plataforma e, no Unix, confere a permissão de execução —
/// um diretório com um `rg` sem permissão não é um `rg` utilizável.
#[must_use]
pub fn find_in_path(name: &str) -> Option<PathBuf> {
    if name.is_empty() || name.contains(std::path::MAIN_SEPARATOR) {
        return None;
    }

    let path_var = std::env::var_os("PATH")?;
    for dir in std::env::split_paths(&path_var) {
        for suffix in executable_extensions() {
            let candidate = dir.join(format!("{name}{suffix}"));
            if is_executable_file(&candidate) {
                return Some(candidate);
            }
        }
    }
    None
}

/// Um caminho é um arquivo executável?
#[must_use]
pub fn is_executable_file(path: &Path) -> bool {
    let Ok(metadata) = std::fs::metadata(path) else {
        return false;
    };
    if !metadata.is_file() {
        return false;
    }

    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        metadata.permissions().mode() & 0o111 != 0
    }
    #[cfg(not(unix))]
    {
        true
    }
}

#[cfg(test)]
mod tests {
    use std::path::Path;

    use super::{
        default_shell, executable_extensions, find_in_path, is_executable_file, null_device,
        open_target, shell_command, shell_for_command,
    };

    #[test]
    fn null_device_is_absolute_or_the_windows_token() {
        let device = null_device();
        assert!(
            device == "NUL" || device.starts_with('/'),
            "device: {device}"
        );
    }

    #[test]
    fn shell_command_passes_the_text_as_one_argument() {
        let cmd = shell_command("echo a; echo b");
        let args: Vec<_> = cmd.get_args().collect();

        // O comando inteiro é um argumento. Interpolar numa linha de shell
        // montada à mão é como um `;` vira uma segunda instrução.
        assert_eq!(args.len(), 2);
        assert_eq!(args[1], "echo a; echo b");
    }

    #[test]
    fn shell_for_command_is_a_program_and_a_flag() {
        let (program, flag) = shell_for_command();
        assert!(!program.is_empty());
        assert!(flag.starts_with('-') || flag.starts_with('/'));
    }

    #[test]
    fn default_shell_is_not_empty() {
        assert!(!default_shell().is_empty());
    }

    #[test]
    fn find_in_path_locates_a_shell_that_must_exist() {
        // `sh` existe em qualquer Unix, e o teste roda onde o editor roda.
        if cfg!(unix) {
            assert!(find_in_path("sh").is_some());
        }
    }

    #[test]
    fn find_in_path_rejects_empty_and_paths() {
        assert_eq!(find_in_path(""), None);
        assert_eq!(find_in_path("/usr/bin/sh"), None);
    }

    #[test]
    fn find_in_path_returns_nothing_for_a_missing_binary() {
        assert_eq!(find_in_path("oride-nao-existe-este-binario"), None);
    }

    #[test]
    fn is_executable_file_rejects_directories_and_missing_paths() {
        assert!(!is_executable_file(Path::new("/")));
        assert!(!is_executable_file(Path::new("/nao/existe/oride")));
    }

    #[test]
    fn executable_extensions_are_never_empty() {
        assert!(!executable_extensions().is_empty());
    }

    #[test]
    fn open_target_on_a_missing_helper_is_a_value_not_a_panic() {
        // Falhar é aceitável; derrubar o editor não é.
        let _ = open_target("");
    }
}
