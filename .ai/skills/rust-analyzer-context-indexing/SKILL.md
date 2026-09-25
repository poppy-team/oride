# rust-analyzer Context Indexing

## Purpose
Build and query code-context indexes for Rust workspaces through rust-analyzer: Cargo metadata-driven crate graph loading, LSP symbol, hover, definition, and reference extraction across crates, traits, and macros, flycheck-gated freshness, and ranked query answers bounded for AI agent consumption.

## Use when
- An agent needs accurate Rust definitions, trait implementations, macro expansions, or cross-crate references from a Cargo workspace.
- rust-analyzer rolling 2026-09-01 or newer plus a working `cargo metadata` invocation are available.
- Index freshness must survive file edits via rust-analyzer incremental salsa revisions without full workspace reloads.
- Query answers must fit a fixed token budget with deterministic ranking.

## Do not use when
- The target language is not Rust; use the matching engine skill instead.
- Only lexical search over `.rs` files is required with no need for trait resolution, generics, or macro expansion.
- The workspace cannot run `cargo metadata` successfully, for example missing vendored registries with no network access and no vendor directory.

## Required context
- Workspace root, `cargo metadata --format-version 1` output, rust-analyzer version, and sysroot source for std symbols.
- Token budget for a single query answer: at most 40 symbols and 12000 tokens of emitted context.
- Performance targets: workspace load time, query p95 latency, and the golden symbol recall set.
- Flycheck command used for diagnostics, normally `cargo check --workspace --message-format=json`.

## Procedure
1. **Load the crate graph**: run `cargo metadata --format-version 1 --no-deps` to enumerate packages and targets, then start rust-analyzer with `cargo.targetDir` set and confirm the status notification reaches ready. Abort indexing if metadata fails; an empty symbol set must never masquerade as an index.
2. **Extract symbols per crate**: request `workspace/symbol` with an empty query capped at 500 results per crate for the bulk pass, then `textDocument/documentSymbol` per file for hierarchical detail including traits, impl blocks, structs, enums, functions, and macros.
3. **Resolve relations**: map `textDocument/hover` for type signatures, `textDocument/definition` for jump targets, and `textDocument/references` with `includeDeclaration: true` for call sites. Record trait-to-impl edges explicitly since they are the highest-value Rust context.
4. **Gate freshness on flycheck**: subscribe to `textDocument/publishDiagnostics` fed by `cargo check --workspace --message-format=json`; documents with errors keep their last-good symbols flagged stale, and salsa revisions advance the index incrementally on each saved change.
5. **Serve ranked queries**: expose `symbols --name 'ledger::apply_entry'`, `impls --trait 'LedgerStore'`, and `refs --symbol <path>` with ranking by exact path match, then same-crate proximity, then dependency-distance hops, then recency, truncating at the 40-symbol and 12000-token caps.
6. **Verify against goldens**: run `scripts/verify.sh`, assert query p95 under 300 ms warm, and require 100% recall on the golden set covering generics, trait objects, async fns, and macro-generated items.

## Decision rules
- **Metadata first**: no `cargo metadata`, no index; symbol extraction against a failed crate graph is forbidden.
- **Stale flags are explicit**: symbols from files with flycheck errors are served only with a stale marker, never as fresh truth.
- **Trait edges are mandatory**: every indexed trait stores its known implementations; a trait without impl edges is an incomplete record.
- **Macro expansions recorded**: `macro_rules!` and derive-generated items are indexed at their expansion sites with the generating macro named.
- **Bounded answers always**: every query path enforces the symbol and token caps even when the caller requests unlimited results.

## Evidence required
- Index specification following `templates/rust-analyzer-context-indexing-spec.md` with metadata output, versions, and flycheck command.
- Crate list with per-crate symbol counts and trait-to-impl edge counts.
- Golden recall log with measured query p95 latency.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Loaded rust-analyzer workspace with ready status and crate graph report.
- Queryable symbol, reference, trait-implementation, and hover-signature index.
- Freshness report showing stale-flagged files and salsa revision watermark.

## Stop conditions
- Golden symbol recall reaches 100% with query p95 under 300 ms on a warm workspace.
- Single-file save propagates through salsa to fresh query results in under 4 seconds.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the Rust build owner if `cargo metadata` or flycheck fails due to registry access, missing toolchains, or broken feature unification.
- Escalate to the lead architect if proc-macro expansion is nondeterministic or too slow to index within budget, so expansion can be scoped or cached explicitly.
