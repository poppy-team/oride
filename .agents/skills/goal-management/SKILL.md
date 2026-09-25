# Prumo Goal Authoring, Phasing & Lock Management

## Purpose
Author and manage Prumo Goals with measurable acceptance criteria, phased execution plans, evidence linkage, and lock discipline — so scope is explicit, progress is verifiable, and locked criteria can never be silently weakened.

## Use when
- Authoring a new Goal: title, context, acceptance criteria, phase breakdown, evidence plan.
- Phasing execution: sequencing slices, assigning owners, defining per-phase exit gates.
- Locking criteria: freezing acceptance terms before implementation begins.
- Auditing Goal health: stale locks, unmapped criteria, evidence gaps, scope drift.

## Do not use when
- Writing application code toward a Goal (use the domain skill; this skill manages the Goal itself).
- Running git branching or pull request mechanics (use `git-workflow`).
- Measuring performance or auditing complexity inside a Goal's tasks (use `benchmarking`, `code-quality`).
- Weakening a locked criterion because implementation is hard — that requires a recorded re-scope decision, never a quiet edit.

## Required context
- Goal statement in one sentence plus the problem it solves and the explicit non-goals.
- Acceptance criteria as measurable predicates (numbers, named tests, named gates), each with its evidence type.
- Phase list with owners, exit gates, and dependencies between phases.
- Lock state: which criteria are locked, by whom, and at which commit or timestamp.

## Procedure
1. **Author the Goal with measurable criteria**:
   - Write the title as an outcome (`Tiered pricing computes within 0.01 of golden fixtures`), not an activity (`Work on pricing`).
   - Express every acceptance criterion as a falsifiable predicate: `p99 render latency at most 120 ms at 500 rps measured by k6 run load/invoice-endpoint.js`, never `fast enough`.
   - List non-goals explicitly (e.g. multi-currency support excluded; PDF layout untouched) to bound scope disputes later.
2. **Phase the work into gated slices**:
   - Split into phases of 1–3 days each with a single owner: Phase 1 data model, Phase 2 computation, Phase 3 API exposure, Phase 4 rollout.
   - Define each phase's exit gate in the same measurable style (`Phase 2 exits when TestCalculateTieredPrice 18/18 green plus benchstat delta beyond noise`).
   - Record phase dependencies (Phase 3 needs Phase 1's struct; Phase 4 needs the release-engineering rollout plan) so blocked phases are visible, not surprising.
3. **Lock criteria before implementation**:
   - Freeze acceptance criteria with `prumo goal lock billing-tiered-pricing --criteria AC1,AC2,AC3,AC4`; record the lock commit and timestamp in the Goal file.
   - After locking, any criterion change requires a re-scope entry: reason, authorizer, and impact on completed phases. Silent edits to locked lines are prohibited.
   - Verify lock integrity in review: `git diff <lock-commit>..HEAD -- goals/billing-tiered-pricing.md` must show changes only in status and evidence sections.
4. **Link evidence continuously**:
   - Append evidence per criterion as work lands: test run IDs, benchmark summaries, review links (PR billing-api#482 for AC1–AC3).
   - Mark a criterion satisfied only when its named evidence exists and passes; partial evidence keeps the criterion open with a dated note.
   - Run the Goal health check weekly: every open criterion has an owner and a next action dated within 7 days, or it is flagged stale.
5. **Close or re-scope explicitly**:
   - Close the Goal when all criteria are satisfied with linked evidence; record the closing commit and a one-paragraph outcome note.
   - If scope must change, file the re-scope entry first, update phases and owners, then continue — never edit locked criteria mid-flight without the entry.
   - Run the verification script `scripts/verify.sh` from the repo root; it must exit 0.

## Decision rules
- **Criteria Are Predicates or They Do Not Exist**: Unmeasurable criteria (`robust`, `clean`, `fast`) are rewritten before the Goal is accepted.
- **Locked Means Append-Only**: Locked criteria lines change only via recorded re-scope entries; direct edits fail review.
- **No Orphan Criteria**: Every criterion has exactly one owner and one evidence type from authoring time; orphans block Goal activation.
- **Phases Exit Through Gates**: A phase is done when its exit gate passes, not when its timebox expires; expired timeboxes trigger re-planning, not silent carryover.
- **Stale Goals Get a Verdict Weekly**: Any Goal with no evidence update in 7 days is reaffirmed, re-scoped, or archived — never left ambiguous.

## Evidence required
- Goal specification adhering to `templates/goal-management-spec.md` with criteria, phases, locks, and evidence links.
- Lock record: criteria locked, authorizer, commit, and timestamp.
- Per-criterion evidence: test logs, benchmark summaries, review links as named in the Goal.
- Re-scope entries (if any) with reason, authorizer, and impact analysis.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Activated Goal file with measurable criteria, phased plan, owners, and lock record.
- Living evidence ledger: each criterion mapped to current evidence and status.
- Health report: stale flags, blocked phases with blockers named, next actions dated.
- Closure note or re-scope record when the Goal ends or changes.

## Stop conditions
- All acceptance criteria satisfied with linked passing evidence and the Goal formally closed.
- Goal re-scoped with recorded authorization and updated phases, owners, and locks.
- Goal archived as obsolete with rationale and superseding reference (if any).
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to Lead Architect if locked criteria conflict with each other or with architecture invariants discovered mid-flight.
- Escalate to the Goal sponsor if re-scope changes the promised outcome, timeline, or non-goals boundary.
- Escalate immediately if evidence contradicts a claimed satisfied criterion; the criterion reopens until new evidence passes.
