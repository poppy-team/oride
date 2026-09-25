---
name: doc-drift
description: Deterministic detection of divergence between what the documentation claims and what the code does — enforced by tests that fail the build, not by review.
---

# Documentation Drift

## Purpose

Documentation that nobody verifies diverges. This project found eleven such
divergences in one pass, including a config key that did not exist, a command
that was never implemented, and a default documented the opposite of the code.
Reviews do not catch these; tests do.

## Rule

**A claim that can be checked mechanically must be checked mechanically.** If a
document states a key name, a command, a default value, or a keybinding, there
is a test that fails when it stops being true.

## What gets a guard

| Claim | Guard |
|---|---|
| Keybindings listed in the docs | Parsed from the docs and compared against the resolved keymap |
| Defaults documented in `assets/config.example.toml` | Parsed and compared against the product defaults |
| Action ids referenced in examples | Every id must resolve through `parse_action` |
| Config keys in the guides | Must exist in the schema — a key nobody reads is a lie |
| `:` commands in the guides | Must be in the implemented command set |
| Roadmap and changelog feature claims | Must name something that exists in the tree |

## Procedure

1. State the claim in a form a program can read.
2. Write the comparison as a test.
3. Run it — a guard that has never failed has not been shown to work.
   Reintroduce the divergence once and confirm the test fails.
4. Only then fix the document.

Step 3 is not optional. A guard nobody has seen fail is indistinguishable from a
guard that does nothing.

## When code and document disagree

Decide explicitly and write the verdict in the parity ledger:

- **code wins** — the document is wrong; fix the document and add the guard.
- **spec wins** — the code is wrong; fix the code, and note whether the frozen
  oracle must change too.

Never leave the disagreement unrecorded. An unresolved contradiction teaches
every later reader something false.

## Smells

- A guide that names a key, command or flag no test mentions.
- The same fact stated in two documents — one will drift.
- A README section describing a feature the changelog does not list.
- A translated document that lags its source; translations are projections and
  must carry a freshness signal.
