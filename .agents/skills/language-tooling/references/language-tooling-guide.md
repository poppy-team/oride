# Language Tooling (LSP & DAP) — Technical Reference Guide

## 1. Core Concepts

### 1.1 JSON-RPC Framing
LSP and DAP both speak JSON-RPC 2.0 over headers plus body. Each message is prefixed with a byte count:

```
Content-Length: 184\r\n\r\n{"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}
```

Read exactly `Content-Length` bytes for the body; a short read corrupts every subsequent message. Stdio transport spawns the server as a child process, while TCP transport connects to a port — stdio is preferred for single-workspace agents because process lifetime is trivially managed.

### 1.2 The Initialize Handshake
The client opens with `initialize`, declaring what it can handle, and the server answers with what it provides. A minimal Python-oriented handshake:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "processId": 4120,
    "rootUri": "file:///srv/acmepy",
    "capabilities": {
      "textDocument": {
        "completion": { "completionItem": { "snippetSupport": true } },
        "hover": { "contentFormat": ["markdown", "plaintext"] },
        "publishDiagnostics": { "relatedInformation": true }
      }
    }
  }
}
```

If the reply lacks `definitionProvider`, the client must disable go-to-definition rather than calling it and logging errors.

### 1.3 Document Lifecycle and Diagnostics
Documents move through `didOpen` with full text, `didChange` with incremental edits, `save` notifications, and `didClose`. Diagnostics flow back either as `textDocument/publishDiagnostics` pushes or via `textDocument/diagnostic` pulls. Debounce 250 ms after the final keystroke so typing bursts produce one diagnostic pass, not thirty.

### 1.4 DAP Session Shape
A debug session runs `initialize`, then `launch` with the program path or `attach` with a process id, then `setBreakpoints`, `configurationDone`, and event-driven stops. `threads` plus `stackTrace` plus `variables` reconstruct the paused state; `disconnect` with `terminateDebuggee: true` ends everything cleanly.

## 2. Server Selection Notes
- pyright 1.1.390: fastest Python type queries, strict mode raises0052 false-positive rate on untyped legacy code — scope it per directory.
- rust-analyzer rolling 2026-09-01: needs `cargo metadata` to succeed before workspace load; without it, symbols are empty.
- clangd 17.0.3: reads `compile_commands.json` directly; background indexing warms the first query from minutes to milliseconds.
- typescript-language-server 4.3.3: delegates to the workspace TypeScript 5.6 SDK; mismatched SDK versions cause phantom errors.

## 3. Common Pitfalls
- Calling unadvertised methods and treating the error response as a server bug.
- Letting the server view drift: applying an edit locally without sending `didChange` makes the next definition land on stale offsets.
- Forgetting `initialized` notification after `initialize`, which leaves some servers permanently muted.
