# Prumo Canonical Navigation & Authority Resolution

## Purpose
Navigate Prumo project documentation and resolve canonical information fast using the authority order, routing map, and source map — never guessing when a locked decision exists.

## Use when
- Answering "what is the approved design / policy / scope?" for any Prumo Goal.
- Starting implementation in an unfamiliar area and needing the smallest sufficient context.
- Resolving conflicts between documents (repo spec vs. agent adapter vs. Living Book vs. inference).
- Locating owners, schemas, ADRs, and acceptance criteria for a task.

## Do not use when
- Authoring new architecture or policy (use `architecture-quality` / ADR flow).
- Performing full-text research across external sources (use web search + `documentation-for-llms`).
- Editing docs structure or governance itself (requires canonical-document update process).

## Required context
- The question or Goal ID needing an authoritative answer.
- `docs/governance/authority.md` + `docs/AUTHORITY_MAP.json` (who owns what, canonical vs. projection).
- `docs/SOURCE_MAP.json` (local doc → Living Book page mapping).
- `AGENTS.md` routing table for entry points.

## Procedure
1. **Classify the question**: product, architecture, development process, runtime, security, migration, or governance. Route via the `AGENTS.md` table first (e.g., scope → `docs/product/scope-v0.4.md`, dependencies → `docs/architecture/dependency-rules.md`).
2. **Apply authority order**: `docs/governance/authority.md` is canonical; repository specs outrank the Notion Living Book; the Living Book outranks agent inference; external sources rank last. Cite the winning source.
3. **Resolve conflicts explicitly**: When two sources disagree, record both, apply the authority order, and flag drift if a projection contradicts its canonical source (projection must be fixed, never the reverse).
4. **Expand progressively**: Read the smallest sufficient section first (Lean Progressive Context). Fetch Living Book pages on demand only — one relevant page per question, never the whole book.
5. **Verify machine contracts**: When the answer involves a schema, manifest, or policy file, read the file itself (`schemas/`, `.prumo/repository/policy.json`, `prumo.json`) rather than quoting prose about it.
6. **Hand off with pointers**: Return the answer plus exact file paths and sections so the next agent can verify without re-navigating. Record the trail in `templates/prumo-navigation-spec.md` format for contested decisions.

## Decision rules
- **Canonical wins**: A canonical repository spec beats any adapter, summary, cache, or memory. Projections never override sources.
- **Cite or it didn't happen**: Every authoritative claim carries a file path + section (or page ID for Notion). Uncited claims are inference and must be labeled as such.
- **Smallest sufficient context**: Stop expanding once the question is answered with a cited source; do not preload adjacent docs "just in case".
- **Drift gets flagged**: A projection contradicting canonical content is reported as drift with both pointers; never silently reconciled.

## Evidence required
- Cited source paths for every authoritative claim.
- Navigation trail (route taken through routing table → doc → section).
- Drift report when canonical/projection conflicts are found.

## Output contract
- Direct answer with winning-source citations.
- Navigation trail a second agent can replay.
- Drift findings, if any, with both-side pointers.

## Stop conditions
- Question answered with cited canonical source.
- No canonical source exists and the gap is recorded as an open question with owner.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the documentation owner (per `AUTHORITY_MAP.json`) when canonical sources contradict each other.
- Escalate to a human when a locked decision must change — navigation never edits locked docs.
