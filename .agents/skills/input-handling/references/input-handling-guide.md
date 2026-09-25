# Input Handling Reference Guide

## Action Abstraction

Gameplay consumes named actions, not device codes. A binding table maps keyboard, mouse, gamepad, and touch controls to actions such as `jump`, `confirm`, and `aim`. Device adapters translate hardware state into timestamped action events.

## Buffered Input

Discrete presses may be consumed within a declared forgiveness window. Queues have a maximum size, expiration rule, and drop counter. Overflow policy is explicit; silent loss makes gameplay bugs unreproducible.

## Analog Calibration

Use radial dead zones, a documented response curve, saturation handling, and per-device overrides. Triggers need separate press and release thresholds to prevent flutter. Calibration tables define expected output for representative hardware values.

## Chords and Gestures

Chords require all participants within a simultaneity window. Sequences use a trie or equivalent state machine with per-step timeout and reset rules. Reject conflicting bindings when they are created, not when a player later presses them.

## Persistence and Churn

Store user rebinds as versioned data separate from code defaults. Resolve new-default versus user-binding conflicts deterministically and record the override. Handle hot-plug without using stale device state or crashing mid-frame.

## Patterns and Anti-Patterns

| Do | Avoid |
|---|---|
| Map actions through a central table | Branch on raw keycodes in gameplay |
| Record drops and overflow | Grow input queues without bound |
| Calibrate radial sticks correctly | Scale both axes independently from a boxed square |
| Reject conflicts at bind time | Resolve conflicts by whichever event arrives last |
| Replay timestamped events | Reconstruct input from frame polling only |

## Short Example

`jump` maps to Space, gamepad south button, and the touch jump zone. A press 80 ms before landing remains valid under a 120 ms buffer; the tenth queued press is not possible because the ring holds eight events.
