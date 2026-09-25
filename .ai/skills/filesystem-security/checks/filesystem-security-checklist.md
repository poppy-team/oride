# Filesystem Security & Access Control Checklist

- [ ] Path traversal eliminated via filepath.Clean and workspace prefix validation
- [ ] Symlinks evaluated and blocked from pointing outside project workspace
- [ ] Files created using O_CREATE|O_EXCL or atomic temporary write-and-rename
- [ ] Permissions restricted: 0600 for private files, 0700 for private dirs; 0777 banned
- [ ] File read buffers bounded via io.LimitReader to prevent memory exhaustion
- [ ] Protected directories (.git/, .prumo/credentials) completely shielded from writes
- [ ] Temporary files cleaned up deterministically on all exit code paths
