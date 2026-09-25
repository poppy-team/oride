# Language Tooling & LSP — Verification Checklist

## 1. Session Handshake & Capability Negotiation
- [ ] `initialize` carries processId, rootUri, and the full client capability set including snippet support.
- [ ] Server `capabilities` reply is recorded; every method the workflow needs is advertised before use.
- [ ] Server version is pinned to an exact release, never a floating latest tag.

## 2. Document Synchronization & Semantic Queries
- [ ] Files open with `textDocument/didOpen`, stream incremental `didChange` events, and close with `didClose`.
- [ ] Hover, definition, references with includeDeclaration, and workspace symbol lookup all resolve against the current document version.
- [ ] `workspace/symbol` results are capped at 50 entries and definition responses are cached per document version.

## 3. Diagnostics, Formatting & Code Actions
- [ ] Diagnostics arrive via push or pull with at most 250 ms debounce after the last keystroke.
- [ ] Stale markers for closed files are cleared immediately on `didClose`.
- [ ] Formatting is idempotent: formatting twice yields byte-identical output.
- [ ] Every `workspace/applyEdit` has a confirmation record with before and after text.

## 4. Debug Adapter Sessions
- [ ] DAP `launch` or `attach` reaches a stopped-at-breakpoint state on the sample target.
- [ ] `threads`, `stackTrace`, and `variables` return coherent frames for the stopped thread.
- [ ] `disconnect` reaps the adapter process; no orphaned debug processes remain.

## 5. Performance & Verification Evidence
- [ ] Server startup completes in under 2 seconds on the reference workspace.
- [ ] Hover p95 under 150 ms and completion p95 under 300 ms on the replay script.
- [ ] 200-request replay covering hover, definition, references, rename, and formatting shows zero protocol errors.
- [ ] `scripts/verify.sh` executes with exit code 0.
