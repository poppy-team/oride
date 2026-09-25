//! Cliente LSP com thread de leitura.

use std::collections::{HashMap, HashSet};
use std::io::BufReader;
use std::path::{Path, PathBuf};
use std::process::{Child, ChildStdin, Command, Stdio};
use std::sync::mpsc::{self, Receiver, Sender, TryRecvError};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::{Duration, Instant};

use serde_json::{json, Value};
use thiserror::Error;

use crate::protocol::{notification, path_to_uri, read_message, request, write_message};
use crate::types::{
    character_column, CompletionItem, Diagnostic, DiagnosticSeverity, HoverInfo, Location,
    Position, Range,
};

#[derive(Debug, Error)]
pub enum LspError {
    #[error("io: {0}")]
    Io(#[from] std::io::Error),
    #[error("protocol: {0}")]
    Protocol(String),
    #[error("lsp not running")]
    NotRunning,
    #[error("timeout waiting for response")]
    Timeout,
    #[error("server error: {0}")]
    Server(String),
}

/// Eventos assíncronos do servidor.
#[derive(Debug, Clone)]
pub enum LspEvent {
    Diagnostics {
        uri: String,
        diagnostics: Vec<Diagnostic>,
    },
    ServerMessage(String),
    Exited,
}

struct Pending {
    tx: Sender<Result<Value, LspError>>,
}

/// Cliente LSP stdio.
pub struct LspClient {
    child: Option<Child>,
    stdin: Option<ChildStdin>,
    next_id: u64,
    pending: Arc<Mutex<HashMap<u64, Pending>>>,
    events: Receiver<LspEvent>,
    _reader: Option<thread::JoinHandle<()>>,
    root: PathBuf,
    timeout: Duration,
    open_documents: HashSet<String>,
    pub ready: bool,
    pub last_error: Option<String>,
}

impl LspClient {
    /// Spawna o comando (ex.: `rust-analyzer` ou `clangd`) com cwd = root.
    pub fn spawn(
        command: &[String],
        root: impl AsRef<Path>,
        timeout_ms: u64,
    ) -> Result<Self, LspError> {
        if command.is_empty() {
            return Err(LspError::Protocol("empty lsp command".into()));
        }
        let root = root.as_ref().to_path_buf();
        let mut cmd = Command::new(&command[0]);
        if command.len() > 1 {
            cmd.args(&command[1..]);
        }
        cmd.current_dir(&root)
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .stderr(Stdio::null());

        let mut child = cmd.spawn()?;
        let stdin = child
            .stdin
            .take()
            .ok_or_else(|| LspError::Protocol("no stdin".into()))?;
        let stdout = child
            .stdout
            .take()
            .ok_or_else(|| LspError::Protocol("no stdout".into()))?;

        let pending: Arc<Mutex<HashMap<u64, Pending>>> = Arc::new(Mutex::new(HashMap::new()));
        let pending_r = Arc::clone(&pending);
        let (ev_tx, ev_rx) = mpsc::channel::<LspEvent>();

        let reader = thread::spawn(move || {
            let mut reader = BufReader::new(stdout);
            loop {
                match read_message(&mut reader) {
                    Ok(msg) => {
                        if let Some(id) = msg.get("id").and_then(|v| v.as_u64()) {
                            let result = if let Some(err) = msg.get("error") {
                                Err(LspError::Server(err.to_string()))
                            } else {
                                Ok(msg.get("result").cloned().unwrap_or(Value::Null))
                            };
                            if let Ok(mut map) = pending_r.lock() {
                                if let Some(p) = map.remove(&id) {
                                    let _ = p.tx.send(result);
                                }
                            }
                        } else if let Some(method) = msg.get("method").and_then(|m| m.as_str()) {
                            if method == "textDocument/publishDiagnostics" {
                                if let Some(params) = msg.get("params") {
                                    let uri = params
                                        .get("uri")
                                        .and_then(|u| u.as_str())
                                        .unwrap_or("")
                                        .to_string();
                                    let diags = parse_diagnostics(params);
                                    let _ = ev_tx.send(LspEvent::Diagnostics {
                                        uri,
                                        diagnostics: diags,
                                    });
                                }
                            }
                        }
                    }
                    Err(_) => {
                        if let Ok(mut map) = pending_r.lock() {
                            for (_, pending) in map.drain() {
                                let _ = pending.tx.send(Err(LspError::NotRunning));
                            }
                        }
                        let _ = ev_tx.send(LspEvent::Exited);
                        break;
                    }
                }
            }
        });

        let mut client = Self {
            child: Some(child),
            stdin: Some(stdin),
            next_id: 1,
            pending,
            events: ev_rx,
            _reader: Some(reader),
            root,
            timeout: Duration::from_millis(timeout_ms.max(500)),
            open_documents: HashSet::new(),
            ready: false,
            last_error: None,
        };
        client.initialize()?;
        Ok(client)
    }

    fn initialize(&mut self) -> Result<(), LspError> {
        let root_uri = path_to_uri(&self.root);
        let params = json!({
            "processId": std::process::id(),
            "rootUri": root_uri,
            "rootPath": self.root.display().to_string(),
            "capabilities": {
                "textDocument": {
                    "synchronization": { "didSave": true, "dynamicRegistration": false },
                    "completion": { "completionItem": { "snippetSupport": false } },
                    "hover": { "contentFormat": ["plaintext", "markdown"] },
                    "definition": { "linkSupport": false },
                    "publishDiagnostics": { "relatedInformation": false },
                    "formatting": { "dynamicRegistration": false }
                },
                "workspace": { "workspaceFolders": false }
            },
            "clientInfo": { "name": "oride", "version": "0.1.0" }
        });
        let _ = self.request("initialize", params)?;
        self.notify("initialized", json!({}))?;
        self.ready = true;
        Ok(())
    }

    fn next_id(&mut self) -> u64 {
        let id = self.next_id;
        self.next_id += 1;
        id
    }

    fn write(&mut self, msg: &Value) -> Result<(), LspError> {
        let stdin = self.stdin.as_mut().ok_or(LspError::NotRunning)?;
        write_message(stdin, msg).map_err(|e| LspError::Protocol(e.to_string()))
    }

    fn notify(&mut self, method: &str, params: Value) -> Result<(), LspError> {
        self.write(&notification(method, params))
    }

    fn request(&mut self, method: &str, params: Value) -> Result<Value, LspError> {
        let id = self.next_id();
        let (tx, rx) = mpsc::channel();
        {
            let mut map = self
                .pending
                .lock()
                .map_err(|_| LspError::Protocol("lock".into()))?;
            map.insert(id, Pending { tx });
        }
        if let Err(error) = self.write(&request(id, method, params)) {
            if let Ok(mut map) = self.pending.lock() {
                map.remove(&id);
            }
            return Err(error);
        }
        let deadline = Instant::now() + self.timeout;
        loop {
            let remaining = deadline.saturating_duration_since(Instant::now());
            if remaining.is_zero() {
                let mut map = self
                    .pending
                    .lock()
                    .map_err(|_| LspError::Protocol("lock".into()))?;
                map.remove(&id);
                return Err(LspError::Timeout);
            }
            match rx.recv_timeout(remaining.min(Duration::from_millis(50))) {
                Ok(r) => return r,
                Err(mpsc::RecvTimeoutError::Timeout) => continue,
                Err(mpsc::RecvTimeoutError::Disconnected) => {
                    return Err(LspError::Protocol("pending disconnected".into()));
                }
            }
        }
    }

    /// Drena eventos (diagnostics etc.).
    pub fn poll_events(&mut self) -> Vec<LspEvent> {
        let mut out = Vec::new();
        loop {
            match self.events.try_recv() {
                Ok(e) => out.push(e),
                Err(TryRecvError::Empty) => break,
                Err(TryRecvError::Disconnected) => {
                    out.push(LspEvent::Exited);
                    break;
                }
            }
        }
        out
    }

    pub fn did_open(&mut self, path: &Path, language_id: &str, text: &str) -> Result<(), LspError> {
        let uri = path_to_uri(path);
        if self.open_documents.contains(&uri) {
            return Ok(());
        }
        self.notify(
            "textDocument/didOpen",
            json!({
                "textDocument": {
                    "uri": &uri,
                    "languageId": language_id,
                    "version": 1,
                    "text": text
                }
            }),
        )?;
        self.open_documents.insert(uri);
        Ok(())
    }

    pub fn did_change(&mut self, path: &Path, version: i32, text: &str) -> Result<(), LspError> {
        let uri = path_to_uri(path);
        self.notify(
            "textDocument/didChange",
            json!({
                "textDocument": { "uri": uri, "version": version },
                "contentChanges": [{ "text": text }]
            }),
        )
    }

    pub fn did_save(&mut self, path: &Path, text: &str) -> Result<(), LspError> {
        let uri = path_to_uri(path);
        self.notify(
            "textDocument/didSave",
            json!({
                "textDocument": { "uri": uri },
                "text": text
            }),
        )
    }

    pub fn did_close(&mut self, path: &Path) -> Result<(), LspError> {
        let uri = path_to_uri(path);
        if !self.open_documents.contains(&uri) {
            return Ok(());
        }
        self.notify(
            "textDocument/didClose",
            json!({ "textDocument": { "uri": &uri } }),
        )?;
        self.open_documents.remove(&uri);
        Ok(())
    }

    pub fn hover(&mut self, path: &Path, pos: Position) -> Result<Option<HoverInfo>, LspError> {
        let uri = path_to_uri(path);
        let result = self.request(
            "textDocument/hover",
            json!({
                "textDocument": { "uri": uri },
                "position": { "line": pos.line, "character": pos.character }
            }),
        )?;
        Ok(parse_hover(&result))
    }

    pub fn completion(
        &mut self,
        path: &Path,
        pos: Position,
    ) -> Result<Vec<CompletionItem>, LspError> {
        let uri = path_to_uri(path);
        let result = self.request(
            "textDocument/completion",
            json!({
                "textDocument": { "uri": uri },
                "position": { "line": pos.line, "character": pos.character }
            }),
        )?;
        Ok(parse_completions(&result))
    }

    pub fn definition(&mut self, path: &Path, pos: Position) -> Result<Option<Location>, LspError> {
        let uri = path_to_uri(path);
        let result = self.request(
            "textDocument/definition",
            json!({
                "textDocument": { "uri": uri },
                "position": { "line": pos.line, "character": pos.character }
            }),
        )?;
        Ok(parse_location(&result))
    }

    pub fn formatting(
        &mut self,
        path: &Path,
        source: &str,
        tab_size: u8,
        insert_spaces: bool,
    ) -> Result<Option<String>, LspError> {
        let uri = path_to_uri(path);
        let result = self.request(
            "textDocument/formatting",
            json!({
                "textDocument": { "uri": uri },
                "options": { "tabSize": tab_size, "insertSpaces": insert_spaces }
            }),
        )?;
        apply_text_edits(source, &result).map_err(LspError::Protocol)
    }

    pub fn shutdown(&mut self) {
        let configured_timeout = self.timeout;
        self.timeout = self.timeout.min(Duration::from_millis(500));
        let _ = self.request("shutdown", json!(null));
        self.timeout = configured_timeout;
        let _ = self.notify("exit", json!(null));
        if let Some(mut child) = self.child.take() {
            let _ = child.kill();
            let _ = child.wait();
        }
        self.stdin = None;
        self.open_documents.clear();
        self.ready = false;
    }
}

impl Drop for LspClient {
    fn drop(&mut self) {
        self.shutdown();
    }
}

fn parse_diagnostics(params: &Value) -> Vec<Diagnostic> {
    let mut out = Vec::new();
    let Some(arr) = params.get("diagnostics").and_then(|d| d.as_array()) else {
        return out;
    };
    for d in arr {
        let range = parse_range(d.get("range")).unwrap_or(Range {
            start: Position {
                line: 0,
                character: 0,
            },
            end: Position {
                line: 0,
                character: 0,
            },
        });
        let severity = match d.get("severity").and_then(|s| s.as_u64()).unwrap_or(1) {
            1 => DiagnosticSeverity::Error,
            2 => DiagnosticSeverity::Warning,
            3 => DiagnosticSeverity::Information,
            _ => DiagnosticSeverity::Hint,
        };
        let message = d
            .get("message")
            .and_then(|m| m.as_str())
            .unwrap_or("")
            .to_string();
        let source = d.get("source").and_then(|s| s.as_str()).map(str::to_string);
        out.push(Diagnostic {
            range,
            severity,
            message,
            source,
        });
    }
    out
}

fn parse_range(v: Option<&Value>) -> Option<Range> {
    let v = v?;
    let start = v.get("start")?;
    let end = v.get("end")?;
    Some(Range {
        start: Position {
            line: start.get("line")?.as_u64()? as u32,
            character: start.get("character")?.as_u64()? as u32,
        },
        end: Position {
            line: end.get("line")?.as_u64()? as u32,
            character: end.get("character")?.as_u64()? as u32,
        },
    })
}

fn parse_hover(result: &Value) -> Option<HoverInfo> {
    if result.is_null() {
        return None;
    }
    let contents = result.get("contents")?;
    let text = if let Some(s) = contents.as_str() {
        s.to_string()
    } else if let Some(obj) = contents.as_object() {
        obj.get("value")
            .and_then(|v| v.as_str())
            .unwrap_or("")
            .to_string()
    } else {
        let arr = contents.as_array()?;
        arr.iter()
            .filter_map(|v| {
                v.as_str()
                    .map(str::to_string)
                    .or_else(|| v.get("value").and_then(|x| x.as_str()).map(str::to_string))
            })
            .collect::<Vec<_>>()
            .join("\n")
    };
    if text.is_empty() {
        None
    } else {
        Some(HoverInfo { contents: text })
    }
}

fn parse_completions(result: &Value) -> Vec<CompletionItem> {
    let items = if let Some(arr) = result.as_array() {
        arr.clone()
    } else if let Some(arr) = result.get("items").and_then(|i| i.as_array()) {
        arr.clone()
    } else {
        return Vec::new();
    };
    items
        .iter()
        .filter_map(|it| {
            let label = it.get("label")?.as_str()?.to_string();
            let detail = it
                .get("detail")
                .and_then(|d| d.as_str())
                .map(str::to_string);
            let mut insert_text = it
                .get("insertText")
                .and_then(|d| d.as_str())
                .or_else(|| {
                    it.get("textEdit")
                        .and_then(|edit| edit.get("newText"))
                        .and_then(|text| text.as_str())
                })
                .map(str::to_string);
            if it.get("insertTextFormat").and_then(Value::as_u64) == Some(2) {
                insert_text = insert_text.map(|text| snippet_to_plain_text(&text));
            }
            Some(CompletionItem {
                label,
                detail,
                insert_text,
            })
        })
        .collect()
}

fn snippet_to_plain_text(snippet: &str) -> String {
    let mut plain = String::new();
    let mut cursor = 0usize;
    while cursor < snippet.len() {
        let remaining = &snippet[cursor..];
        if let Some(body) = remaining.strip_prefix("${") {
            if let Some(close) = body.find('}') {
                let placeholder = &body[..close];
                if let Some((_, default)) = placeholder.split_once(':') {
                    plain.push_str(default);
                }
                cursor += 2 + close + 1;
                continue;
            }
        }
        if let Some(body) = remaining.strip_prefix('$') {
            let digits = body.bytes().take_while(u8::is_ascii_digit).count();
            if digits > 0 {
                cursor += 1 + digits;
                continue;
            }
        }
        let character = remaining.chars().next().expect("cursor within snippet");
        plain.push(character);
        cursor += character.len_utf8();
    }
    plain
}

fn parse_location(result: &Value) -> Option<Location> {
    if result.is_null() {
        return None;
    }
    let loc = if let Some(arr) = result.as_array() {
        arr.first()?
    } else {
        result
    };
    // LocationLink vs Location
    let uri = loc
        .get("uri")
        .or_else(|| loc.get("targetUri"))
        .and_then(|u| u.as_str())?
        .to_string();
    let range = parse_range(
        loc.get("range")
            .or_else(|| loc.get("targetSelectionRange"))
            .or_else(|| loc.get("targetRange")),
    )?;
    Some(Location { uri, range })
}

fn apply_text_edits(source: &str, result: &Value) -> Result<Option<String>, String> {
    let arr = result
        .as_array()
        .ok_or_else(|| "format response is not an array".to_string())?;
    if arr.is_empty() {
        return Ok(None);
    }

    let mut edits = Vec::with_capacity(arr.len());
    for edit in arr {
        let range = parse_range(edit.get("range"))
            .ok_or_else(|| "format edit has an invalid range".to_string())?;
        let new_text = edit
            .get("newText")
            .and_then(Value::as_str)
            .ok_or_else(|| "format edit has no newText".to_string())?;
        let start = lsp_position_to_byte(source, range.start)
            .ok_or_else(|| "format edit start is outside the document".to_string())?;
        let end = lsp_position_to_byte(source, range.end)
            .ok_or_else(|| "format edit end is outside the document".to_string())?;
        if start > end {
            return Err("format edit range is reversed".into());
        }
        edits.push((start, end, new_text));
    }

    edits.sort_by_key(|(start, _, _)| std::cmp::Reverse(*start));
    let mut next_start = source.len();
    let mut formatted = source.to_string();
    for (start, end, new_text) in edits {
        if end > next_start {
            return Err("format edits overlap".into());
        }
        formatted.replace_range(start..end, new_text);
        next_start = start;
    }

    Ok(Some(formatted))
}

fn lsp_position_to_byte(text: &str, position: Position) -> Option<usize> {
    let mut line_start = 0usize;
    for _ in 0..position.line {
        let newline = text.get(line_start..)?.find('\n')?;
        line_start += newline + 1;
    }
    let remainder = text.get(line_start..)?;
    let line_end = remainder.find('\n').unwrap_or(remainder.len());
    let line = remainder[..line_end]
        .strip_suffix('\r')
        .unwrap_or(&remainder[..line_end]);
    let character_column = character_column(line, position.character)?;
    let byte_column = line
        .chars()
        .take(character_column)
        .map(char::len_utf8)
        .sum::<usize>();
    Some(line_start + byte_column)
}

#[cfg(test)]
mod tests {
    // A snippet comes from the server, so it is untrusted input: a malformed one
    // must produce a degraded completion, never a panic that takes down the
    // editor. These cases pin that.
    #[test]
    fn malformed_snippets_do_not_panic() {
        use super::snippet_to_plain_text;

        let cases = [
            "",
            "$",
            "${",
            "$}",
            "${}",
            "${:}",
            "${a",
            "${a:",
            "${a:${b}",
            "$1",
            "$99999999999999999999",
            "a$",
            "a${",
            "${nome}",
            "${1:nome}",
            "まで${1:日本語}",
            "${1:日本語",
            "$$",
            "${}}",
            "${1:${2:x}}",
            "\u{1F600}${1:😀}",
            "${1:😀",
            "$😀",
            "}",
        ];
        for snippet in cases {
            // The assertion is that this returns at all.
            let _ = snippet_to_plain_text(snippet);
        }
    }

    #[test]
    fn converts_plain_snippets() {
        use super::snippet_to_plain_text;

        assert_eq!(snippet_to_plain_text("return"), "return");
        assert_eq!(snippet_to_plain_text("${1:nome}"), "nome");
        assert_eq!(snippet_to_plain_text("${1}"), "");
        assert_eq!(snippet_to_plain_text("$1aceita"), "aceita");
    }

    use super::*;

    #[test]
    fn parse_diag_severity() {
        let params = json!({
            "uri": "file:///x.oris",
            "diagnostics": [{
                "range": {
                    "start": {"line": 1, "character": 0},
                    "end": {"line": 1, "character": 3}
                },
                "severity": 1,
                "message": "boom"
            }]
        });
        let d = parse_diagnostics(&params);
        assert_eq!(d.len(), 1);
        assert_eq!(d[0].message, "boom");
        assert_eq!(d[0].severity, DiagnosticSeverity::Error);
    }

    #[test]
    fn applies_multiple_format_edits_with_utf16_ranges() {
        let source = "a😀b\nsecond\n";
        let edits = json!([
            {
                "range": {
                    "start": {"line": 0, "character": 1},
                    "end": {"line": 0, "character": 3}
                },
                "newText": "X"
            },
            {
                "range": {
                    "start": {"line": 1, "character": 0},
                    "end": {"line": 1, "character": 6}
                },
                "newText": "2nd"
            }
        ]);

        assert_eq!(
            apply_text_edits(source, &edits).unwrap(),
            Some("aXb\n2nd\n".into())
        );
    }

    #[test]
    fn parses_completion_text_edits_and_plain_snippets() {
        let result = json!([
            {
                "label": "return",
                "textEdit": { "newText": "return" }
            },
            {
                "label": "if",
                "insertText": "if ${1:condition}\n    ${0}\nend",
                "insertTextFormat": 2
            }
        ]);

        let items = parse_completions(&result);
        assert_eq!(items[0].insert_text.as_deref(), Some("return"));
        assert_eq!(
            items[1].insert_text.as_deref(),
            Some("if condition\n    \nend")
        );
    }
}
