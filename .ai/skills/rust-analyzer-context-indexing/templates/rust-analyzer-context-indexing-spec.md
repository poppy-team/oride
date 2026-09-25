# rust-analyzer Context Index — Deliverable Specification

## 1. Workspace & Toolchain
- **Workspace**: Northstar ledger at /srv/northstar, 22 crates, 640 Rust files, 188,000 lines
- **rust-analyzer version**: rolling release 2026-09-01 over stdio
- **Toolchain**: rustc 1.83.0, sysroot sources present for std resolution
- **Metadata**: `cargo metadata --format-version 1 --no-deps` succeeded 2026-09-15, archived at .prumo/index/cargo-metadata.json
- **Flycheck**: `cargo check --workspace --message-format=json`, 6-minute full pass, incremental under 20 seconds

## 2. Extraction Coverage
- **Symbols stored**: 34,908 rows across 22 crates with path, kind, file, and line
- **Document symbols**: per-file hierarchies for all 640 files including 412 impl blocks
- **Trait edges**: 186 traits mapped to 521 implementations, zero traits without impl edges
- **Macro items**: 96 macro-generated items indexed at expansion sites with generating macro named

## 3. Freshness Policy
- **Engine**: salsa revisions on save, watermark recorded per query answer
- **Stale rule**: files with flycheck errors serve last-good symbols flagged stale; 4 files stale at report time, listed at .prumo/index/rust-stale.log
- **Measured propagation**: single-file save in ledger-core fresh in 2.6 seconds

## 4. Query API & Ranking
- **Commands**: `symbols --name 'ledger::apply_entry'`, `impls --trait 'LedgerStore'`, `refs --symbol 'ledger::apply_entry'`
- **Ranking weights**: exact path match 1.0, same-crate proximity 0.6, dependency-distance hops 0.3, recency 0.1
- **Caps enforced**: 40 symbols and 12000 tokens per answer on every code path

## 5. Measured Results
- **Workspace load wall clock**: 2 min 5 s cold including flycheck seed
- **Query p95 warm**: 211 ms over 400 sampled queries
- **Golden recall**: 150 of 150 golden symbols resolved, 100%, including 18 generic instantiations, 12 trait objects, 9 async fns, and 14 macro-generated items

## 6. Verification Evidence
- Crate graph report at .prumo/index/rust-crates-2026-09-15.log, 22 of 22 crates
- Recall log at .prumo/index/rust-recall-2026-09-15.log
- `scripts/verify.sh` exit code 0 on 2026-09-15
