# Code Review — Verification Checklist

## 1. Contract & Scope
- [ ] Acceptance criteria restated before reading code.
- [ ] Base revision and full diff range identified; no unreviewed commits in range.
- [ ] Locked invariants listed; review checks each one explicitly.

## 2. Finding Quality
- [ ] Every finding carries severity (blocker / major / minor / nit).
- [ ] Every blocker and major cites file:line and a concrete failure scenario.
- [ ] No severity-free commentary presented as a finding.

## 3. Silent-Failure Patterns
- [ ] No acceptance criteria weakened or deleted without ADR/policy update.
- [ ] No widened permissions, new privileged capabilities, or bypassed gates.
- [ ] No deleted, skipped, or `t.Skip`-muted tests covering changed behavior.
- [ ] No new unpinned or unvetted external dependencies.

## 4. Behavioral Verification
- [ ] Affected package tests executed and green (`go test`, `pytest`, `npm test`).
- [ ] At least one static analyzer run on touched packages with zero new diagnostics.
- [ ] Commit-message claims cross-checked against the actual diff.

## 5. Verdict & Re-verification
- [ ] Verdict is one of APPROVE / REQUEST CHANGES / ESCALATE, recorded with rationale.
- [ ] All blockers and majors re-verified against fix commits, never approved on prose.
- [ ] Escalations include a written handoff with scope, risk, and evidence.
