# Documentation Engineering — Verification Checklist

## 1. Authority & Ownership
- [ ] Each touched page carries an authority role: canonical, projection, or historical.
- [ ] Frontmatter lists a named owner and last-review date; no ownerless pages changed.
- [ ] User-facing pages state audience and reading goal in the first 5 lines.
- [ ] Canonical contradictions (two normative pages disagreeing) escalated, never silently resolved.

## 2. Example-First Prose Quality
- [ ] Every procedure leads with runnable commands using realistic values and expected outputs.
- [ ] Examples pin versions and absolute paths (`golang:1.24.3-bookworm`, k6 0.57.2); no floating references.
- [ ] Pages stay under 150 lines; longer guides split with explicit cross-links.
- [ ] No generated dumps pasted as prose; build artifacts linked, not embedded.

## 3. Drift Against Implementation
- [ ] Each documented command, flag, config key, and default verified against current code (grep logs attached).
- [ ] Every documented file path exists (`test -f` per path); ports, endpoints, and env vars match the running service.
- [ ] Drift findings filed as defects with owners and fix-by dates, or explicit clean bill recorded.
- [ ] Docs merged alongside the code they describe (same PR for behavior changes).

## 4. Links, Lint & Build
- [ ] `lychee --no-progress docs/**/*.md` reports zero broken links, internal and external.
- [ ] `markdownlint-cli2 "docs/**/*.md"` clean under the repo config; no bulk rule disables added.
- [ ] Site builds from a clean checkout (`npm run build` in `docs-site/`, Astro 4.16) with zero errors.
- [ ] Changed pages render correctly under both `latest` and the pinned release version.

## 5. Review Cadence & Sign-Off
- [ ] Last-review dates updated on every touched page; next reviews scheduled (runbooks quarterly, API refs per release).
- [ ] Documentation specification follows `templates/documentation-spec.md` with roles and drift results.
- [ ] Stale pages (past review date) reaffirmed, updated, or deprecated with a banner — none ignored.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
