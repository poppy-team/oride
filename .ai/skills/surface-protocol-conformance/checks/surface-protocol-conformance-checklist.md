# Surface & Protocol Conformance — Verification Checklist

## 1. Single Domain Source
- [ ] Each operation resolves to exactly one domain-service entry point; no duplicated rules in surfaces.
- [ ] Client-side pre-validation mirrors domain errors only; authoritative checks stay server-side.
- [ ] New operations added with domain entry point cited in the conformance matrix.

## 2. Non-Bypassable Governance
- [ ] CLI, TUI, GUI, and daemon all route mutations through the Permission Engine verdict path.
- [ ] Budget Manager enforced on every metered operation from every surface.
- [ ] Gauntlet Quality Gates block non-conforming changes on all surfaces equally.

## 3. Protocol Negotiation & Replay
- [ ] Client/daemon handshake negotiates a versioned protocol within the supported skew window.
- [ ] Reconnect replays missed events idempotently; sequence numbers monotonic, no state corruption.
- [ ] Behavior outside the supported window fails safe with a clear version-mismatch error.

## 4. Semantic Parity Evidence
- [ ] Conformance matrix complete: every surface × operation cell cites the shared entry point and gate verdict.
- [ ] Golden-trace replay across all surfaces yields identical domain-state hashes.
- [ ] Error-taxonomy fixture passes: identical domain errors surface identical codes everywhere.

## 5. Sign-Off
- [ ] Drift findings (bypasses, duplicated logic, divergent errors) filed with severity and owner.
- [ ] `scripts/verify.sh` exits 0 from the repository root.
