# Goal Management — Verification Checklist

## 1. Criterion Quality
- [ ] Goal title states an outcome, not an activity (`Tiered pricing computes within 0.01 of golden fixtures`).
- [ ] Every acceptance criterion is a falsifiable predicate with numbers, named tests, or named gates.
- [ ] No weasel criteria survive (`robust`, `clean`, `fast enough` rewritten or removed).
- [ ] Non-goals listed explicitly (multi-currency excluded, PDF layout untouched).

## 2. Phasing & Ownership
- [ ] Phases sized 1–3 days each with exactly one owner per phase.
- [ ] Each phase has a measurable exit gate in predicate style, not a date alone.
- [ ] Phase dependencies recorded (Phase 3 needs Phase 1's struct; Phase 4 needs the rollout plan).
- [ ] Blocked phases name their blocker and the unblocking owner, not just `blocked`.

## 3. Lock Discipline
- [ ] Acceptance criteria locked before implementation (`prumo goal lock` record with commit and timestamp).
- [ ] `git diff <lock-commit>..HEAD -- <goal-file>` shows changes only in status and evidence sections.
- [ ] Any locked-line change carries a re-scope entry with reason, authorizer, and impact on completed phases.
- [ ] No silent edits to locked criteria; violations reopen the Goal for review.

## 4. Evidence Linkage
- [ ] Every criterion maps to named evidence (test run, benchmark summary, review link) with current status.
- [ ] Criteria marked satisfied only with existing passing evidence; partial evidence keeps them open with dated notes.
- [ ] Weekly health check current: no open criterion without an owner and a next action dated within 7 days.
- [ ] Contradicted evidence reopens its criterion immediately; no satisfied flag survives a failing gate.

## 5. Closure & Sign-Off
- [ ] Goal specification follows `templates/goal-management-spec.md` with criteria, phases, locks, and links.
- [ ] Closure note recorded with closing commit and outcome paragraph, or re-scope/archive entry filed.
- [ ] Stale Goals (7 days without evidence update) reaffirmed, re-scoped, or archived with rationale.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
