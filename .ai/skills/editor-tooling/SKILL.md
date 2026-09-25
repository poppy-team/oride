# Editor Tooling & UI

## Purpose
Implement desktop engine editors, scene inspectors, and gizmo overlays.

## Use when
- Active Goal or Task explicitly requires editor tooling & ui operations.
- Operating in mode(s): `implementation`.

## Do not use when
- Task is out of scope or unrelated to editor tooling & ui.
- Bounded token budget or capability policy denies required operations.

## Required context
- Engine architecture specification
- Target hardware / GPU constraints
- Benchmark fixtures

## Procedure
1. Review memory and performance constraints for editor tooling & ui.
2. Implement data-oriented, cache-friendly data structures.
3. Establish deterministic visual fixtures or benchmark baselines.
4. Verify zero memory leaks, GPU synchronization correctness, and frame budgets.
5. Emit benchmark and test evidence.

## Decision rules
- Always prefer smallest sufficient changeset over broad refactoring.
- Preserve existing public contracts and architecture invariants.
- Never weaken locked Goal acceptance criteria.

## Evidence required
- test
- benchmark

## Output contract
- Optimized engine subsystem
- Deterministic benchmark evidence
- Visual test fixtures

## Stop conditions
- Task acceptance criteria satisfied with evidence
- Token budget exhausted
- Blocked on external dependency

## Escalation rules
- Escalate to lead architect or human if locked Goal criteria cannot be met.
- Escalate immediately upon discovering unexpected security vulnerabilities or data loss risks.
