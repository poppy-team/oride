//! Painel de terminal embutido (PTY) — shell interativo.

use std::io::{Read, Write};
use std::path::Path;
use std::sync::mpsc::{self, Receiver, TryRecvError};
use std::thread;
use std::time::Duration;

use portable_pty::{CommandBuilder, NativePtySystem, PtySize, PtySystem};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum TerminalError {
    #[error("pty: {0}")]
    Pty(String),
    #[error("io: {0}")]
    Io(#[from] std::io::Error),
}

/// Terminal embutido com scrollback (ANSI stripped para o painel texto).
pub struct EmbeddedTerminal {
    writer: Box<dyn Write + Send>,
    rx: Receiver<Vec<u8>>,
    scrollback: String,
    cursor_col: usize,
    pub visible: bool,
    pub height_lines: u16,
    pub last_error: Option<String>,
    _child: Box<dyn portable_pty::Child + Send + Sync>,
    _master: Box<dyn portable_pty::MasterPty + Send>,
}

impl EmbeddedTerminal {
    /// Spawna shell **interativo** em `cwd`.
    pub fn spawn(
        cwd: &Path,
        cols: u16,
        rows: u16,
        configured_shell: Option<&str>,
    ) -> Result<Self, TerminalError> {
        let pty_system = NativePtySystem::default();
        let pair = pty_system
            .openpty(PtySize {
                rows: rows.max(2),
                cols: cols.max(20),
                pixel_width: 0,
                pixel_height: 0,
            })
            .map_err(|e| TerminalError::Pty(e.to_string()))?;

        let shell = configured_shell
            .filter(|shell| !shell.trim().is_empty())
            .map(str::to_owned)
            .or_else(|| std::env::var("SHELL").ok());
        let mut cmd = match shell.as_deref() {
            Some(shell) => CommandBuilder::new(shell),
            None => CommandBuilder::new_default_prog(),
        };
        if let Some(shell) = shell.as_deref() {
            if shell.contains("bash")
                || shell.contains("zsh")
                || shell.contains("fish")
                || shell.ends_with("/sh")
            {
                cmd.arg("-i");
            }
        }
        cmd.cwd(cwd);
        cmd.env("TERM", "xterm-256color");
        cmd.env("COLORTERM", "truecolor");

        let child = pair
            .slave
            .spawn_command(cmd)
            .map_err(|e| TerminalError::Pty(e.to_string()))?;

        let mut reader = pair
            .master
            .try_clone_reader()
            .map_err(|e| TerminalError::Pty(e.to_string()))?;
        let writer = pair
            .master
            .take_writer()
            .map_err(|e| TerminalError::Pty(e.to_string()))?;

        let (tx, rx) = mpsc::channel::<Vec<u8>>();
        thread::spawn(move || {
            let mut buf = [0u8; 8192];
            loop {
                match reader.read(&mut buf) {
                    Ok(0) => break,
                    Ok(n) => {
                        if tx.send(buf[..n].to_vec()).is_err() {
                            break;
                        }
                    }
                    Err(error)
                        if matches!(
                            error.kind(),
                            std::io::ErrorKind::Interrupted | std::io::ErrorKind::WouldBlock
                        ) =>
                    {
                        thread::sleep(Duration::from_millis(15));
                    }
                    Err(_) => break,
                }
            }
        });

        Ok(Self {
            writer,
            rx,
            scrollback: String::new(),
            cursor_col: 0,
            visible: false,
            height_lines: 12,
            last_error: None,
            _child: child,
            _master: pair.master,
        })
    }

    pub fn poll_output(&mut self) {
        loop {
            match self.rx.try_recv() {
                Ok(chunk) => {
                    let text = String::from_utf8_lossy(&chunk);
                    apply_terminal_chunk(&mut self.scrollback, &mut self.cursor_col, &text);
                    trim_scrollback(&mut self.scrollback, 400_000);
                }
                Err(TryRecvError::Empty) => break,
                Err(TryRecvError::Disconnected) => {
                    self.last_error = Some("shell encerrado".into());
                    break;
                }
            }
        }
    }

    #[must_use]
    pub fn cursor_col(&self) -> usize {
        self.cursor_col
    }

    pub fn write_bytes(&mut self, data: &[u8]) -> Result<(), TerminalError> {
        match self
            .writer
            .write_all(data)
            .and_then(|_| self.writer.flush())
        {
            Ok(()) => {
                self.last_error = None;
                Ok(())
            }
            Err(e) => {
                self.last_error = Some(e.to_string());
                Err(TerminalError::Io(e))
            }
        }
    }

    pub fn write_str(&mut self, s: &str) -> Result<(), TerminalError> {
        self.write_bytes(s.as_bytes())
    }

    #[must_use]
    pub fn visible_lines(&self, max_lines: usize) -> Vec<String> {
        let lines: Vec<&str> = self.scrollback.lines().collect();
        let start = lines.len().saturating_sub(max_lines.max(1));
        lines[start..].iter().map(|s| (*s).to_string()).collect()
    }

    pub fn toggle_visible(&mut self) {
        self.visible = !self.visible;
    }

    pub fn resize(&mut self, cols: u16, rows: u16) {
        let _ = self._master.resize(PtySize {
            rows: rows.max(2),
            cols: cols.max(20),
            pixel_width: 0,
            pixel_height: 0,
        });
    }

    pub fn grow(&mut self, delta: u16) {
        self.height_lines = self.height_lines.saturating_add(delta).clamp(3, 40);
        self.visible = true;
    }

    pub fn shrink(&mut self, delta: u16) {
        self.height_lines = self.height_lines.saturating_sub(delta).max(3);
    }
}

pub(crate) fn apply_terminal_chunk(scrollback: &mut String, cursor_col: &mut usize, text: &str) {
    if text.is_empty() {
        return;
    }

    // Separa o prefixo de linhas já finalizadas e a linha atual em buffer
    let (prefix, current) = match scrollback.rfind('\n') {
        Some(idx) => (&scrollback[..=idx], &scrollback[idx + 1..]),
        None => ("", scrollback.as_str()),
    };

    let mut out_prefix = prefix.to_string();
    let mut line_chars: Vec<char> = current.chars().collect();
    let mut col: usize = (*cursor_col).min(line_chars.len());

    let mut chars = text.chars().peekable();
    while let Some(c) = chars.next() {
        match c {
            '\r' => {
                if chars.peek() == Some(&'\n') {
                    chars.next();
                    out_prefix.extend(line_chars.drain(..));
                    out_prefix.push('\n');
                    col = 0;
                } else {
                    col = 0;
                }
            }
            '\n' => {
                out_prefix.extend(line_chars.drain(..));
                out_prefix.push('\n');
                col = 0;
            }
            '\x08' => {
                col = col.saturating_sub(1);
            }
            '\t' => {
                let next_tab = (col + 8) & !7;
                while col < next_tab {
                    if col < line_chars.len() {
                        line_chars[col] = ' ';
                    } else {
                        line_chars.push(' ');
                    }
                    col += 1;
                }
            }
            '\x1b' => match chars.peek() {
                Some(']') => {
                    // Sequência OSC: \x1b] ... (\x07 | \x1b\)
                    chars.next();
                    while let Some(ch) = chars.next() {
                        if ch == '\x07' {
                            break;
                        }
                        if ch == '\x1b' && chars.peek() == Some(&'\\') {
                            chars.next();
                            break;
                        }
                    }
                }
                Some('[') => {
                    // Sequência CSI: \x1b[ [params] [cmd]
                    chars.next();
                    let mut param_str = String::new();
                    let mut cmd = None;
                    while let Some(&ch) = chars.peek() {
                        if ch.is_ascii_digit() || ch == ';' || ch == '?' {
                            param_str.push(ch);
                            chars.next();
                        } else {
                            cmd = chars.next();
                            break;
                        }
                    }
                    if let Some(cmd) = cmd {
                        match cmd {
                            'K' => {
                                // Limpar linha
                                let param = param_str.parse::<usize>().unwrap_or(0);
                                if param == 0 {
                                    line_chars.truncate(col);
                                } else if param == 1 {
                                    for ch in line_chars.iter_mut().take(col + 1) {
                                        *ch = ' ';
                                    }
                                } else if param == 2 {
                                    line_chars.clear();
                                    col = 0;
                                }
                            }
                            'J' => {
                                // Limpar display
                                let param = param_str.parse::<usize>().unwrap_or(0);
                                if param == 0 {
                                    line_chars.truncate(col);
                                } else if param == 2 {
                                    line_chars.clear();
                                    col = 0;
                                }
                            }
                            'D' => {
                                // Cursor para trás
                                let count = param_str.parse::<usize>().unwrap_or(1).max(1);
                                col = col.saturating_sub(count);
                            }
                            'C' => {
                                // Cursor para frente
                                let count = param_str.parse::<usize>().unwrap_or(1).max(1);
                                col = col.saturating_add(count);
                            }
                            'G' | '`' => {
                                // Cursor horizontal absoluto
                                let target = param_str.parse::<usize>().unwrap_or(1).max(1);
                                col = target.saturating_sub(1);
                            }
                            'H' | 'f' => {
                                // Posição do cursor \x1b[row;colH
                                if let Some((_, c_str)) = param_str.split_once(';') {
                                    let target = c_str.parse::<usize>().unwrap_or(1).max(1);
                                    col = target.saturating_sub(1);
                                } else if !param_str.is_empty() {
                                    let target = param_str.parse::<usize>().unwrap_or(1).max(1);
                                    col = target.saturating_sub(1);
                                }
                            }
                            _ => {
                                // Outros comandos CSI (SGR cores 'm', modos '?2004h', etc.) são ignorados
                            }
                        }
                    }
                }
                Some('=' | '>' | '<') => {
                    chars.next();
                }
                Some('(' | ')') => {
                    chars.next();
                    let _ = chars.next();
                }
                _ => {}
            },
            ch if (ch as u32) < 32 => {
                // Outros caracteres de controle ASCII são ignorados
            }
            ch => {
                if col < line_chars.len() {
                    line_chars[col] = ch;
                } else {
                    while line_chars.len() < col {
                        line_chars.push(' ');
                    }
                    line_chars.push(ch);
                }
                col += 1;
            }
        }
    }

    out_prefix.extend(line_chars);
    *scrollback = out_prefix;
    *cursor_col = col;
}

fn trim_scrollback(scrollback: &mut String, max_bytes: usize) {
    if scrollback.len() <= max_bytes {
        return;
    }

    let mut drain_end = scrollback.len() - max_bytes;
    while drain_end < scrollback.len() && !scrollback.is_char_boundary(drain_end) {
        drain_end += 1;
    }
    scrollback.drain(..drain_end);
}

#[cfg(test)]
mod tests {
    use super::trim_scrollback;

    #[test]
    fn trimming_scrollback_preserves_utf8_boundaries() {
        let mut scrollback = "a✨b✨c".to_string();

        trim_scrollback(&mut scrollback, 6);

        assert!(scrollback.len() <= 6);
        assert_eq!(scrollback, "b✨c");
    }

    #[test]
    fn test_apply_terminal_chunk_carriage_return_and_backspace() {
        use super::apply_terminal_chunk;

        let mut scrollback = String::new();
        let mut col = 0;
        // Simula progress bar ou prompt zsh
        apply_terminal_chunk(
            &mut scrollback,
            &mut col,
            "loading 10%\rloading 50%\rloading 100%\n",
        );
        assert_eq!(scrollback, "loading 100%\n");

        // Simula digitação com backspace
        apply_terminal_chunk(&mut scrollback, &mut col, "sh-5.2$ el\x08cho oi\r\n");
        assert_eq!(scrollback, "loading 100%\nsh-5.2$ echo oi\n");

        // Simula sobrescrita de prompt zsh com espaços
        apply_terminal_chunk(&mut scrollback, &mut col, "%        \r \ruser@arch$ cargo");
        assert_eq!(
            scrollback,
            "loading 100%\nsh-5.2$ echo oi\nuser@arch$ cargo"
        );
    }

    #[test]
    fn test_apply_terminal_chunk_ansi_csi_zsh() {
        use super::apply_terminal_chunk;

        let mut scrollback = String::new();
        let mut col = 0;
        // Prompt com EOL marker do zsh e limpeza de linha via CSI J/K
        let chunk1 = "%\x1b[0m               \r \r\x1b[Juser@arch:~$ \x1b[K";
        apply_terminal_chunk(&mut scrollback, &mut col, chunk1);
        assert_eq!(scrollback, "user@arch:~$ ");
        assert_eq!(col, 13);

        // Digitação com cores/syntax highlight zsh
        let chunk2 = "echo\x1b[4D\x1b[32me\x1b[32mc\x1b[32mh\x1b[32mo\x1b[39m";
        apply_terminal_chunk(&mut scrollback, &mut col, chunk2);
        assert_eq!(scrollback, "user@arch:~$ echo");
        assert_eq!(col, 17);
    }

    #[test]
    fn test_spawn_and_write_echo() {
        let shell = if std::path::Path::new("/usr/bin/zsh").exists() {
            Some("/usr/bin/zsh")
        } else {
            Some(oride_osutil::default_shell())
        };
        let mut term = super::EmbeddedTerminal::spawn(std::path::Path::new("."), 80, 24, shell)
            .expect("spawn terminal");
        std::thread::sleep(std::time::Duration::from_millis(300));
        term.poll_output();
        println!("INITIAL SCROLLBACK: {:?}", term.scrollback);
        println!("INITIAL VISIBLE LINES: {:?}", term.visible_lines(20));

        term.write_str("echo HELLO_ORIDE\r").expect("write str");
        std::thread::sleep(std::time::Duration::from_millis(300));
        term.poll_output();
        let lines = term.visible_lines(20);
        println!("LINES AFTER ECHO: {:?}", lines);
        println!("SCROLLBACK AFTER ECHO: {:?}", term.scrollback);
        assert!(
            lines.iter().any(|l| l.contains("HELLO_ORIDE")),
            "Deve conter HELLO_ORIDE, mas obteve: {:?}",
            lines
        );
    }
}
