//! Binário `oride` — editor TUI.

use std::env;
use std::path::PathBuf;
use std::process::ExitCode;

use oride_core::DocumentStore;

const VERSION: &str = env!("CARGO_PKG_VERSION");

fn main() -> ExitCode {
    let args: Vec<String> = env::args().skip(1).collect();

    if args.iter().any(|a| a == "-h" || a == "--help") {
        print_help();
        return ExitCode::SUCCESS;
    }
    if args.iter().any(|a| a == "-V" || a == "--version") {
        println!("oride {VERSION}");
        return ExitCode::SUCCESS;
    }
    if args.iter().any(|a| a == "--demo") {
        return run_demo();
    }

    // Subcomando primeiro: `oride conformance …` não é um caminho de arquivo.
    if let Some(first) = args.first() {
        if first == "conformance" {
            return run_conformance(&args[1..]);
        }
    }

    let mut path: Option<PathBuf> = None;
    let mut stat = false;
    let mut force_headless = false;

    for arg in &args {
        match arg.as_str() {
            "--stat" => stat = true,
            "--headless" => force_headless = true,
            other if other.starts_with('-') => {
                eprintln!("unknown argument: {other}");
                print_help();
                return ExitCode::from(2);
            }
            other => path = Some(PathBuf::from(other)),
        }
    }

    if stat {
        let Some(path) = path else {
            eprintln!("error: --stat requires a file path");
            return ExitCode::from(2);
        };
        return match print_file_stat(&path) {
            Ok(()) => ExitCode::SUCCESS,
            Err(err) => {
                eprintln!("error: {err}");
                ExitCode::from(1)
            }
        };
    }

    if force_headless {
        eprintln!(
            "error: --headless requires --stat; \
             for scripted headless runs use `oride conformance run --case <FILE> --workspace <DIR>`"
        );
        return ExitCode::from(2);
    }

    // TUI: path opcional
    match oride_app::run(path) {
        Ok(()) => ExitCode::SUCCESS,
        Err(err) => {
            eprintln!("error: {err:#}");
            ExitCode::from(1)
        }
    }
}

fn print_help() {
    println!(
        "oride {VERSION} — TUI code editor\n\n\
         USAGE:\n\
           oride [path]                     file, directory, or empty (CWD workspace)\n\
           oride --version\n\
           oride --demo                     headless core smoke\n\
           oride <file> --stat              print line/byte stats\n\
           oride conformance run --case <FILE> --workspace <DIR> [--json]\n\n\
         KEYS (see README; rebind in ~/.config/oride/config.toml):\n\
           Ctrl+S save · Ctrl+P open · Ctrl+Shift+P commands\n\
           Ctrl+B tree · Ctrl+` terminal · Ctrl+N/W tabs\n\
         Config: docs/guides/en/config.md · assets/config.example.toml"
    );
}

/// `oride conformance run --case <FILE> --workspace <DIR> [--json]`
///
/// Modo headless determinístico que serve de oráculo ao porte em Go: executa um
/// caso e imprime o estado observável após cada passo. Com `--json` o stdout é
/// reservado ao JSON e todo diagnóstico vai para o stderr, sem ANSI.
fn run_conformance(args: &[String]) -> ExitCode {
    let Some(rest) = args.split_first() else {
        return conformance_usage_error("missing subcommand (expected `run`)");
    };
    match rest.0.as_str() {
        "run" => {}
        other => return conformance_usage_error(&format!("unknown subcommand `{other}`")),
    }

    let mut case_path: Option<PathBuf> = None;
    let mut workspace: Option<PathBuf> = None;
    let mut as_json = false;

    let mut index = 0;
    while index < rest.1.len() {
        let arg = rest.1[index].clone();
        match arg.as_str() {
            "--json" => as_json = true,
            "--case" | "--workspace" => {
                index += 1;
                let Some(value) = rest.1.get(index).cloned() else {
                    return conformance_usage_error(&format!("`{arg}` requires a value"));
                };
                if arg == "--case" {
                    case_path = Some(PathBuf::from(value));
                } else {
                    workspace = Some(PathBuf::from(value));
                }
            }
            other => return conformance_usage_error(&format!("unknown argument `{other}`")),
        }
        index += 1;
    }

    let (Some(case_path), Some(workspace)) = (case_path, workspace) else {
        return conformance_usage_error("both --case and --workspace are required");
    };

    let case = match oride_app::ConformanceCase::from_path(&case_path) {
        Ok(case) => case,
        Err(error) => {
            eprintln!("error: caso `{}`: {error}", case_path.display());
            return ExitCode::from(1);
        }
    };
    let report = match oride_app::run_case(&case, &workspace) {
        Ok(report) => report,
        Err(error) => {
            eprintln!("error: caso `{}`: {error}", case_path.display());
            return ExitCode::from(1);
        }
    };

    if as_json {
        match report.to_json() {
            Ok(json) => {
                println!("{json}");
                ExitCode::SUCCESS
            }
            Err(error) => {
                eprintln!("error: serializando relatório: {error}");
                ExitCode::from(1)
            }
        }
    } else {
        println!(
            "case {}: {} passos, schema {}",
            case_path.display(),
            report.steps,
            report.schema
        );
        for frame in &report.frames {
            println!(
                "  {:>3}. {:<28} buffer {} bytes, {} linhas, caret {:?}",
                frame.after_step,
                frame.step,
                frame.state.document.as_ref().map_or(0, |doc| doc.bytes),
                frame.state.document.as_ref().map_or(0, |doc| doc.lines),
                frame
                    .state
                    .document
                    .as_ref()
                    .and_then(|doc| doc.caret.as_ref())
                    .map(|caret| (caret.line, caret.column))
            );
        }
        ExitCode::SUCCESS
    }
}

fn conformance_usage_error(message: &str) -> ExitCode {
    eprintln!("error: {message}");
    eprintln!("usage: oride conformance run --case <FILE> --workspace <DIR> [--json]");
    ExitCode::from(2)
}

fn run_demo() -> ExitCode {
    let mut store = DocumentStore::new();
    let id = store.open_empty();
    let doc = store.get_mut(id).expect("doc just opened");
    if let Err(err) = doc.insert_text("module demo\n\nfn main() {\n  print(\"hi\")\n}\n") {
        eprintln!("demo insert failed: {err}");
        return ExitCode::from(1);
    }
    doc.commit_edit_group();
    let lines = doc.buffer().line_count();
    let bytes = doc.buffer().len_bytes();
    println!(
        "oride demo ok — {lines} lines, {bytes} bytes, dirty={}",
        doc.is_dirty()
    );
    println!("---");
    print!("{}", doc.buffer().as_string());
    ExitCode::SUCCESS
}

fn print_file_stat(path: &std::path::Path) -> Result<(), oride_core::DocumentError> {
    let mut store = DocumentStore::new();
    let id = store.open_path(path)?;
    let doc = store.get(id).expect("opened");
    println!(
        "{}: {} lines, {} bytes, dirty={}",
        doc.tab_title(),
        doc.buffer().line_count(),
        doc.buffer().len_bytes(),
        doc.is_dirty()
    );
    Ok(())
}
