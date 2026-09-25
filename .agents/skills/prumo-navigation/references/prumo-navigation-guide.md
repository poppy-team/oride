# Prumo Navigation Reference Guide

## 1. Authority Order (Memorize)

1. Canonical repository specs, schemas, and ADRs (`docs/`, `schemas/`).
2. Accepted local engineering documentation.
3. Prumo Living Book in Notion (approved design, on demand).
4. Agent inference (labeled as such).
5. External sources (last resort).

Once a design decision is promoted into a canonical repository specification, the repository version outranks the Notion version for implementation.

## 2. Routing Table (Primary Entries)

| Need | Start here |
|---|---|
| Current build target | `docs/product/vision.md` |
| Implementation plan | `docs/development/waves.md` |
| Open gaps | `docs/harness/gap-register.md` |
| Doc ownership / canonical status | `docs/governance/authority.md`, `docs/AUTHORITY_MAP.json` |
| Scope in/out | `docs/product/scope-v0.4.md` |
| Package dependencies | `docs/architecture/dependency-rules.md` |
| Repo layout | `docs/development/repository-layout.md` |
| Coding style | `docs/development/coding-standards.md` |
| Testing | `docs/development/testing-strategy.md` |
| Trust model | `docs/security/trust-model.md` |
| Doc provenance | `docs/SOURCE_MAP.json` |
| Repo mutation rules | `.prumo/repository/policy.json` |

## 3. Source Roles

- **Canonical**: source of truth; edits follow the governance process.
- **Projection**: generated view (adapters, indexes, capsules); regenerated, never hand-edited for facts.
- **Historical**: superseded; readable for context, never authoritative.
- **Cache/runtime**: derived, disposable; `runtime/cache` must never become canonical.

## 4. Worked Example

Question: "Can the core framework require a specific model provider?"
Route: governance → `AGENTS.md` core policy ("Do not make any orchestrator/model provider mandatory for the core framework") → `docs/governance/authority.md` for ownership. Answer: no, with two citations. Total reads: 2 files. A Living Book fetch is unnecessary because a canonical local source answers it.

## 5. Anti-Patterns

- Loading the entire Living Book to answer one question.
- Quoting an agent adapter as if it were project policy.
- Treating `runtime/cache` capsules as citable sources.
- Answering from memory when a locked spec exists and is readable.
