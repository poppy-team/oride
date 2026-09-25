# Documentation Engineering Reference Guide

## 1. Core Concepts & Models

### 1.1 Authority Roles: Canonical, Projection, Historical
Every page holds exactly one role. Canonical pages are normative: code and other
docs conform to them, and changing them requires an ADR. Projection pages are
derived views (dashboards, indexes, generated summaries) that are regenerated,
never hand-edited against their source. Historical pages are frozen records;
edit them only to add deprecation banners. Drift resolution order: canonical
beats projection, current code beats stale prose, ADR beats both on intent.

### 1.2 Drift Half-Life
Undocumented-coupled prose decays with every code change touching its subject:

$$P(\text{accurate} \mid n \text{ changes since review}) \approx (1 - d)^n$$

With per-change drift probability $d \approx 0.15$ for CLI flags and defaults, a
page is more likely wrong than right after 5 unreviewed changes. Counter with
review cadences matched to churn (runbooks quarterly, API references per
release) and same-PR docs for behavior changes, which reset $n$ to zero.

### 1.3 The Example-First Contract
Readers under incident load do not parse theory; they copy commands. A procedure
is complete when its commands run verbatim on a clean machine and produce the
stated output. Every example therefore carries: exact invocation, realistic
values (migration 47 to 48, staging target, 4 min 12 s expected), and the
success signal to look for. Prose explains only what the example cannot show.

### 1.4 Link Rot Budget
External links die at roughly 5 percent per year; internal links die on every
rename. `lychee` on every docs change plus a weekly scheduled run bounds rot:
zero broken links at merge, and the weekly run files defects (not silent decay)
for anything that died upstream. Prefer links to pinned versions over `latest`
where the content must remain stable.

## 2. Patterns and Anti-Patterns

**Do: frontmatter with owner, date, role.**
`owner: Marina Duarte, last-review: 2026-09-10, role: canonical` makes staleness
queryable and responsibility undeniable. Ownerless pages are where drift hides.

**Do: docs in the behavior PR.**
The PR that adds `--dry-run` updates the CLI reference in the same diff. Reviewers
verify prose against code while both are in front of them; later is never.

**Do not: paste generated output as prose.**
A 400-line API dump pasted into a guide rots on the next release and buries the
3 lines a human needs. Link the artifact; write the 3 lines.

**Do not: write `TODO: document this`.**
An acknowledged gap with no owner and no date is a permanent gap with good
intentions. File the docs defect with an owner and a fix-by version, or delete
the section and admit the gap in the page header.

**Do not: duplicate content across pages.**
Two copies of the flag table diverge within a quarter. Write once canonically,
link everywhere else; projections regenerate, they do not retype.

## 3. Tooling Configuration Example

```yaml
# .markdownlint-cli2.jsonc — prose gates for the docs tree
{
  "config": {
    "default": true,
    "line-length": { "line_length": 120, "tables": false, "code_blocks": false },
    "no-trailing-spaces": true,
    "no-multiple-blanks": true,
    "single-title": true
  },
  "globs": ["docs/**/*.md"]
}
```

```bash
# Docs validation sequence from a clean checkout
markdownlint-cli2 "docs/**/*.md"
lychee --no-progress "docs/**/*.md"
npm run build --prefix docs-site   # Astro 4.16 + Starlight, zero errors required
# verify drift for one documented default (example: billing tier cap)
grep -rn "max-tier\|MaxTier" internal/billing/ | head -5
```

```yaml
# Page frontmatter contract (Astro Starlight)
---
title: Billing Runbook
owner: Marina Duarte
last-review: 2026-09-10
next-review: 2026-12-10
role: canonical
audience: billing operators
goal: restore invoice rendering in under 15 minutes
---
```
