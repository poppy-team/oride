# Independent Code Review Reference Guide

## 1. Severity Model

| Severity | Definition | Example | Verdict effect |
|---|---|---|---|
| `blocker` | Certainty of defect, data loss, security breach, or contract violation in a plausible scenario | Unauthenticated admin endpoint; migration that drops a column with live data | Forces REQUEST CHANGES |
| `major` | Likely defect or significant maintainability damage | Race condition under concurrent load; public API breaking change without migration | Forces REQUEST CHANGES |
| `minor` | Real issue with low blast radius | Confusing naming in new public function; missing edge-case test | Advisory |
| `nit` | Style or preference, no behavioral consequence | Import ordering; comment wording | Advisory, batched |

Blast radius is scored as `severity × exposure`: a `minor` in authentication middleware outranks a `major` in a deprecated CLI flag. Always state the scenario, not just the pattern.

## 2. Reading Order

1. Contracts first: interfaces, schemas, migrations, public signatures.
2. State changes: where data is written, cached, queued, or deleted.
3. Control flow: error paths, retries, timeouts, concurrency primitives.
4. Tests last: do they assert the contract from step 1, or merely the implementation?

If step 4 only mirrors the implementation, the change is untested regardless of coverage numbers.

## 3. The Four Silent Failures

- **Weakened criteria**: acceptance thresholds lowered, assertions removed, goldens regenerated without justification.
- **Widened authority**: new filesystem/network/process capabilities, elevated roles, disabled policy gates.
- **Test erosion**: deleted tests, added skips, narrowed table cases, mocked boundaries that were previously integrated.
- **Dependency drift**: new external packages, version-range widening, vendored binaries, curl-piped install scripts.

```sh
# Fast silent-failure sweep over a diff (illustrative)
git diff --stat HEAD~1          # scope: files vs. tests ratio
git diff HEAD~1 -- '*test*' | grep -c '^[-]'   # removed test lines deserve scrutiny
git diff HEAD~1 -- go.mod package.json Cargo.toml pnpm-lock.yaml | head -n 40
```

## 4. Verdict Discipline

- APPROVE requires zero open blockers/majors and green required checks.
- REQUEST CHANGES must enumerate every blocker/major with file:line; the author fixes, the reviewer re-verifies.
- ESCALATE when the diff touches cryptography, key management, billing, personal data pipelines, or governance policy itself. The escalation note states what was reviewed, what was not, and the exact question for the authority.

## 5. Anti-Patterns

- Rubber-stamping large diffs ("LGTM" on >400 changed lines without a reading trail).
- Reviewing prose instead of code (approving on description, screenshots, or author reputation).
- Fix-driving (rewriting the author's code inside review comments instead of stating the required property).
- Moving goalposts (introducing new requirements mid-review that were never acceptance criteria).
