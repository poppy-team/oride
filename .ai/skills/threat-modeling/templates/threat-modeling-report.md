# Threat Modeling Audit Report

## Scope
- Component: customer login and session service
- Assessor: threat-modeling-agent
- Review date: 2026-09-23
- Model revision: LOGIN-2026-09-2

## Trust Boundaries
- TB-1: Internet client to edge gateway
- TB-2: gateway to authentication service
- TB-3: authentication service to session store
- TB-4: service to audit log sink

## Prioritized Threats
- **HIGH / Spoofing**: replayed refresh token; mitigated by rotation, audience binding, and reuse detection.
- **HIGH / Elevation of Privilege**: client-supplied role claim; mitigated by server-side policy lookup.
- **MEDIUM / Information Disclosure**: verbose auth error; mitigated by generic responses and redacted logs.
- **MEDIUM / Denial of Service**: password-spray traffic; mitigated by per-account and per-IP limits.

## Decisions
- Residual risk: mobile devices may lose a session during network switching; accepted until 2026-10-01.
- Re-review trigger: new identity provider, new session store, or any authentication incident.
