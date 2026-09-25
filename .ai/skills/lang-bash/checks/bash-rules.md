# Bash Strict Defensive Engineering Checklist

## 1. Strict Execution Flags
- [ ] **Shebang**: Declares appropriate shebang (`#!/usr/bin/env bash` or `#!/bin/sh`).
- [ ] **Strict Safety Options**: `set -euo pipefail` declared immediately below shebang.
- [ ] **Field Separator**: `IFS=$'\n\t'` configured if handling filename or line-based iterations.

## 2. Quoting & Parameter Expansion Hygiene
- [ ] **Double Quoting**: All variable references (`"$var"`) and command substitutions (`"$(cmd)"`) are enclosed in double quotes.
- [ ] **Array Expansion**: Array elements expanded with `"${array[@]}"`; never bare `$array` or unquoted `${array[*]}`.
- [ ] **Defensive Defaults**: Unset variables safeguarded with `${var:-default}` or `${var:?error}`.
- [ ] **Catastrophic Path Defense**: Operations like `rm -rf -- "$dir"` verify that `$dir` is non-empty and not `/`.

## 3. Resource Management & Trap Handlers
- [ ] **Trap Registration**: All `mktemp` directories/files paired with `trap cleanup EXIT INT TERM HUP`.
- [ ] **Exit Code Preservation**: Cleanup procedures capture `$?` and re-exit with the original error code.
- [ ] **Locking**: Shared resources protected by atomic `mkdir` or `flock`.

## 4. Security & Safety Proscriptions
- [ ] **Zero Eval**: No usage of `eval` on dynamic variables.
- [ ] **Safe Directory Changes**: All `cd` commands guarded: `cd -- "$dir" || exit 1`.
- [ ] **Error Streams**: Error messages and warnings emitted strictly to standard error (`>&2`).
- [ ] **ShellCheck Clean**: Script passes `bash -n` and `shellcheck` with zero warnings.
