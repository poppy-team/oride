# Language Tooling — Deliverable Specification

## 1. Workspace & Servers
- **Workspace**: AcmePy billing monorepo at /srv/acmepy, 4,120 Python files, 310,000 lines
- **Primary server**: pyright 1.1.390 over stdio, basic type-check mode for legacy/shop, strict mode for billing-core/
- **Secondary server**: ruff-lsp 0.11.2 for lint diagnostics on the same workspace
- **Debug adapter**: debugpy 1.8.14, launch configuration for pytest targets
- **Excluded paths**: .venv/, build/, generated/migrations_rendered/

## 2. Capability Matrix
- **Advertised by pyright**: hover, definition, references, workspace symbol, rename, formatting, pull diagnostics
- **Client-declared**: snippetSupport true, markdown hover, relatedInformation diagnostics, dynamic file watching
- **Disabled by negotiation**: semantic tokens, the server build predates that provider
- **Handshake recorded**: 2026-09-17, initialize reply archived at .prumo/lsp/initialize-reply.json

## 3. Diagnostics & Edit Policy
- **Transport**: push via textDocument/publishDiagnostics, debounced 250 ms after last keystroke
- **Stale-marker rule**: didClose clears all markers for that URI within the same tick
- **applyEdit rule**: every server edit requires a confirmation record with before and after text, stored under .prumo/lsp/edits/
- **Formatting**: ruff format through the server, idempotency verified byte-identical on second pass

## 4. Debug Flow Proven
- **Target**: tests/test_invoice_totals.py::test_pro_rata_refund, launched via debugpy
- **Breakpoints hit**: billing/proration.py line 88 and line 112, variables frame shows amount_cents 4990 and rate 0.75
- **Teardown**: disconnect with terminateDebuggee true, zero orphaned python processes confirmed with pgrep

## 5. Measured Results
- **Server startup**: 1.4 s cold on the reference workspace
- **Hover p95**: 96 ms over 200 sampled hovers
- **Completion p95**: 231 ms over 200 sampled completions
- **Replay script**: 200 requests across hover, definition, references, rename, and formatting, zero protocol errors

## 6. Verification Evidence
- Replay log at .prumo/lsp/replay-2026-09-17.log with latency percentiles
- Formatting idempotency diff at .prumo/lsp/format-idempotent-2026-09-17.diff, empty as required
- `scripts/verify.sh` exit code 0 on 2026-09-17
