# Navigation Trail Record Specification Template

## 1. Question
- **Question**: Where is the caching skill's verification evidence defined?
- **Goal context**: Workforce placeholder remediation, skill `caching`.

## 2. Route Taken
1. `AGENTS.md` routing table → "How are we testing?" → `docs/development/testing-strategy.md`.
2. Skill manifest `required_evidence: ["test"]` + `SKILL.md` "Evidence required" section.
3. `scripts/verify.sh` exit-0 log (19/19 checks) as the evidence artifact.

## 3. Answer With Citations
- Evidence contract: `src/prumo/resources/workforce/skills/caching/SKILL.md` ("Evidence required").
- Machine check: `src/prumo/resources/workforce/skills/caching/scripts/verify.sh` (exit 0, 2026-09-23).
- No Living Book fetch needed — canonical local sources sufficed.

## 4. Conflicts / Drift
- None found. Catalog entry mirrors manifest (`scripts/verify.sh` check 3).

## 5. Expansion Accounting
- Files read: 3. Living Book pages fetched: 0. Stopping condition met at step 2.
