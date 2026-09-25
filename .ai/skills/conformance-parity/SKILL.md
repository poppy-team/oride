---
name: conformance-parity
description: Differential parity against a frozen reference implementation — fixture format, normalisation, determinism, and the rule that a divergence without a ledger entry is a regression.
---

# Conformance & Parity

## Purpose

Turn "the port behaves the same" from a claim into a number a program produces.
This project migrates Oride from Rust to Go while a frozen Rust binary remains
the behavioural reference.

## The two guarantees everything rests on

**The oracle is deterministic.** The same case run twice, in two different
directories, must produce identical reports. Differential parity is meaningless
otherwise. This is asserted by a test, not assumed — and that test is what caught
the workspace basename leaking into the state dump.

**The comparator detects divergence.** A comparator that always answers "equal"
would let parity climb to 100% on its own. That is the only failure mode of this
harness that would look like success, so the comparator is tested with a
divergence it must find.

## What parity means, and what it cannot mean

Byte-for-byte parity of rendered frames between ratatui and Lip Gloss is
unattainable. Claiming it would fail silently. Parity has two tiers:

**Tier A — behavioural, byte-exact.** Observable state serialised as canonical
JSON plus any file the run produced: buffer text, carets, selection, undo labels,
find results with offsets, git status, tree order, session on disk, effective
config. Compared without tolerance.

**Tier B — presentation, contractual.** Not escape sequences. The documented
interaction contract: the keymap in the docs matches the bindings in the code,
the focus graph and state vocabulary match `docs/ui-ux/`, widths and capability
degradation behave as documented. Cross-implementation frame equivalence is
checked on a semantic grid (cell to rune plus semantic token name).

## Normalisation

A fixture must not depend on where it ran:

- Paths are workspace-relative, with the workspace replaced by a placeholder.
- Volatile values are redacted or reduced to a count.
- Anything derived from the environment — installed tools, terminal capabilities,
  locale — appears as a boolean or a count, never as text.
- Numbers stay exact through decoding; a byte count must not round-trip through
  floating point.

## The ledger rule

Every deliberate divergence from the oracle lives in
`docs/migration/parity-ledger.md`, with a verdict: *spec wins*, *code wins*,
*removed*, or *improvement*. A divergence that is not in the ledger is a
regression, not a decision.

Where the verdict is *spec wins* and the oracle fix is cheap, **fix the oracle
too** — every exception costs conformance coverage.

## Failure must not look like success

- A missing oracle skips with instructions. A broken oracle fails.
- A suite where nothing was comparable reports that, instead of 0% or 100%.
- A harness fault — schema mismatch, step-count mismatch — returns an error, not
  a parity failure.

## Verification

```bash
cargo build -p oride            # `cargo test` does not refresh the binary
go test ./conformance/...
```
