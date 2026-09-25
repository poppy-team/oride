---
name: process-execution-security
description: Command injection elimination, strict binary allowlisting, sanitized environment variables, execution timeouts, and process group reaping
---
# Subprocess & Execution Security

## 1. Command Injection Elimination
Never pass raw formatted strings to a shell interpreter (/bin/sh -c, cmd.exe). Always execute commands using direct argument vectors (exec.Command(binary, arg1, arg2)) where arguments are passed as isolated strings without shell parsing.

## 2. Strict Binary Allowlisting
Execute only binaries present on an explicit, pre-approved project allowlist (e.g. git, go, npm, cargo, prumo). Reject execution of arbitrary or user-specified binaries not registered in the project capability policy.

## 3. Environment Variable Sanitization
Never inherit the full environment of the parent process. Construct an isolated environment slice containing only necessary, sanitized variables (e.g. PATH, HOME, LANG). Strip API keys, tokens, and credentials.

## 4. Working Directory Confinement
Confine subprocess execution strictly within the current project repository root directory. Validate that the working directory exists and is owned by the current user before launching.

## 5. Mandatory Execution Timeouts
Always bind subprocess lifecycles to an explicit context with timeout (context.WithTimeout). Default to a 60-second limit for developer commands and 10 seconds for diagnostics, preventing runaway deadlocks.

## 6. Output Buffer Limiting
Wrap stdout and stderr capture buffers with bounded readers (io.LimitReader) to prevent unbounded subprocess output from consuming all system RAM.

## 7. Privilege Demotion
Execute all subprocesses under the current unprivileged user. Explicitly forbid execution under root/superuser accounts. Block commands attempting sudo, su, or setuid elevation.

## 8. Process Group Cleanup & Orphan Reaping
Configure child processes with a dedicated process group (Setpgid: true). On cancellation or timeout, send SIGKILL to the entire process group (-pgid) to guarantee zero orphaned background processes.

## 9. Banned Construct Scanners
Run automated static checks across the project codebase to detect and ban dangerous dynamic execution functions (eval(), exec(), system(), popen(), Function()).

## 10. Process Audit Journal
Log every command invocation to .prumo/history/journal.json: capture binary name, sanitized argument list, working directory, start time, duration, exit code, and captured error output.
