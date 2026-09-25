# Project Intelligence

## Purpose
Record honest task-level cost, effort, quality, and token metrics in a durable intelligence ledger, then use estimate-vs-actual variance to calibrate future planning instead of relying on gut feel.

## Use when
- Closing out a Task or Goal that must leave behind cost, duration, token spend, and defect metrics.
- Estimating a new Task by analogy to measured historical tasks rather than guessing.
- Auditing planning accuracy across a wave or release (estimate drift, systematic underestimation).
- Detecting quality signals early: reopen rates, review rounds, evidence failures per task.

## Do not use when
- Selecting which context to load for the current task (use `lean-progressive-context`).
- Compressing already-loaded context (use `context-optimization`).
- Writing people-performance reviews; this skill measures tasks, never individuals.

## Required context
- Task ledger location (default evidence/intelligence.jsonl) and schema version in use.
- Estimate baseline for the task: hours, tokens, and complexity class recorded before work starts.
- Quality signals available: test results, review rounds, reopen counts, evidence gate outcomes.
- Historical ledger entries for analogous tasks used in calibration.

## Procedure
1. **Record the estimate before starting**: log estimated hours (e.g. 6), estimated tokens (e.g. 25,000), and complexity class (S/M/L) with the task id and date 2026-09-23; estimates written after the fact are rejected.
2. **Instrument the run**: capture wall-clock duration, measured token spend from the ledger, test pass/fail counts, and number of review rounds as the work proceeds.
3. **Close the ledger entry at completion**: append one JSONL line with estimate vs actual for hours and tokens, defect count found post-merge within 7 days, and reopen flag.
4. **Compute variance**: variance_pct = (actual - estimate) / estimate * 100 per task; aggregate by complexity class weekly to find systematic bias (e.g. class M tasks average +38% hours).
5. **Calibrate the next estimate**: multiply the raw guess by the class calibration factor (1 + mean variance) and record both raw and calibrated numbers.
6. **Run `scripts/verify.sh`** to confirm schema validity and variance computation before publishing the report.

## Decision rules
- **Estimates precede work**: an estimate recorded after completion is marked void, never backfilled.
- **Actuals are measured, not recalled**: durations and token counts come from logs, never from memory.
- **Variance is reported with sign**: underruns (-15%) and overruns (+40%) are both published; hiding underruns corrupts calibration equally.
- **Class-based calibration only**: adjust using the complexity class mean, never a single outlier task.
- **Tasks, not people**: ledger entries carry task ids and classes; attributing variance to individuals is prohibited.

## Evidence required
- JSONL ledger entry for the task conforming to the intelligence schema.
- Variance computation showing estimate, actual, and signed percentage per metric.
- Calibration factor applied to the next estimate with its historical basis.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Closed ledger entry with estimate-vs-actual metrics and quality signals.
- Variance report aggregated by complexity class.
- Calibrated estimate for the next analogous task.

## Stop conditions
- Ledger entry closed, variance computed, and calibration applied with evidence.
- Historical sample too small (fewer than 5 tasks per class); report uncalibrated with a warning.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the planning owner if systematic bias exceeds +50% for two consecutive weeks; the estimation model needs revision, not per-task excuses.
- Escalate to a human if ledger data reveals a security incident or data-loss event pattern.
