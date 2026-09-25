# Defensive Bash Programming: Technical Reference Guide

## 1. Strict Mode Anatomy (`set -euo pipefail`)

The defensive header is the single most effective tool for preventing cascading failures in shell scripts:

```bash
#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'
```

### 1.1 Flag Semantics
- `set -e`: Aborts execution immediately upon command failure.
  - *Caution*: Does not trigger if command is part of an `if`, `while`, or `until` test, or inverted with `!`.
- `set -u`: Treats unset variables as fatal errors rather than expanding to empty strings.
  - *Example*: Prevents catastrophic commands like `rm -rf /tmp/$UNSET_VAR/*` resolving to `rm -rf /tmp//*`.
- `set -o pipefail`: Propagates pipeline failure.
  - In `cmd1 | cmd2 | cmd3`, default bash returns the exit code of `cmd3`. With `pipefail`, if `cmd1` fails, the whole pipeline fails with `cmd1`'s exit code.
- `IFS=$'\n\t'`: Reduces Internal Field Separator to newlines and tabs, eliminating whitespace splitting of filenames containing spaces.

---

## 2. Robust Signal Trapping (`trap`)

```bash
TMP_DIR="$(mktemp -d -t "prumo-job-XXXXXX")"

cleanup() {
    local exit_code=$?
    # Temporarily disable traps to prevent recursive trap loops
    trap - EXIT INT TERM HUP
    echo "[Bash] Cleaning up temporary working directory: $TMP_DIR" >&2
    rm -rf -- "$TMP_DIR"
    exit "$exit_code"
}

# Catch normal exit and termination signals
trap cleanup EXIT INT TERM HUP
```

---

## 3. Quoting Invariants & Pitfalls

| Syntax | Behavior | Safety |
| :--- | :--- | :--- |
| `"$var"` | Treats value as a single word | **Safe** |
| `$var` | Undergoes word splitting and globbing | **Unsafe** |
| `"${array[@]}"` | Expands each element as an individually quoted argument | **Safe** |
| `"${array[*]}"` | Combines all elements into a single argument string | Use only for display |
| `"$@"` | Passes all script arguments intact as individual words | **Safe** |
| `"$*"` | Collapses all script arguments into a single word | Avoid for delegation |

---

## 4. Atomic File & Directory Locking

POSIX systems guarantee that `mkdir` is an atomic system call (`mkdir(2)`):

```bash
acquire_lock() {
    local lock_dir="$1"
    local max_retries="${2:-5}"
    local count=0

    while ! mkdir "$lock_dir" 2>/dev/null; do
        count=$((count + 1))
        if [[ $count -ge $max_retries ]]; then
            echo "ERROR: Failed to acquire lock on $lock_dir after $max_retries attempts." >&2
            return 1
        fi
        sleep 1
    done
    return 0
}
```
