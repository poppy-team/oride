# Code Quality Metrics, Models & Remediation Reference Guide

## 1. Core Concepts & Formulas

### 1.1 Cyclomatic Complexity
McCabe cyclomatic complexity counts independent paths through a function:

$$M = E - N + 2P$$

Where $E$ is edges, $N$ is nodes in the control-flow graph, and $P$ is connected components (1 per function).
In practice: start at 1 and add 1 per `if`, `for`, `while`, `case`, `&&`, `||`, `catch`, ternary.

| M range | Risk | Action |
|---|---|---|
| 1–10 | Low | Ship |
| 11–20 | Moderate | Refactor within the sprint |
| 21–50 | High | Decompose before release; blocks merge |
| 50+ | Untestable | Escalate; redesign required |

### 1.2 Cognitive Complexity
Sonar-style cognitive complexity penalizes nesting over flat branching: each nesting level
adds its depth as weight. A flat 8-arm switch scores ~8; the same logic triple-nested scores ~24.
Ceiling: 15 per function. Prefer early returns and guard clauses, which add 1 without nesting weight.

### 1.3 Maintainability Index
The classic SEI maintainability index combines volume, complexity, and size:

$$MI = \max(0, 171 - 5.2 \ln(HV) - 0.23 M - 16.2 \ln(LOC)) \times \frac{100}{171}$$

Where $HV$ is Halstead volume, $M$ is average cyclomatic complexity, $LOC$ is lines of code.
Interpretation: 65–100 maintainable, 40–64 watch list, below 40 remediation candidate.

### 1.4 Duplication Ratio
$$D = \frac{\text{duplicated lines}}{\text{total non-blank lines}} \times 100$$

Measured with `jscpd --min-lines 15 --min-tokens 80`. Gate: $D \le 3\%$ per module.
Below the 15-line / 80-token floor, clones are noise; above it, they are backlog items.

## 2. Patterns and Anti-Patterns

**Do: Extract Function with a domain name.**
A 90-line `processInvoice` scoring M=18 becomes `validateInvoice` (M=4),
`computeTotals` (M=6), `persistInvoice` (M=5). Each part is independently testable.

**Do: Replace Conditional with Polymorphism.**
A 12-arm switch on `payment_method` with M=13 becomes a registry map of 12
single-path handlers; adding method 13 no longer touches existing code.

**Do not: shotgun surgery.** Fixing the same tax-rounding bug in 4 cloned
helpers instead of extracting one `roundTax` helper guarantees the 5th copy drifts.

**Do not: suppression without expiry.** A `//nolint:funlen // TODO legacy, revisit 2026-12-01`
is traceable; a bare `//nolint` hides decay permanently.

**Do not: bulk reformat with remediation.** Mixing `gofmt -w ./...` into a hotspot
slice destroys blame history and reviewability; run formatting as a separate change.

## 3. Tool Configuration Example

```yaml
# .golangci.yaml — complexity and duplication gates for a Go service
linters:
  enable:
    - gocyclo
    - cyclop
    - dupl
    - funlen
    - revive
    - misspell
linters-settings:
  gocyclo:
    min-complexity: 10
  cyclop:
    max-complexity: 10
    package-skip: ".*_test|.*generated.*"
  dupl:
    threshold: 80
  funlen:
    lines: 80
    statements: 50
issues:
  exclude-rules:
    - path: "_test\\.go"
      linters: [funlen]
    - path: "internal/mocks/"
      linters: [dupl]
```

```ini
# jscpd duplication gate (package.json script)
# "quality:duplicates": "jscpd --min-lines 15 --min-tokens 80 --threshold 3 src/"
```

Run order for a full baseline: `golangci-lint run ./...`, `ruff check .`,
`npx eslint .`, `cargo clippy --all-targets -- -D warnings`, then
`jscpd --min-lines 15 --min-tokens 80 src/`.
