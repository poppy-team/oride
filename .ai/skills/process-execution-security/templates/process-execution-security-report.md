# Process Execution Security Audit Report

## Scope
- Component: `repository-runner` process launcher
- Assessor: platform-security-agent
- Review date: 2026-09-23
- Inputs: command policy, subprocess call sites, and cancellation requirements

## Findings
- **HIGH**: `shell=True` accepted a user-provided branch name in the legacy helper; all affected calls now use an argument vector.
- **MEDIUM**: The parent environment exposed `DATABASE_URL`; the child environment now contains only `PATH`, `HOME`, `LANG`, and `GIT_TERMINAL_PROMPT=0`.
- **LOW**: A diagnostic command had no timeout; the shared launcher now applies a 10-second context.

## Controls Verified
- Executables are limited to `git`, `go`, `npm`, and `prumo`.
- Working directories resolve below the repository root.
- stdout and stderr are capped at 1 MiB each.
- Cancellation terminates the complete process group.

## Remediation and Evidence
- Added regression cases for `; rm -rf /`, newline arguments, and oversized output.
- `scripts/verify.sh` passes from the repository root.
- Audit journal records binary, argument hashes, duration, and exit status without secret values.
