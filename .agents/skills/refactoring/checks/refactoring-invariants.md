# Refactoring — Verification Checklist

## 1. Behavior Baseline
- [ ] Baseline test suite passes before any transformation begins.
- [ ] Missing coverage locked with characterization (golden-master) tests for affected behavior.
- [ ] Public API contracts and observable behavior documented as the invariant to preserve.

## 2. Transformation Discipline
- [ ] Transformations applied in small, single-smell steps (one Extract Method / Rename / Inline per commit).
- [ ] Automated IDE refactoring tools used where available instead of manual structural edits.
- [ ] Each step re-runs the suite immediately; failures trigger revert-and-split, never fix-forward on red.

## 3. Code-Quality Targets
- [ ] Targeted smells resolved: duplication removed, long methods split (≤ ~30 lines guide), magic numbers named.
- [ ] No new public API surface or signature changes introduced by the refactoring.
- [ ] Complexity (cyclomatic) reduced or held flat on touched modules.

## 4. Regression & Evidence
- [ ] Full suite green after every committed step; golden-master outputs unchanged.
- [ ] Behavior-diff proof recorded (before/after outputs identical on fixture inputs).
- [ ] `scripts/verify.sh` exits 0 from the repository root.

## 5. Handoff
- [ ] Refactoring change log lists each transformation with rationale and verification commit.
- [ ] Residual smells outside scope noted as follow-ups, not silently expanded into this task.
