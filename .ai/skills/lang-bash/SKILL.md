---
name: lang-bash
description: Production-grade Bash shell engineering, strict mode (set -euo pipefail), comprehensive quoting hygiene, trap-driven deterministic resource cleanup, safe file locking, and zero-eval security discipline.
---

# Bash Strict Defensive Engineering Contract

## 1. Purpose
Define, author, and audit hardened, deterministic shell scripts in Bash, guaranteeing strict error propagation (`set -euo pipefail`), total quoting discipline to eliminate word splitting and globbing vulnerabilities, deterministic trap-driven cleanup of temporary resources, atomic file operations, and zero code injection risks (`eval` proscription).

---

## 2. Use When
- Authoring CI/CD pipeline automation, developer environment bootstrap scripts, or build runners.
- Implementing local CLI wrappers, tooling orchestrators, or deployment hooks.
- Automating file system maintenance, backup routines, or container entrypoints.
- Managing low-level POSIX process orchestration where higher-level runtimes are unavailable.

---

## 3. Do Not Use When
- Building complex domain logic, data models, or microservices (use Go, Rust, or Python).
- Implementing mathematical calculations, 3D graphics, or raw memory allocations.
- Complex multi-threaded concurrent applications requiring fine-grained synchronization.

---

## 4. Required Context
Before writing or refactoring Bash scripts, verify:
1. **Target Shell Interpreter**: `bash` (version 4.0+) vs POSIX `/bin/sh` compatibility requirements.
2. **Environment Constraints**: Operating system tools (GNU coreutils on Linux vs BSD coreutils on macOS).
3. **Execution Privileges**: Unprivileged user execution vs elevated root/sudo operations.
4. **Interactive vs Non-Interactive**: Terminal output (colors, progress indicators) vs automated headless CI.

---

## 5. Procedure

### Step 1: Strict Execution Baseline
1. Begin every Bash script with the standard shebang and strict safety flags:
   ```bash
   #!/usr/bin/env bash
   set -euo pipefail
   IFS=$'\n\t'
   ```
   - `-e`: Exit immediately if any pipeline or command fails.
   - `-u`: Treat unset variables and parameters as errors and exit immediately.
   - `-o pipefail`: Return the exit status of the last command in the pipeline that failed, preventing silent intermediate pipe failures.
   - `IFS=$'\n\t'`: Split only on newlines and tabs by default, avoiding dangerous word splitting on spaces.

### Step 2: Parameter & Variable Expansion Discipline
1. Quote every variable and command substitution without exception:
   ```bash
   # CORRECT
   rm -rf -- "${TARGET_DIR:?Target directory must be defined}"
   cp -- "$source_file" "$dest_file"

   # PROHIBITED (Catastrophic vulnerability if TARGET_DIR is unset or contains spaces)
   # rm -rf $TARGET_DIR
   ```
2. Distinguish array expansions:
   - `"${array[@]}"`: Expands each element as a separate quoted word (REQUIRED).
   - `"${array[*]}"`: Concatenates all elements into a single word separated by IFS.
3. Use defensive parameter expansion:
   - `${var:-default}`: Provide fallback default if unset.
   - `${var:?error message}`: Abort with error message if unset or null.

### Step 3: Trap-Driven Resource Cleanup
1. Never leave temporary files or directories upon script termination or interruption.
2. Pair every temporary creation (`mktemp`) with an immediate `trap` handler:
   ```bash
   WORK_DIR="$(mktemp -d -t "prumo-runner-XXXXXX")"
   cleanup() {
       local exit_code=$?
       trap - EXIT INT TERM HUP
       rm -rf -- "$WORK_DIR"
       exit "$exit_code"
   }
   trap cleanup EXIT INT TERM HUP
   ```
3. Ensure trap handlers preserve the original exit code of the failing command.

### Step 4: Atomic Locking & Concurrency Defense
1. When multiple scripts or background processes may access shared resources, use atomic directory creation or `flock`:
   ```bash
   LOCKDIR="/tmp/prumo-build.lock"
   if ! mkdir "$LOCKDIR" 2>/dev/null; then
       echo "ERROR: Process is already running (locked by $LOCKDIR)" >&2
       exit 1
   fi
   trap 'rmdir "$LOCKDIR" 2>/dev/null || true' EXIT
   ```

### Step 5: Input Parsing & Dependency Verification
1. Validate external command dependencies before execution:
   ```bash
   for cmd in git jq curl; do
       command -v "$cmd" >/dev/null 2>&1 || {
           echo "ERROR: Required command '$cmd' is not installed or not in PATH." >&2
           exit 1
       }
   done
   ```
2. Parse command-line flags defensively using POSIX `getopts` or explicit flag matching with full `--help` documentation.

---

## 6. Decision Rules
1. **Absolute Eval Proscription**: Never call `eval` on dynamic, external, or user-supplied variables. Use arrays or bash indirect parameter expansions (`${!var}`) instead.
2. **Never Ignore Exit Codes**: If a command is legitimately expected to fail, handle it explicitly:
   ```bash
   command_that_may_fail || true
   # or
   if ! result=$(command); then ...
   ```
3. **No Unchecked `cd`**: Never execute `cd` without an explicit guard:
   ```bash
   cd -- "$target_dir" || { echo "Failed to enter $target_dir" >&2; exit 1; }
   ```
4. **ShellCheck Compliance**: All scripts must pass ShellCheck without warnings when available.

---

## 7. Evidence Required
- **Syntax Check**: `bash -n script.sh` exits with 0 errors.
- **Strict Mode Verification**: `set -euo pipefail` confirmed at head of script.
- **Trap Cleanup Execution**: Automated test asserting temporary directories are deleted upon exit.

---

## 8. Output Contract
A production Bash artifact must provide:
1. Hardened executable script (`chmod +x script.sh`) with `set -euo pipefail`.
2. Fully quoted variable expansions and array handling.
3. Clean error reporting with output directed to stderr (`>&2`).
4. Trap signal handlers covering `EXIT`, `INT`, `TERM`, `HUP`.

---

## 9. Stop Conditions
- Script syntax verified with `bash -n`.
- All temporary directories cleaned up on both success and early error exits.
- No unquoted parameter expansions.

---

## 10. Escalation Rules
- Escalate to Lead Architect if script complexity exceeds 300 lines of procedural shell code; refactor into a compiled language (Go, Rust).
- Escalate to Security Officer if a script requires root privileges or handles secrets/credentials.
