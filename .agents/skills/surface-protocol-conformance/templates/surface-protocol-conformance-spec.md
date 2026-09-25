# Surface & Protocol Conformance Specification

## Contract
- Protocol: Prumo Protocol `2026-09`
- Operation under review: `budget.commit`
- Surfaces: CLI, TUI, GUI, and daemon
- Canonical entry point: `domain/budget.commit`

## Shared Semantics
- Permission Engine verdict is required before the ledger changes.
- Budget Manager consumes one unit from the active Goal.
- Exceeded budgets return `BUDGET_EXCEEDED`; denied access returns `PERMISSION_DENIED`.
- No surface calculates a local price or bypasses the daemon gate.

## Conformance Matrix
| Surface | Entry point | Gate | Error fixture | Result |
|---|---|---|---|---|
| CLI | `domain/budget.commit` | Budget Manager | `errors.json#BUDGET_EXCEEDED` | pass |
| TUI | `domain/budget.commit` | Permission Engine | `errors.json#PERMISSION_DENIED` | pass |
| GUI | `domain/budget.commit` | Budget Manager | `errors.json#BUDGET_EXCEEDED` | pass |
| daemon | `domain/budget.commit` | Gauntlet Quality Gates | `errors.json#VERSION_MISMATCH` | pass |

## Replay Evidence
- Golden session `trace-0042.json` replays to ledger hash `b7e2c1a9` on every surface.
- Client version 3 negotiates with daemon version 4; missed event 118 is applied exactly once.
- A version outside the supported window fails closed with `PROTOCOL_VERSION_UNSUPPORTED`.
