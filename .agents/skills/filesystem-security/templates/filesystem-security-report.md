# Filesystem Security Audit — Workspace Import

## Scope
- **Component**: archive import and project extraction
- **Authorized root**: `/srv/app/workspaces/{workspaceId}`
- **Review date**: 2026-09-23
- **Assessor**: filesystem-security-agent

## Finding FS-01
- **Severity**: Critical
- **Observed**: Entry names were normalized but the extractor followed an archive symlink into `/etc`.
- **Impact**: An archive could overwrite files outside the workspace.
- **Remediation**: Reject links and special files; resolve every destination beneath the authorized root; use no-follow opens.

## Finding FS-02
- **Severity**: High
- **Observed**: Extraction expanded one entry to 2.4 GB before enforcing a post-write limit.
- **Remediation**: Enforce declared and observed byte, entry-count, nesting, and compression-ratio limits while streaming.

## Control Verification
- Private workspace directories use `0700`; user files use `0600`.
- Writes use exclusive temporary files, `fsync`, and atomic rename in the destination filesystem.
- Temporary data is removed on success, failure, and cancellation.
- Abusive entries: `../`, absolute paths, links, devices, duplicate normalized names, and oversized streams are rejected.

## Release Decision
FS-01 and FS-02 block release. Re-run the adversarial archive corpus after remediation.
