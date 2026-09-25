# Documentation Engineering & Canonical Markdown Publishing

## Purpose
Maintain canonical Markdown documentation for users, developers, and operators — with explicit ownership, authority roles, drift detection against code, link integrity, and versioned publishing — so docs stay synchronized with implementation instead of decaying beside it.

## Use when
- Creating or updating user guides, API references, runbooks, architecture notes, or ADRs.
- Auditing documentation health: stale pages, broken links, drift from code, missing ownership.
- Publishing versioned docs: changelog entries, release notes, migration guides tied to a release.
- Defining documentation authority: which pages are canonical, which are projections, and how drift is resolved.

## Do not use when
- Writing code, tests, or pipeline definitions (use the domain skill; docs describe the work, not replace it).
- Generating changelogs mechanically from commits during a release (use `release-engineering`; this skill owns prose quality).
- Authoring Prumo Goal criteria or phase plans (use `goal-management`).
- Storing generated API dumps or coverage reports as hand-maintained prose (generated artifacts belong in build output, not in docs).

## Required context
- Documentation scope: audience (users, developers, operators), pages in scope, and their authority roles (canonical, projection, historical).
- Source of truth links: code modules, schemas, or configs each page documents, so drift can be checked mechanically.
- Publishing setup: docs site toolchain (Astro 4.16 with Starlight, or MkDocs 1.6), link checker (lychee 0.15), target version branch.
- Ownership map: named owner per page plus review cadence (e.g. runbooks reviewed quarterly).

## Procedure
1. **Scope the change with authority roles**:
   - Classify each touched page: canonical (normative, overridden only by ADR), projection (derived view, regenerated on change), historical (frozen, never edited except for deprecation banners).
   - Confirm ownership: every page lists an owner and last review date in frontmatter; ownerless pages get an owner before content changes.
   - State the audience and reading goal at the top of user-facing pages (`Who: billing operators. Goal: restore service in under 15 minutes.`).
2. **Write small, verifiable, example-first docs**:
   - Lead with a runnable example, then explain: every procedure shows exact commands with realistic values (`./scripts/migrate --to 48 --target staging`, expected `4 min 12 s`).
   - Keep pages under 150 lines; split longer guides and link explicitly rather than scrolling past reader attention budgets.
   - Use absolute include paths and pinned versions in examples (`golang:1.24.3-bookworm`, k6 0.57.2); floating references rot first.
3. **Check drift against implementation**:
   - For each documented command, flag, config key, and default, verify against the current code (`grep -r "max-tier" internal/billing/`, compare with the documented default 8).
   - Verify documented file paths exist (`test -f` each one) and documented ports, endpoints, and env vars match the running service.
   - Record drift findings as defects with owners and fix-by dates; docs that contradict code are bugs, not opinions.
4. **Validate links, formatting, and build**:
   - Run `lychee --no-progress docs/**/*.md` (zero broken links required) and the site build (`npm run build` in `docs-site/`, Astro 4.16) from a clean checkout.
   - Enforce Markdown lint (`markdownlint-cli2 "docs/**/*.md"`) with the repo config; fix violations, do not bulk-disable rules.
   - Preview versioned output: the changed pages render under both `latest` and the pinned release version without layout breakage.
5. **Publish and schedule the next review**:
   - Merge docs with the code they describe whenever possible (docs-then-code for new flags; same PR for behavior changes).
   - Update the page's last-review date and set the next review per cadence (runbooks quarterly, API references per release, vision yearly).
   - Run the verification script `scripts/verify.sh` from the repo root; it must exit 0.

## Decision rules
- **Docs That Contradict Code Are Bugs**: Drift findings get defect tracking with owners and dates, never `docs-todo-someday` limbo.
- **Example First, Theory Second**: Procedures without runnable examples fail review; prose-only runbooks are rejected.
- **Ownership Is Mandatory**: No page ships or updates without a named owner and a last-review date in frontmatter.
- **Small Pages Win**: Pages over 150 lines are split; link explicitly instead of burying the reader.
- **Generated Output Is Not Documentation**: Coverage dumps, API JSON, and SBOMs are linked as artifacts, never pasted as prose.

## Evidence required
- Documentation specification adhering to `templates/documentation-spec.md` with page roles, drift findings, and review dates.
- Drift check log: each documented command, flag, path, and default verified against current code.
- `lychee` link report with zero broken links and the clean site-build log.
- Markdown lint log clean under the repo config.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Updated pages with frontmatter (owner, last review, authority role), runnable examples, and verified links.
- Drift report: findings with owners and fix-by dates, or an explicit clean bill.
- Publishing record: built site version, preview links, version branches updated.
- Review schedule: next review dates per touched page.

## Stop conditions
- All touched pages accurate against current code with zero broken links and a clean site build.
- Drift findings fixed or tracked as defects with owners and fix-by dates.
- Owners and next review dates recorded for every touched page.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the owning team if drift reveals the code is wrong (not the docs); stop editing prose and file the code defect.
- Escalate to Lead Architect if two canonical pages contradict each other; do not pick a winner unilaterally.
- Escalate to Release Captain if release notes or migration guides block a dated release train.
