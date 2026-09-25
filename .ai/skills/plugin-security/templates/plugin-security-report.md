# Plugin & Extension Security Audit Report

## Scope
- Component: `extension-broker` and `csv-importer@1.4.0`
- Assessor: security-review-agent
- Review date: 2026-09-23
- Evidence: signed manifest, broker audit log, and adversarial test run

## Capability Matrix
| Capability | Declared | Exercised | Verdict |
|---|---|---|---|
| `fs.read:workspace/uploads/**` | yes | 212 reads | allow |
| `net.fetch:https://telemetry.example.net` | no | 3 attempts | deny |
| `clipboard.read` | yes | 0 uses | remove declaration |

## Findings
- **HIGH**: Three undeclared telemetry requests reached the broker; the network policy denied each request.
- **MEDIUM**: An unused `clipboard.read` capability widened the manifest without a product need.
- **LOW**: The plugin process restarted once after an injected fault; the supervisor stayed within its three-restart budget.

## Remediation
- Removed `clipboard.read` and pinned the bundle to the approved content hash.
- Kept the network capability absent and added a regression test for metadata-IP fetches.
- Recorded broker decisions in the append-only audit log.

## Sign-Off
- Manifest: approved
- Sandbox: approved
- Next review: 2026-10-23
