# Project Documentation Architect

Build and maintain implementation-ready, living project documentation —
product, architecture, contracts, evidence, quality gates — in canonical
`en-US` plus synchronized `pt-BR`, with the least workforce sufficient.
Portable core: `~/project-documentation-architect/` (agentskills.io format);
this package is the Prumo-native adapter of that core.

## Purpose
Turn `idea -> context -> requirements -> research -> decisions ->
contracts -> documentation -> readiness -> execution -> evidence ->
delta -> continuous maintenance`, producing a versioned, evidence-backed
documentation system (not a pile of Markdown) that lets future agents work
without rediscovering everything.

## Use when
Creating, initializing, bootstrapping, adopting, or formally planning a
software project before substantial implementation begins; adopting an
existing repo that lacks adequate docs; maintaining docs of an already
documented project (delta); or when asked for an implementation bible,
living book, full architecture, or agent handoff documentation.

## Do not use when
Single-function bug fixes, trivial edits inside an already documented
project, explicitly disposable one-off scripts, or pure code review with
no documentation mandate. For small changes inside documented projects,
run Mode DELTA only.

## Required context
- User intent or briefing (GENESIS) or repo snapshot (ADOPTION/DELTA).
- Detected documentation profiles (see `references/profiles.md`).
- `docs/project-docs.json` manifest when it already exists; create it first
  when it does not (template: `templates/project-docs.json`).

## Procedure
1. Select mode: GENESIS (no trusted docs yet), ADOPTION (repo exists, docs
   missing/partial/untrusted — inspect code, tests, CI, config, runtime
   first, then write a Gap Matrix and fix only the delta), DELTA (docs exist
   — update only impacted capabilities, contracts, tests, docs, risks,
   migrations, translations). Detail: `references/modes.md`.
2. Detect profiles and resolve the mandatory contract pack per profile
   (`references/profiles.md`). Write/update `docs/project-docs.json` first:
   identity, profiles, contracts, semantic roles, statuses, owners.
3. Select the least workforce sufficient (`references/orchestration.md`):
   tiny CLI = orchestrator + architect + one documenter + translator;
   add product, research, UI/UX, a11y, security, performance, dependency,
   ADR, traceability, parity-review, audit roles only as profiles demand.
   Never assign two writers to the same canonical artifact concurrently.
4. Draft canonical `en-US` docs, then translate impacted pairs to `pt-BR`,
   then parity-review (`references/bilingual.md`). Technical errors found
   during translation are fixed in `en-US` first and re-propagated — never
   patch `pt-BR` alone.
5. Run `scripts/verify.sh`, complete `checks/post-flight.md`, and record a
   documentation delta (changed capabilities/contracts, impacted
   docs/tests/translations, severity, requires_review).

## Decision rules
- Never invent features, requirements, commands, APIs, benchmarks, files,
  or decisions. Unknowns become OPEN, ASSUMPTION, RESEARCH REQUIRED, or
  BLOCKING QUESTION.
- Material claims need `claim -> source -> evidence -> verification ->
  decision`. `verified`, `safe`, `complete`, and `done` are forbidden
  without verifiable semantics.
- Depth follows `complexity + risk + impact + irreversibility`. No empty
  checklist files; group content for simple projects.
- Human-approved decisions outrank inference; real behavior + valid tests
  + approved contracts outrank unverified assertions.
- Contradictions are preserved as CONTRADICTION records (both claims plus
  evidence) and resolved via the authority model or the human — never
  merged silently.

## Evidence required
- Automated test runs, build outputs, or verification logs where the skill
  touches behavior; `scripts/verify.sh` output for documentation runs.
- Every normative MUST links to an acceptance criterion plus a test pointer.
- Required evidence for sign-off: `review` (and `test` when implementation
  or executable checks are involved).

## Output contract
- Updated `docs/project-docs.json` manifest.
- Canonical `en-US` documents plus synchronized `pt-BR` versions for every
  required pair (states CURRENT or IN_REVIEW, never silent STALE).
- Documentation delta record plus handoff microcontexts (one loadable
  context per work stream — never a full dump).
- Completed `checks/post-flight.md`; residual risks and open questions
  recorded explicitly.
