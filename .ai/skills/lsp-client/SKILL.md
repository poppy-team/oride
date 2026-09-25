---
name: lsp-client
description: Language server client over stdio JSON-RPC — framing, lifecycle, UTF-16 column conversion, on-demand spawn, and fail-closed behaviour.
---

# LSP Client

## Purpose

Obtain language intelligence from a server the editor does not own. The server
is a third-party process that may be missing, slow, crash, or speak a version the
client did not expect. None of those may take the editor down.

## Framing

Content-Length framing over stdio, with the header parser tolerant of the order
and case real servers emit. A malformed frame is a protocol error to report, not
a reason to panic or to resynchronise blindly into the middle of a message.

## Lifecycle

- One client per language per workspace. Two servers for the same language is a
  bug that shows up as duplicated diagnostics.
- Spawned on demand, when a document of that language opens — not at startup.
- `initialize` completes before any request that needs capabilities; the client
  tracks the negotiated capability set and does not send requests the server
  declared unsupported.
- Shutdown is graceful, with a bounded wait, then terminated. A server left
  running after the editor exits holds locks and ports.
- Crashes are contained: the client reports unavailability for that language and
  the editor keeps working without intelligence for it.

## Position conversion — the classic silent corruption

LSP counts UTF-16 code units. A caret counts runes, and the buffer counts bytes.
Three units, three conversions, and getting one wrong corrupts text at the first
emoji or astral character.

- Conversion lives in one place and is covered by tests with surrogate pairs,
  CJK, combining marks and RTL.
- Positions from the server are clamped to the document; a server may report a
  position past the end and the editor must not panic.
- `TextEdit` application is ordered by position and applied from the end
  backwards, so earlier edits do not invalidate later offsets.

## Fail closed

Missing binary, failed handshake, timeout, crash: each produces a status the user
can read and an editor that still edits. Never a crash, never a silent hang,
never a feature that appears to work and does nothing.

## Sharing

Language intelligence is expensive. Before spawning, check whether an equivalent
server is already running for this workspace; do not duplicate a server the
harness already owns.

## Verification

- A stub server drives the handshake, a diagnostic push, a completion and an
  edit round-trip, without depending on any real server being installed.
- A test asserts the client survives a server that closes mid-message.
