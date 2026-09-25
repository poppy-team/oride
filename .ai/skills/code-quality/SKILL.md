# Code Quality Auditing, Complexity & Maintainability

## Purpose
Audit cyclomatic and cognitive complexity, code duplication, and maintainability metrics across polyglot codebases using deterministic static analysis gates (ruff, eslint, golangci-lint, clippy), enforce complexity ceilings and duplication thresholds, and produce prioritized remediation plans with before/after evidence.

## Use when
- Auditing a module, package, or repository for complexity hotspots, cloned logic, or maintainability decay before a release or refactoring Goal.
- Enforcing quality gates in review or audit mode: cyclomatic complexity ceilings, duplication ratios, lint-clean baselines.
- Planning a refactoring Goal by ranking hotspots on risk times churn and sizing remediation slices.
- Investigating maintainability regressions flagged by coverage drops or rising defect counts.

## Do not use when
- Fixing a single known bug with an obvious minimal patch (apply the patch directly; no audit needed).
- Designing runtime caching topologies or eviction policies (use `caching`).
- Tuning CI pipeline speed, runner caching, or container image size (use `ci-cd` or `containers`).
- Measuring wall-clock performance or tail latency (use `benchmarking`).

## Required context
- Repository checkout at a pinned commit plus language toolchain versions (Go 1.24.3, Python 3.12.4, Node 22.6.0).
- Lint and analysis configuration in force: `.golangci.yaml`, ruff section of `pyproject.toml`, `eslint.config.js`, clippy lints in `Cargo.toml`.
- Complexity ceilings and duplication thresholds from the Goal or team policy (defaults: cyclomatic complexity at most 10 per function, duplication at most 3 percent of lines).
- Hotspot history: per-module churn from `git log --stat` and defect counts from the tracker for risk ranking.

## Procedure
1. **Establish the measurement baseline**:
   - Run the full lint and analysis suite from a clean state and record counts: `golangci-lint run ./...`, `ruff check .`, `npx eslint .`, `cargo clippy --all-targets -- -D warnings`.
   - Capture complexity metrics per function with `gocyclo -over 10 .`, `ruff` C901 rules, `eslint` complexity rule, or `cargo clippy::cognitive_complexity`.
   - Capture duplication ratios with `jscpd --min-lines 15 --min-tokens 80 src/` and record the percentage of duplicated lines.
2. **Rank hotspots by risk times churn**:
   - Compute risk score per module as `risk = max_complexity * log(1 + churn_commits_90d) * (1 + defects_90d)`.
   - Sort modules descending by risk score; the top 5 modules become the remediation backlog.
   - Discard hotspots in generated code, vendored dependencies, and migration snapshots; document each exclusion with its path pattern.
3. **Verify each hotspot manually**:
   - Open each flagged function and confirm the metric reflects real branching (nested conditionals, switch arms, boolean chains), not tool miscounts on match expressions or JSX conditionals.
   - Confirm duplication findings are semantic clones (same logic, renamed identifiers), not coincidental structural similarity under 15 lines.
4. **Remediate in smallest sufficient slices**:
   - Apply Extract Function, Replace Conditional with Polymorphism, or Extract Shared Helper; keep each slice under 200 changed lines.
   - Re-run the exact baseline commands after every slice; complexity and duplication numbers must monotonically decrease.
5. **Close the loop with evidence**:
   - Record before/after metric pairs per hotspot in `templates/code-quality-spec.md`.
   - Run the affected package test suite (`go test ./internal/billing/`, `pytest tests/billing -q`) and the verification script `scripts/verify.sh`; both must exit 0.

## Decision rules
- **Ceilings Are Hard Gates**: No function ships with cyclomatic complexity above 10; no module ships with duplication above 3 percent. Exceeding either blocks merge.
- **Never Bulk-Reformat**: Remediation slices must stay under 200 changed lines; repository-wide reformatting in the same change is prohibited.
- **Generated Code Is Excluded, Never Edited**: Findings inside generated, vendored, or snapshot files must be suppressed via config, never hand-edited.
- **Metrics Must Monotonically Improve**: Every remediation slice must lower or hold steady the hotspot numbers; a slice that raises complexity is rejected.
- **Tool Counts Are Advisory Until Confirmed**: No hotspot enters the backlog until a human confirms the flagged code is genuinely complex or duplicated.

## Evidence required
- Quality audit report adhering to `templates/code-quality-spec.md` with before/after metric pairs.
- Raw linter and analyzer logs (golangci-lint, ruff, eslint, clippy, jscpd) captured at baseline and after remediation.
- Test logs for affected packages showing green suites after each remediation slice.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Ranked hotspot backlog with risk scores, metric values, and file-level locations.
- Remediation slices as small reviewable changes, each with before/after complexity and duplication deltas.
- Updated lint suppression config only where generated-code exclusions were missing.
- Audit report with totals: functions over ceiling, duplication percentage, lint findings by severity.

## Stop conditions
- All functions at or below cyclomatic complexity 10 and duplication at or below 3 percent, verified by a clean analyzer run.
- Remaining hotspots formally accepted as tech-debt items with owner and revisit date recorded in the audit report.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to Lead Architect if a hotspot above complexity 25 cannot be decomposed without changing a public API or locked Goal acceptance criteria.
- Escalate to the owning team if duplication spans service boundaries and requires a shared-library extraction decision.
- Escalate immediately upon discovering hardcoded secrets, private keys, or credentials during the audit.
