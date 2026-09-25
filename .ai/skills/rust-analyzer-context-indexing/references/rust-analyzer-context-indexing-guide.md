# rust-analyzer Context Indexing — Technical Reference Guide

## 1. Core Concepts

### 1.1 Crate Graph from Cargo Metadata
rust-analyzer does not guess the workspace layout; it consumes `cargo metadata`. Run it first and keep the output:

```sh
cargo metadata --format-version 1 --no-deps > .prumo/index/cargo-metadata.json
```

If this command fails, the crate graph is unknown and indexing must stop. An empty result set is a failure signal, never a valid index.

### 1.2 LSP Methods Used for Indexing
Four methods carry the whole index: `workspace/symbol` for bulk discovery, `textDocument/documentSymbol` for per-file hierarchy, `textDocument/definition` for jump targets, and `textDocument/references` with `includeDeclaration: true` for call sites. A bulk request looks like:

```json
{
  "jsonrpc": "2.0",
  "id": 42,
  "method": "workspace/symbol",
  "params": { "query": "apply_entry", "limit": 50 }
}
```

Cap bulk passes at 500 results per crate and page with narrower queries; unbounded symbol dumps blow the token budget before ranking even runs.

### 1.3 Salsa Revisions and Flycheck
rust-analyzer recomputes incrementally through its salsa query engine: each saved edit bumps a revision and only affected queries rerun. Diagnostics arrive from flycheck running `cargo check --workspace --message-format=json`. The indexer treats error diagnostics as a freshness gate — symbols under errors are servable only with a stale flag.

### 1.4 Why Trait Edges Matter Most
In idiomatic Rust, the question is rarely where a function is defined but which `impl` answers a trait call. Storing `trait LedgerStore -> impl SqliteStore, impl MemoryStore` turns a vague name query into an exact dispatch map, which is why traits without implementation edges are treated as incomplete records.

## 2. Startup Configuration Notes
- Set `cargo.targetDir` to the workspace target directory so flycheck artifacts never pollute checkouts.
- Enable `procMacro.enable` only when macro-generated items are in scope; it costs load time on large workspaces.
- Point at the rustc sysroot sources so `std` symbols like `Option` and `Result` resolve instead of dangling.

## 3. Common Pitfalls
- Indexing with a stale `Cargo.lock` while features unify differently under the active toolchain.
- Dropping `includeDeclaration` on references and losing the defining site from call graphs.
- Treating `workspace/symbol` order as ranking; the server returns matches unordered, so the skill ranking model must score them.
