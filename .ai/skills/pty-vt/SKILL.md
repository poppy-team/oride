---
name: pty-vt
description: Embedded PTY sessions and VT emulation — lifecycle, resize, scrollback bounds, teardown on panic, and why shell invocation is an injection boundary.
---

# PTY & VT Emulation

## Purpose

The embedded terminal is a process boundary, not a widget. It owns a real child
process, a file descriptor, and a screen buffer that must survive output the
editor did not anticipate.

## Session lifecycle

1. Spawn once, per panel, on demand. Never on startup — an editor that always
   starts a shell pays for it on every launch and breaks in environments without
   one.
2. The child is reaped. A PTY whose child was never waited on leaks a process
   per session.
3. Teardown is unconditional: on quit, on panic, on signal. A terminal left in
   raw mode or alt-screen after a crash is the failure users remember.
4. Closing the panel does not silently kill a foreground job without saying so.

## Resize

- The window size is propagated to the child, not just to the drawing box. A
  program that keeps rendering 80 columns into a 40-column PTY corrupts its own
  output.
- Resize is coalesced: a drag produces many intermediate sizes, and each one
  costs a SIGWINCH in the child.
- The initial size is the real panel size, not a guess of 80×24.

## Emulation

- Parse into a cell buffer with attributes; never strip escape sequences and hope.
- Alt-screen has its own buffer: the scrollback is not visible while it is
  active, and the primary buffer must be intact when it is left.
- Scrollback is bounded. An unbounded scrollback is an out-of-memory bug with a
  long fuse.
- Grapheme width is resolved at write time, so a wide character occupies the
  cells it claims and alignment downstream is not corrupted.
- Unknown sequences are consumed without corrupting the buffer — real programs
  emit sequences a partial implementation has never seen.

## Injection boundary

The shell command is user configuration. It is executed as a program with
arguments, never interpolated into a shell string, and never built from file
contents or filenames. Working directory is passed explicitly rather than by
`cd`.

## Verification

- A PTY test with a real shell, gated by an environment variable so it can be
  skipped where a shell is unavailable.
- Feed a recorded byte stream through the emulator and compare the resulting
  cell grid — deterministic, no timing.
- A test asserts the terminal is restored after a panic in the draw path.
