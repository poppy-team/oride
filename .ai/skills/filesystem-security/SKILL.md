---
name: filesystem-security
description: Path traversal defense, symlink attack mitigation, TOCTOU prevention, POSIX permissions, atomic file replacement, and protected directories
---
# Filesystem Security & Access Control

## 1. Path Traversal Defense
Normalize and sanitize all file paths using canonical library methods (e.g. filepath.Clean). Verify that the resolved target path is strictly prefixed by the authorized workspace root before performing any read, write, or delete operation.

## 2. Symlink & Hardlink Attack Prevention
Resolve all symbolic links to their absolute target paths using filepath.EvalSymlinks. Explicitly reject any symlink whose target destination lies outside the designated project root directory.

## 3. Time-of-Check to Time-of-Use (TOCTOU) Mitigation
Avoid vulnerable check-then-open sequences. Use atomic file descriptor operations (e.g. os.OpenFile with O_CREATE|O_EXCL) to guarantee that the file examined is the exact file modified without race window exploitation.

## 4. Strict POSIX Permission Discipline
Enforce restrictive file permissions: 0600 (read/write by owner only) for private files and credentials; 0700 for private directories; 0644 for public source files; 0755 for executables. Never create world-writable files (0666/0777).

## 5. Safe Temporary File Management
Create temporary files exclusively via secure system utilities (os.CreateTemp) within designated temporary directories. Set restrictive permissions on temporary files and guarantee cleanup using defer os.Remove.

## 6. Atomic File Replacement
Never overwrite production files directly in-place. Write new content to an adjacent temporary file in the same filesystem directory, call f.Sync() to flush buffers to disk, and execute an atomic rename (os.Rename).

## 7. File Size & Memory Caps
Guard file reads against memory exhaustion attacks using bounded readers (io.LimitReader). Enforce an explicit maximum file size threshold (e.g. max 50MB) for automated ingestion and tool parsing.

## 8. Protected Directory Shield
Hardcode inviolable security barriers that prevent modifications to critical infrastructure paths: .git/, .prumo/credentials, system binary directories, and root filesystem locations.

## 9. Secure Data Wiping
When deleting temporary files containing decrypted keys or sensitive tokens, overwrite the disk blocks with random bytes or zeros prior to unlinking to prevent data recovery from raw block storage.

## 10. Filesystem Audit Logging
Record all file creation, modification, and deletion events in the project security audit log, capturing relative path, actor ID, timestamp, and post-modification cryptographic checksum.
