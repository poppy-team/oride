# Code Quality Auditing — Verification Checklist

## 1. Complexity Ceilings
- [ ] Every function reports cyclomatic complexity at or below 10 (`gocyclo -over 10` empty, ruff C901 clean, eslint complexity rule clean).
- [ ] No function exceeds cognitive complexity 15; deeply nested conditionals (nesting depth above 4) are extracted.
- [ ] Complexity hotspots are ranked by `risk = max_complexity * log(1 + churn_90d) * (1 + defects_90d)` with the top 5 listed.
- [ ] Each flagged hotspot was manually confirmed as genuine branching, not a tool miscount on match expressions or JSX conditionals.

## 2. Duplication Invariants
- [ ] Duplicated lines are at or below 3 percent of total lines per `jscpd --min-lines 15 --min-tokens 80`.
- [ ] Confirmed clones are semantic copies (same logic, renamed identifiers), extracted into shared helpers.
- [ ] Coincidental structural similarity under 15 lines is documented as accepted, not extracted.
- [ ] No copy-pasted fix applied in two places where a shared helper was possible.

## 3. Lint & Static Analysis Gates
- [ ] `golangci-lint run ./...` exits 0 with zero findings above info severity.
- [ ] `ruff check .` and `npx eslint .` exit 0; no new suppressions added without a dated justification comment.
- [ ] `cargo clippy --all-targets -- -D warnings` exits 0 where Rust code exists.
- [ ] Generated code, vendored dependencies, and migration snapshots are excluded via config, never hand-edited.

## 4. Remediation Discipline
- [ ] Each remediation slice changes fewer than 200 lines and targets exactly one hotspot.
- [ ] Complexity and duplication numbers monotonically decrease after every slice (re-run logs attached).
- [ ] Public APIs, exported signatures, and architecture layer boundaries are unchanged.
- [ ] No repository-wide reformatting is mixed into remediation slices.

## 5. Evidence & Sign-Off
- [ ] Audit report follows `templates/code-quality-spec.md` with before/after metric pairs per hotspot.
- [ ] Coverage holds at or above 85 percent for business logic in touched packages; deltas reported per slice.
- [ ] Affected package test suites pass (`go test`, `pytest -q`) after each slice.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
- [ ] Residual hotspots are recorded as tech-debt items with owner and revisit date, or explicitly accepted by the lead.
