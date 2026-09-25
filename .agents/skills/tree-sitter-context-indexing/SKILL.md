# Tree-Sitter Context Indexing

## Purpose
Build and query language-agnostic code-context indexes with tree-sitter grammars: pinned multi-language parsing, node and capture extraction via declarative queries, byte-range chunking for agent context windows, and incremental re-parsing that reuses unchanged subtrees for fast index refreshes.

## Use when
- An agent needs structural context such as function, class, import, or call nodes across many languages from one uniform engine.
- tree-sitter CLI 0.24 or newer with pinned grammar revisions is available for each target language.
- Files change frequently and re-parsing must cost proportional to the edit, not the file size.
- Retrieved chunks must align to syntactic boundaries instead of arbitrary line windows.

## Do not use when
- Semantic resolution such as overloads, trait dispatch, or type inference is required; tree-sitter parses syntax only and a compiler engine skill must be used instead.
- Only one language is indexed and its native compiler API already provides richer symbols at acceptable cost.
- Grammar quality for the target language is poor and would silently mislabel core constructs.

## Required context
- Target languages with pinned grammar revisions, for example tree-sitter-python at v0.23.4 and tree-sitter-typescript at v0.23.2.
- Checked-in query files in `.scm` format declaring the captures the index consumes.
- Chunking policy: maximum bytes per chunk, which node types may split, and overlap rules.
- Token budget for a single query answer: at most 40 chunks and 12000 tokens of emitted context.

## Procedure
1. **Pin grammars and verify parses**: lock each grammar revision in the spec, then run `tree-sitter parse --quiet` over a 50-file sample per language. Reject any grammar producing ERROR nodes on more than 1% of sampled files.
2. **Declare extraction queries**: write one `.scm` file per language capturing definitions, references, and imports, for example `(function_definition name: (identifier) @function.def)` with `#match?` and `#eq?` predicates narrowing overloads. Store queries under version control beside the spec.
3. **Parse incrementally**: keep the previous syntax tree per file and call the incremental entry point with the old tree plus the byte-range edit, so unchanged subtrees are reused. Measure that a one-line edit re-parses in under 15 ms on a 2000-line file.
4. **Chunk by syntax**: emit one chunk per top-level definition node, splitting oversized bodies at nested definition boundaries and carrying the enclosing signature line as header. Never emit a chunk that starts or ends mid-expression.
5. **Serve ranked queries**: expose `nodes --capture function.def --name 'route_*'`, `imports --file <path>`, and `chunks --symbol <name>` ranked by exact capture-name match, then same-file proximity, then import-distance hops, then recency, truncating at the 40-chunk and 12000-token caps.
6. **Verify against goldens**: run `scripts/verify.sh`, assert parse throughput above 2 MB per second per language and 100% capture recall on the golden node set before declaring the index usable.

## Decision rules
- **Pinned grammars only**: floating grammar revisions are forbidden; a grammar bump is a spec change with re-verified recall.
- **Queries are code**: `.scm` files are reviewed, versioned artifacts; inline ad-hoc query strings in index code are forbidden.
- **ERROR nodes are data**: files with parse errors are indexed with an error flag and excluded from recall claims, never silently counted as clean.
- **Syntax bounds respected**: no chunk may cross an expression boundary; oversized nodes split at nested definitions or are truncated with a marker.
- **Throughput floor enforced**: any language parsing below 2 MB per second is profiled and fixed or dropped from the index scope.

## Evidence required
- Index specification following `templates/tree-sitter-context-indexing-spec.md` with pinned grammars and query inventory.
- Per-language ERROR-node rates on the 50-file samples and parse throughput measurements.
- Golden capture-recall log with measured query p95 latency.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Pinned multi-language parser set with verified `.scm` extraction queries.
- Syntax-aligned chunk index queryable by capture, name, imports, and symbol.
- Parse quality and recall report proving the index meets its targets.

## Stop conditions
- Capture recall reaches 100% on goldens with query p95 under 200 ms on a warm index.
- Incremental re-parse of a one-line edit completes in under 15 ms on reference files.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the language owner if a required grammar misparses core constructs above the 1% ERROR threshold with no upstream fix available.
- Escalate to the lead architect if chunk-window requirements conflict with syntactic boundaries, so truncation policy can be decided explicitly.
