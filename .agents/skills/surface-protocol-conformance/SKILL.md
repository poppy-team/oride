---
name: surface-protocol-conformance
description: Guarantees semantic uniformity across CLI, TUI, GUI, and daemon over the public Prumo Protocol.
---
# Surface & Protocol Conformance

## 1. Single Domain Source
A domain rule must exist once in application/domain services. Clients and surfaces must never duplicate domain rules locally or invent private behavior.

## 2. Non-Bypassable Governance
Ensure no surface bypasses the Permission Engine, Budget Manager, or Gauntlet Quality Gates.

## 3. Protocol Version Negotiation & Replay
Verify that clients negotiate protocol versions and can reconnect/replay events without state corruption.
