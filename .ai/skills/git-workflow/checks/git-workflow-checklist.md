# Git Workflow — Verification Checklist

## Pre-Execution Gate
- [ ] Goal or Task is locked and measurable.
- [ ] Required inputs (System requirements and context) are available and schema-validated.
- [ ] Execution token and step budget are within bounded limits.

## Quality & Compliance Criteria
- [ ] Implementation adheres to Clean Code and explicit responsibility principles.
- [ ] No cyclic dependencies or layer boundary violations introduced.
- [ ] Zero secrets, private tokens, or sensitive credentials exposed.
- [ ] Error conditions are handled explicitly with actionable error context.

## Verification & Testing
- [ ] Unit tests pass deterministically (target: >=85% coverage for business logic).
- [ ] Static analysis and formatting checks pass without warnings.
- [ ] Required evidence (test, review) has been generated and recorded.

## Sign-Off
- [ ] Task acceptance criteria verified.
- [ ] Evidence appended to task report / project intelligence.
