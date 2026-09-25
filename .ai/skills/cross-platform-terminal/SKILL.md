---
name: cross-platform-terminal
description: Behaviour of a terminal application across Linux, macOS and Windows — ConPTY, shell discovery, PATH lookup, path separators, and capability detection without sniffing.
---

# Cross-Platform Terminal

## Purpose

An editor that works only on Linux is not the product. The differences that bite
are small, unglamorous, and invisible until someone on Windows runs it.

## Process and shell

- Unix PTY and Windows ConPTY are different mechanisms behind one interface.
- The shell is discovered, not assumed: `$SHELL`, then a search order per
  platform. `/bin/sh` is not a fallback on Windows.
- Argument separation is structural. Never build a command string that a shell
  parses — that is an injection boundary and a quoting bug in one.
- A program is executable if the platform says so: the execute bit on Unix, the
  extension set on Windows. One check does not cover both.

## Paths

- Separators are normalised where they cross a boundary and nowhere else. A path
  used as a key keeps its platform form, or two spellings of one file become two
  entries.
- Comparisons that matter (session identity, document identity) use the
  canonical form, and canonicalisation must survive symlinks — on macOS `/tmp`
  and `/var` are symlinks, so a path can differ before and after.
- Windows has path lengths, reserved names and UNC prefixes. Assume only what the
  standard library guarantees.
- A path from the user may contain any byte. It is never interpolated into a
  command, and never assumed to be valid UTF-8.

## Capability detection

Detect by asking the terminal, not by guessing from environment variables.
`TERM`, `TERM_PROGRAM`, `KITTY_WINDOW_ID` and friends are a heuristic that goes
stale; the terminal protocols report capabilities directly. Where a heuristic is
unavoidable, it is one function, tested, and it fails toward the conservative
answer.

## Degradation

Every capability has a fallback that keeps the product usable:

| Capability | Fallback |
|---|---|
| True colour | 256, then 16, then none |
| Unicode box drawing | ASCII `+ - \|` |
| Mouse | Keyboard equivalents that remain complete |
| Graphics protocol | A text placeholder with the image's dimensions |
| Clipboard | The editor's internal register |

A missing capability changes how something looks, never whether it can be done.

## Verification

- Tests must not write to the process environment to simulate a terminal. Inject
  the capability. A test that mutates global state is a flaky test under
  parallelism — this repository already has one, ledger item B16.
- Path and argument construction is pure and tested without a filesystem.
