# Independent Evidence-Based Code Review

## Purpose
Perform independent, evidence-based code reviews of multi-model implementations with deterministic verdicts (Approve / Request Changes / Escalate), severity-graded findings, and mandatory re-verification of fixes.

## Use when
- Reviewing a diff, pull request, or agent-produced implementation before merge or handoff.
- Auditing agent output for contract violations, security gaps, or silently weakened acceptance criteria.
- Acting as Independent Verifier for high/critical-risk changes under repository governance.
- Comparing competing implementations against the same acceptance criteria.

## Do not use when
- Writing the implementation itself (use `grounded-implementation` or domain skills).
- Running security-focused threat analysis (use `security-review`).
- Measuring stylistic lint only without semantic review (use `clean-code`).

## Required context
- The unified diff or commit range under review plus the base revision.
- Active Goal acceptance criteria and any locked invariants.
- Repository policy (branch rules, required checks, reviewer requirements).
- Test and static-analysis reports for the change (coverage, linters, SAST).

## Procedure
1. **Establish the contract**: Restate the acceptance criteria and locked invariants the change must satisfy before reading any code.
2. **Read the diff in dependency order**: Start from public interfaces and contracts, then implementations, then tests. Flag any file that cannot be understood in isolation.
3. **Grade every finding**: Assign severity (`blocker`, `major`, `minor`, `nit`) with a concrete failure scenario for blockers and majors. No severity-free commentary.
4. **Verify behavior, not intent**: Cross-check each claim in commit messages against the actual diff. Run the affected tests (`go test`, `pytest`, `npm test`) and at least one static analyzer on the touched packages.
5. **Check the four silent-failure patterns**: weakened acceptance criteria, widened permissions, deleted or skipped tests, and new unpinned external dependencies.
6. **Issue a deterministic verdict**: `APPROVE` (no blockers/majors), `REQUEST CHANGES` (blockers/majors with file:line references), or `ESCALATE` (risk exceeds reviewer authority, e.g., cryptographic or data-loss surface).
7. **Re-verify fixes**: Re-run failing checks against the fix commit; never approve on prose promises. Record the full trail in `templates/code-review-spec.md` format.

## Decision rules
- **Criteria first**: Findings that do not trace to acceptance criteria, policy, or a demonstrable defect are nits, never blockers.
- **Evidence over prose**: A claim without a file:line reference and reproduction is not a finding.
- **One blocker vetoes**: A single unresolved blocker forces REQUEST CHANGES regardless of change size or author seniority.
- **Never rewrite in review**: Propose the minimal corrective direction; do not author the fix inside the review.
- **Independence required**: The reviewer must not have authored the change; self-review is recorded as advisory only.

## Evidence required
- Completed review record following `templates/code-review-spec.md` with per-finding severity and verdict.
- Test and analyzer logs executed during the review.
- Re-verification log for every blocker/major after the fix commit.

## Output contract
- Review verdict (APPROVE / REQUEST CHANGES / ESCALATE) with severity-graded findings.
- File:line-anchored defect list with failure scenarios.
- Re-verification report closing each blocker and major.

## Stop conditions
- Verdict issued with all blockers and majors either resolved and re-verified, or explicitly waived by policy owner.
- Review scope exceeds authority (cryptographic, legal, data-governance) and is escalated with a written handoff.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to Lead Architect when acceptance criteria themselves are ambiguous or contradictory.
- Escalate immediately on discovering intentionally concealed defects, credential material in diffs, or suspected supply-chain tampering.
