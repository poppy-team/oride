# Rust Analyzer Context Indexing — Verification Checklist

## 1. Crate Graph & Workspace Loading
- [ ] `cargo metadata --format-version 1 --no-deps` succeeds and its output is archived.
- [ ] rust-analyzer reaches ready status with `cargo.targetDir` set; version is recorded.
- [ ] Every workspace crate appears in the crate graph; missing crates fail the run loudly.

## 2. Symbol & Relation Extraction
- [ ] Bulk pass uses `workspace/symbol` capped at 500 results per crate; per-file detail uses `textDocument/documentSymbol`.
- [ ] Traits, impl blocks, structs, enums, functions, and macros are all represented.
- [ ] Hover signatures, definition targets, and references with declarations are stored per symbol.

## 3. Trait Edges & Macro Policy
- [ ] Every indexed trait lists its known implementations; traits without impl edges are flagged incomplete.
- [ ] `macro_rules!` and derive-generated items are indexed at expansion sites with the generating macro named.

## 4. Freshness & Flycheck Gating
- [ ] Flycheck runs `cargo check --workspace --message-format=json` and diagnostics flow via push.
- [ ] Files with errors serve last-good symbols only with an explicit stale marker.
- [ ] A single-file save propagates to fresh results in under 4 seconds through salsa revisions.

## 5. Ranking & Verification Evidence
- [ ] Ranking applies exact-path, same-crate, dependency-distance, and recency order.
- [ ] All query paths enforce the 40-symbol and 12000-token caps.
- [ ] Golden recall is 100% including generics, trait objects, async fns, and macros.
- [ ] Query p95 under 300 ms warm and `scripts/verify.sh` exits 0.
