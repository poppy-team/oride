# Interaction Research

## Purpose
Research interaction models, state transitions, micro-animations, and gestures.

## Use when
- Active Goal or Task explicitly requires interaction research operations.
- Operating in mode(s): `research`.

## Do not use when
- Task is out of scope or unrelated to interaction research.
- Bounded token budget or capability policy denies required operations.

## Required context
- Product requirements
- Design tokens
- Wireframe / UI view

## Procedure
1. Analyze UX requirements and design constraints for interaction research.
2. Check compliance with design tokens and accessibility guidelines.
3. Specify structured layout, interaction states, and responsive breakpoints.
4. Verify with automated visual or keyboard navigation checks.
5. Emit validated specification or review findings.

## Decision rules
- Always prefer smallest sufficient changeset over broad refactoring.
- Preserve existing public contracts and architecture invariants.
- Never weaken locked Goal acceptance criteria.

## Evidence required
- test
- review

## Output contract
- Design specification / findings
- Accessibility audit scorecard
- UI tests

## Stop conditions
- Task acceptance criteria satisfied with evidence
- Token budget exhausted
- Blocked on external dependency

## Escalation rules
- Escalate to lead architect or human if locked Goal criteria cannot be met.
- Escalate immediately upon discovering unexpected security vulnerabilities or data loss risks.
