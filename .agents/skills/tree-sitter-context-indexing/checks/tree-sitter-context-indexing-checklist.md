# Tree-Sitter Context Indexing — Verification Checklist

## 1. Grammar Pinning & Parse Quality
- [ ] Every target language lists its pinned grammar revision in the spec; floating revisions are absent.
- [ ] `tree-sitter parse --quiet` over 50 sample files per language shows ERROR nodes on at most 1% of files.
- [ ] Files with parse errors carry an error flag and are excluded from recall claims.

## 2. Extraction Queries as Versioned Code
- [ ] One `.scm` file per language lives under version control beside the spec.
- [ ] Queries capture definitions, references, and imports using `#eq?` and `#match?` predicates where narrowing is needed.
- [ ] No inline ad-hoc query strings exist in the indexing code.

## 3. Incremental Re-Parsing
- [ ] Re-parses reuse the previous tree with byte-range edits; full re-parses on small edits are absent.
- [ ] A one-line edit on a 2000-line reference file re-parses in under 15 ms.
- [ ] Parse throughput exceeds 2 MB per second per indexed language.

## 4. Syntax-Aligned Chunking
- [ ] Each chunk maps to whole definition nodes with the enclosing signature as header.
- [ ] No chunk starts or ends mid-expression; oversized nodes split at nested definitions or carry a truncation marker.
- [ ] Chunk inventory records byte ranges, capture kinds, and parent symbols.

## 5. Ranking & Verification Evidence
- [ ] Ranking applies exact capture-name, same-file, import-distance, and recency order.
- [ ] All query paths enforce the 40-chunk and 12000-token caps.
- [ ] Golden capture recall is 100% with query p95 under 200 ms warm.
- [ ] `scripts/verify.sh` executes with exit code 0.
