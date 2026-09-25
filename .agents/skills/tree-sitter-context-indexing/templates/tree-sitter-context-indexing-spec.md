# Tree-Sitter Context Index — Deliverable Specification

## 1. Languages & Grammars
- **CLI version**: tree-sitter 0.24.7
- **Python grammar**: tree-sitter-python v0.23.4, ERROR rate 0 of 50 sampled files
- **TypeScript grammar**: tree-sitter-typescript v0.23.2, ERROR rate 1 of 50 sampled files, flagged file listed in section 6
- **Go grammar**: tree-sitter-go v0.23.1, ERROR rate 0 of 50 sampled files
- **Corpus**: Meridian API monorepo, 2,840 files, 96 MB of source

## 2. Query Inventory
- **Python defs.scm**: 14 captures covering function.def, class.def, import.ref, call.ref, decorator.ref
- **TypeScript defs.scm**: 17 captures covering function.def, class.def, interface.def, import.ref, call.ref, type.ref
- **Go defs.scm**: 12 captures covering function.def, method.def, type.def, import.ref, call.ref
- **Predicate use**: 9 `#match?` narrowings for route_, test_, and handle_ prefixes, 6 `#eq?` pins for framework entry points

## 3. Chunking Policy
- **Unit**: one chunk per top-level definition node with enclosing signature header
- **Budget**: 6000 bytes soft cap, split at nested definition boundaries above cap
- **Chunks emitted**: 18,204 chunks, 312 truncated with explicit markers, zero mid-expression cuts found in audit sample of 200 chunks
- **Index storage**: SQLite at .prumo/index/ts-chunks.sqlite3 keyed by file plus byte range

## 4. Incremental Performance
- **One-line edit on 2000-line reference file**: 9 ms with old-tree reuse versus 61 ms full parse
- **Throughput**: Python 4.8 MB per second, TypeScript 3.1 MB per second, Go 5.6 MB per second
- **Full corpus parse**: 96 MB in 26 seconds cold

## 5. Measured Results
- **Golden capture recall**: 320 of 320 golden nodes captured, 100%
- **Query p95 warm**: 142 ms over 400 sampled capture, import, and chunk queries
- **Ranking weights**: exact capture-name 1.0, same-file proximity 0.6, import-distance hops 0.3, recency 0.1, caps 40 chunks and 12000 tokens

## 6. Verification Evidence
- Flagged TypeScript file: web/client/legacy_bundle.d.ts with ERROR node at byte offset 44120, excluded from recall claims
- Recall log at .prumo/index/ts-recall-2026-09-14.log
- `scripts/verify.sh` exit code 0 on 2026-09-14
