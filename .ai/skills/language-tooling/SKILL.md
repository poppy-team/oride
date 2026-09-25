# Language Tooling: LSP & DAP Integration

## Purpose
Integrate, configure, and verify Language Server Protocol (LSP) and Debug Adapter Protocol (DAP) tooling: JSON-RPC lifecycle management, capability negotiation, document synchronization, hover, definition, completion, references, workspace symbols, diagnostics, formatting, code actions, and debug sessions with breakpoints, threads, and variable inspection.

## Use when
- An agent or editor feature needs semantic language services such as go-to-definition, find-references, hover types, or rename across Python, Rust, C++, C#, or TypeScript codebases.
- Selecting, upgrading, or pinning a language server such as pyright 1.1.390, rust-analyzer 2026-09-01, clangd 17.0.3, OmniSharp, or typescript-language-server 4.3.3.
- Wiring Debug Adapter Protocol launch, attach, breakpoint, stepping, and thread inspection flows.
- Diagnosing slow completions, stale diagnostics, or capability mismatches between client and server.

## Do not use when
- Only tree-level syntax highlighting without types or cross-file resolution is needed; tree-sitter grammars alone are cheaper.
- The task is pure static analysis policy with no live editor or agent query path; use a linting workflow instead.
- Debugging requires kernel-level or embedded JTAG access that DAP adapters cannot reach.

## Required context
- Target languages with pinned server versions and transport used, stdio versus TCP socket.
- Client capability set: completion snippet support, pull versus push diagnostics, workspace file watching.
- Latency budgets: server startup under 2 seconds, hover p95 under 150 ms, completion p95 under 300 ms.
- Workspace root, file globs served, and files explicitly excluded such as generated code and vendored dependencies.

## Procedure
1. **Negotiate the session**: send `initialize` with `processId`, `rootUri`, and `capabilities`, for example declaring `textDocument.completion.completionItem.snippetSupport: true`. Record the server `capabilities` reply and abort if a required method such as `textDocument/definition` is absent.
2. **Synchronize documents**: open files with `textDocument/didOpen` carrying the full text and `languageId`, stream edits via `textDocument/didChange` with incremental content changes, and close with `textDocument/didClose`. Never let the server view drift from the buffer on disk.
3. **Serve semantic queries**: route hover to `textDocument/hover`, definitions to `textDocument/definition`, references to `textDocument/references` with `includeDeclaration: true`, and project-wide lookup to `workspace/symbol` capped at 50 results. Cache definition responses per document version.
4. **Stream diagnostics and edits**: accept `textDocument/publishDiagnostics` pushes or poll with `textDocument/diagnostic`, debounce at 250 ms after the last keystroke, and apply server code actions only through `workspace/applyEdit` with a user-visible diff.
5. **Drive debug sessions**: start DAP with `initialize` plus `launch` or `attach`, set breakpoints via `setBreakpoints`, inspect frames with `threads`, `stackTrace`, and `variables`, and always terminate with `disconnect` so no orphaned adapter process survives.
6. **Verify end to end**: run `scripts/verify.sh`, measure startup, hover p95, and completion p95 against the budgets, and replay a 200-request script covering hover, definition, references, rename, and formatting with zero protocol errors.

## Decision rules
- **Capabilities gate everything**: never call an LSP method the server did not advertise in its `initialize` result.
- **Pin server versions**: floating `latest` servers are forbidden in reproducible pipelines; record exact versions in the spec.
- **Debounce diagnostics**: diagnostics may lag typing by at most 250 ms of debounce, and stale markers for closed files must be cleared immediately.
- **No silent edits**: every `workspace/applyEdit` requires an explicit confirmation record with before and after text.
- **Kill adapters on exit**: each DAP session owns its process handle and must reap it on disconnect, success, or failure.

## Evidence required
- Tooling specification following `templates/language-tooling-spec.md` with pinned server versions and capability matrix.
- Protocol replay log of the 200-request verification script with zero errors and measured latency percentiles.
- Diagnostics debounce measurement and formatting idempotency check output.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Initialized LSP session with negotiated capabilities and synchronized workspace.
- Verified semantic query path for hover, definition, references, symbols, and formatting.
- DAP launch and breakpoint flow proven against a sample target with clean teardown.

## Stop conditions
- All latency budgets met with zero protocol errors on the replay script.
- Formatting proven idempotent and rename proven cross-file on the sample workspace.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the tooling owner if a pinned server crashes more than twice per hour or omits a capability the task requires.
- Escalate to the lead architect if two candidate servers disagree on semantics, for example conflicting rename ranges, so a canonical engine can be chosen.
