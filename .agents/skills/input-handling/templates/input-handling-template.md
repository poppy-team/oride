# Input Subsystem Report — Rebindable Action Map

## 1. Metadata
- **Skill**: input-handling (Input Subsystem)
- **Date**: 2026-09-07
- **Author / Agent**: gameplay-team
- **Target Goal / Phase**: GOAL-117 action-map-rebind

## 2. Executive Summary
Shipped the action-mapped input subsystem: 24 gameplay actions across keyboard/mouse/gamepad/touch with a rebinding UI, 120 ms input buffering, calibrated stick curves (0.15 dead zone, cubic response), and chord detection (50 ms window). Scripted replay yields identical action streams at 30/60/120 fps; input-to-photon measured at 58 ms against the 80 ms budget.

## 3. Inputs & Scope
- **Inputs Evaluated**: 24-action inventory (`actions.yaml`), device matrix (KB+M, XInput, DualShock, touch zones), 80 ms latency budget
- **Artifacts Modified**: `input/action_map.cpp`, `input/buffer.cpp`, `input/rebind_store.cpp`, `ui/rebind_screen.*`

## 4. Key Findings & Implementation Details
- **Action map**: `jump ← Space | GamepadA | Touch(jump-zone)`; gameplay reads actions only (grep audit: 0 raw keycodes in `gameplay/`).
- **Buffering**: 120 ms window, 8-event ring; overflow drops oldest with `input.dropped_total` counter (0 drops in 30-min soak).
- **Calibration**: Radial dead zone 0.15/1.0 + cubic curve; per-device override for DualShock (0.18 inner); trigger hysteresis press 0.4 / release 0.3 — flutter test 0/1,000 false releases.
- **Chords**: 50 ms simultaneity window; sequences via trie with 400 ms per-step timeout; bind-time conflict rejection verified (12 conflicting pairs rejected with named errors).
- **Device churn**: Unplug mid-match pauses single-player and migrates bindings; rebinds persist in versioned `bindings_v2.json` (user wins on conflict, override logged).
- **Latency**: Input-to-photon 58 ms p95 on target hardware (budget 80 ms).

## 5. Verification & Evidence
- **Evidence Type**: test, benchmark
- **Test Results**: Passed — replay identical 3/3 frame rates; calibration tables 14/14 in tolerance; conflict rejection 12/12; soak 0 drops
- **Static Analysis Status**: Pass — input lint clean; no gameplay raw-device reads

## 6. Next Steps & Handoff
- Gyro-aim bindings for handheld tier (GOAL-123); owner: gameplay-team.
- Accessibility preset review (one-hand layout); owner: a11y-champion.
