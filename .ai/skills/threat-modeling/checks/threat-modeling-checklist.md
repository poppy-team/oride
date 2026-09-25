# Threat Modeling — Verification Checklist

## 1. Scope & Decomposition
- [ ] System decomposed into processes, data stores, external entities, and data flows (DFD recorded).
- [ ] Trust boundaries drawn everywhere data crosses privilege levels; each crossing numbered.
- [ ] Sensitive-data lifecycle traced from ingest to storage to egress.

## 2. STRIDE Coverage
- [ ] Each DFD element analyzed against applicable STRIDE categories (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege).
- [ ] Abuse cases written alongside use cases for attacker-reachable features.
- [ ] Attack surface enumerated (endpoints, jobs, uploads, integrations) with minimization notes.

## 3. Mitigations & Ownership
- [ ] Every identified threat has a mitigation (auth, MACs/TLS, audit logging, encryption, rate limits, least privilege) and an owner.
- [ ] Mitigations favor layered controls over single-point defenses for high-impact threats.
- [ ] Residual accepted risks listed explicitly with approver and expiry.

## 4. Validation & Freshness
- [ ] Model reviewed against current architecture (no stale components or phantom flows).
- [ ] Findings cross-checked with SAST/DAST or review evidence where applicable.
- [ ] `scripts/verify.sh` exits 0 from the repository root.

## 5. Handoff
- [ ] Threat-model report filed with DFD diagram, threat table, and mitigation status.
- [ ] Re-review trigger defined (architecture change, new integration, incident).
