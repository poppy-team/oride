# Subprocess & Execution Security Checklist

- [ ] Shell interpreters (/bin/sh, cmd.exe) bypassed; direct argument vectors used
- [ ] Subprocess commands restricted to pre-approved binary allowlist
- [ ] Parent process environment variables sanitized; secret tokens stripped
- [ ] Mandatory context timeout enforced on every command (default <= 60s)
- [ ] Stdout and stderr capture buffers strictly bounded to prevent RAM exhaustion
- [ ] Process groups terminated cleanly on cancellation to prevent orphan processes
- [ ] Every executed command recorded in audit journal with exit status and timing
