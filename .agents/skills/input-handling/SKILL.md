# Input Subsystem

## Purpose
Implement responsive, remappable multi-device input: action-based mapping (not raw key polling), input buffering with timing windows, dead-zone and sensitivity calibration, chord/gesture detection, gamepad hot-plug, and latency within an explicit input-to-photon budget.

## Use when
- Adding action mapping, rebinding UI, or multi-device support (keyboard, mouse, gamepad, touch).
- Implementing input buffering, coyote-time/jump-forgiveness windows, or combo/chord detection.
- Calibrating sticks (dead zones, response curves), triggers, or pointer sensitivity.
- Auditing dropped inputs, double-fires, focus-stealing, or input latency.

## Do not use when
- Designing the simulation that consumes input (use `game-runtime`).
- Building editor gizmo manipulation UX (use `editor-tooling`).
- Securing networked input against cheating (server-authoritative validation concern, pair with `network-security` for transport).

## Required context
- Device matrix: keyboard, mouse, gamepad, touch with platform APIs.
- Action inventory: gameplay actions, chords, and rebinding requirements.
- Latency budget: input-to-photon ceiling and buffering window (ms).

## Procedure
1. **Map Actions, Not Keys**: Define gameplay actions (`jump`, `dodge`, `confirm`) with per-device binding tables. Gameplay code reads actions; devices produce them. Example: `jump ← Space | GamepadA | TouchButton(jump-zone)`. Raw keycodes never appear in gameplay logic.
2. **Buffer with Windows**: Queue discrete action presses with timestamps; consume within an explicit window (e.g., 120 ms buffering, 100 ms coyote time). Buffer size is bounded (e.g., 8 events); overflow drops oldest with a debug counter, never newest.
3. **Calibrate Sticks & Triggers**: Apply radial dead zones (e.g., 0.15 inner, 1.0 outer) plus a response curve (e.g., cubic ease) per stick; document per-device overrides. Triggers use press/release hysteresis (press > 0.4, release < 0.3) to prevent flutter.
4. **Detect Chords & Gestures Deterministically**: Chords require all members within a simultaneity window (e.g., 50 ms); sequences match against a trie with per-step timeouts. Conflicting bindings (same gesture → two actions) are rejected at bind time with a named error.
5. **Survive Device Churn**: Handle gamepad hot-plug/unplug mid-frame: pause-on-disconnect for single-player, migrate bindings to the replacement device, and persist user rebinds to versioned storage with conflict resolution on load (user binding wins; log the override).
6. **Verify**: Run `scripts/verify.sh`. Prove buffered-input execution rates at 30/60/120 fps simulation, dead-zone curves against calibration tables, chord timing boundaries, and zero dropped actions in a scripted input replay.

## Decision rules
- **Actions Only in Gameplay**: Gameplay polling raw device state is a defect; all input arrives as named actions.
- **Bounded Buffers**: Unbounded input queues are forbidden; every buffer has size, window, and drop policy.
- **No Silent Conflicts**: Overlapping bindings fail loudly at bind time, never at press time.
- **Rebinds Persist**: Default bindings are code; user bindings are versioned data that survives updates.

## Evidence required
- Input-replay test: scripted sequence yields identical action streams across frame rates.
- Calibration tables: measured vs expected stick curves and trigger thresholds.
- Chord/gesture boundary tests and conflict-rejection cases.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Action-mapped input subsystem with buffering and dead-zone calibration.
- Rebinding persistence with conflict-resolution rules.
- Latency and dropped-input benchmark evidence.

## Stop conditions
- All actions mapped, buffered, and rebindable within latency budget.
- Zero dropped or double-fired actions in scripted replay.
- Token budget exhausted.

## Escalation rules
- Escalate to lead architect if platform input latency floor exceeds the design budget (needs game-feel tradeoff).
- Escalate immediately on device-churn crashes or input handling on the wrong thread causing races.
