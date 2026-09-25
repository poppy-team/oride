# Input Handling Verification Checklist

## Action Mapping
- [ ] Gameplay consumes named actions and never polls raw device codes.
- [ ] Keyboard, mouse, gamepad, and touch bindings have an explicit action inventory.
- [ ] Default and user bindings have separate ownership and precedence rules.

## Buffered and Discrete Input
- [ ] Every action buffer has a size, time window, expiration, and overflow policy.
- [ ] Buffer consumption is deterministic at supported simulation rates.
- [ ] Dropped, duplicated, or reordered actions expose counters and replay fixtures.

## Analog and Gesture Calibration
- [ ] Stick dead zones, response curves, saturation, and per-device overrides are documented.
- [ ] Trigger press and release thresholds prevent flutter.
- [ ] Chord and sequence timing boundaries have exact tests.

## Persistence and Device Churn
- [ ] User rebinds use a versioned format with deterministic migration and conflict resolution.
- [ ] Conflicting bindings are rejected at creation time with a named error.
- [ ] Hot-plug, unplug, focus changes, and device loss do not leak or reuse stale state.

## Latency and Evidence
- [ ] Input-to-action and input-to-photon budgets are measured on target hardware.
- [ ] Accessibility presets remain remappable and do not depend on color or motion alone.
- [ ] `scripts/verify.sh` exits with code 0 from the repository root.
