---
name: bubbletea-v2
description: Charm v2 discipline for Bubble Tea programs — declarative views, Cmd/Msg as the only async boundary, headless golden-frame testing, and terminal capability degradation.
---

# Bubble Tea v2 Discipline

## Purpose

Keep the Oride TUI correct under the Charm v2 (Elm-style) runtime: no I/O in
`View`, no blocking in `Update`, and every terminal capability either declared
or gracefully absent.

## Rules

1. `View()` is pure and cheap. No file reads, no process spawns, no network.
   It runs on every frame; anything slow belongs in a `Cmd`.
2. `Cmd`/`Msg` is the only asynchronous boundary. Goroutines reach the program
   through `Program.Send`, never by mutating shared state.
3. Never block `Update`. Anything that can take time returns a `Cmd`.
4. Declare terminal features in the view: alt screen, mouse mode, keyboard
   enhancements, focus reporting and cursor. Do not toggle them imperatively per
   frame — that is how a terminal is left broken after a crash.
5. Request keyboard enhancements only when the program uses them, and always
   handle the terminal that does not support them. The `ctrl+shift+<letter>`
   ambiguity is real and must be resolved from the key's own fields, not from a
   guess about the terminal.
6. Every colour decision degrades: true colour, then 256, then 16, then none.
   A `no-color` theme must resolve every non-primitive token to nothing, and a
   palette that claims to suppress colour while painting one is an error.
7. No meaning carried by colour alone. A state that is tinted is also spelled
   out in words.
8. Width is conservative: assume ambiguous-width characters are wide, and never
   let a row silently lose its tail — truncate with a visible marker.
9. Cursor shape and blink are view state, not terminal escape sequences written
   by hand.
10. Batching has two forms and they are not interchangeable: `Batch` runs
    concurrently, `Sequence` runs in order. Choose deliberately.

## Anti-patterns

- A `Cmd` that only sends a message to another part of the same program — that
  is an `Update` call.
- Doing work in `Init` that can fail; `Init` has no error path.
- Reading `os.Getenv` for the terminal when the program already receives the
  environment as a message.
- Rendering the whole frame on each keystroke. Announcements are deltas.

## Verification

- Headless tests drive a real `tea.Program` and compare frame buffers byte for
  byte; a frame and the code that draws it change in the same commit.
- A test asserts that with reduced motion the frames differ in the animated line
  and in no other.
- Capability is injected by option, never by writing to the process environment.
