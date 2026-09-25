---
name: tui-editor-engine
description: Text buffer and editing semantics for a terminal code editor — index conversion, transaction-first editing, multi-cursor invariants, undo grouping, and large-file behaviour.
---

# TUI Editor Engine

## Purpose

Keep the editor core correct and testable without a terminal. Everything here
lives behind `internal/buffer` and `internal/editor`; nothing in this skill
requires a TTY.

## Index semantics — the rule that prevents silent corruption

A text buffer has at least four notions of "position", and confusing them is the
single most expensive class of bug in an editor.

- **Byte offset** is the canonical internal index.
- **Rune (scalar) offset** is what a "column" means to a caret.
- **Grapheme cluster** is what a user perceives as one character.
- **Visual column** accounts for tab width and wide characters.

Every conversion between them lives in `internal/buffer`. No caller converts on
its own, and no caller reaches into another package's arithmetic. The Rust
implementation got this from `ropey` for free; Go does not, so it is explicit
here and covered by property tests.

Test cases must include: combining marks, emoji with modifiers, CJK wide
characters, RTL text, a BOM, CRLF line endings, and a lone surrogate in a
malformed file.

## Transaction-first editing

Agent patches, user keystrokes, LSP code actions and undo/redo converge on the
same operation: a transaction with a recorded inverse. This is what makes a
diff, a conflict check and a replay possible without special cases.

- An edit is applied and its inverse is recorded in the same step.
- A group of edits is one undo step and carries the selection it produced.
- Nothing mutates the buffer outside a transaction.

## Undo grouping

A group closes on an explicit boundary, never on a heuristic timer.

| Boundary | Why |
|---|---|
| Newline | Every editor breaks the group on Enter; users expect it |
| Cursor movement | Typing, moving, typing again are two thoughts |
| Save | The saved state is a checkpoint |
| Focus change | A different context is a different edit |
| Explicit call | Paste and programmatic edits are one step |

The Rust implementation closed the group on movement but not on newline, so
`Ctrl+Z` after `abc` + Enter + `xyz` erased all three. That is ledger item B1.

## Undo and redo restore position

Undo and redo restore the selection the group produced — not a clamped version
of wherever the cursor happened to be. Redo of an insertion puts the cursor after
the inserted text. Ledger items B2–B4 cover the Rust behaviour here, including
extra carets left pointing past the end of a shrunken buffer.

## Multi-cursor invariants

- Carets are kept sorted and deduplicated.
- Edits apply from the highest offset down, so a prior edit never invalidates a
  later offset.
- Any operation that shrinks the buffer re-validates every caret.
- A selection and extra carets are mutually exclusive as an edit target: either
  the selection is replaced, or every caret inserts.

## Large files

Decide by capability, not by hope:

- Binary detection before decoding — refuse rather than corrupt.
- Line-indexed access so the viewport never walks the whole file.
- Syntax, diagnostics and search bounded to what is visible or requested.
- Search incremental, never a full-document regex scan on a 100 MB file.
- The statusline says when a mode is degraded, instead of silently dropping
  features.

## Verification

- Property tests over random edit sequences: apply, undo everything, and the
  buffer equals the original byte for byte.
- Conformance cases compare the buffer, carets, selection and undo labels
  against the Rust oracle.
