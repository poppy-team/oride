#!/usr/bin/env bash
# Verification script for lang-bash (Bash Strict Defensive Engineering)
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
cd "${SKILL_DIR}"

echo "[Prumo Skill: lang-bash] Starting verification routine..."

# 1. Syntax check on all shell scripts using bash -n
echo "[1/4] Running syntax checks (bash -n)..."
SHELL_SCRIPTS=()
while IFS= read -r -d '' script; do
    SHELL_SCRIPTS+=("$script")
done < <(find "${SKILL_DIR}" -type f \( -name "*.sh" -o -name "*.bash" \) -print0)

SYNTAX_ERRORS=0
for script in "${SHELL_SCRIPTS[@]}"; do
    if ! bash -n "$script" 2>&1; then
        echo "ERROR: Syntax error detected in $script"
        SYNTAX_ERRORS=1
    fi
done

if [ "$SYNTAX_ERRORS" -ne 0 ]; then
    echo "ERROR: One or more lang-bash scripts failed syntax validation."
    exit 1
fi

# 2. Strict flags audit (verify set -e or set -euo pipefail is declared)
echo "[2/4] Auditing scripts for strict error handling flags..."
for script in "${SHELL_SCRIPTS[@]}"; do
    if ! grep -qE 'set\s+-[a-zA-Z0-9]*e' "$script"; then
        echo "WARNING: Script $script missing 'set -e' or 'set -euo pipefail'"
    fi
done

# 3. Prohibit eval on dynamic variables
echo "[3/4] Scanning for dangerous eval usage..."
EVAL_CALLS=""
if [ "${#SHELL_SCRIPTS[@]}" -gt 0 ]; then
    EVAL_CALLS=$(grep -HnE '\beval\s+["\$\`]' "${SHELL_SCRIPTS[@]}" 2>/dev/null || true)
fi

if [ -n "$EVAL_CALLS" ]; then
    echo "ERROR: Discovered potentially dangerous eval statement:"
    echo "$EVAL_CALLS"
    exit 1
fi

# 4. Functional execution of example runner
echo "[4/4] Executing safe_runner.sh example test..."
if [ -x "examples/safe_runner.sh" ]; then
    ./examples/safe_runner.sh >/dev/null
    echo "safe_runner.sh executed cleanly."
else
    echo "ERROR: examples/safe_runner.sh is missing or not executable."
    exit 1
fi

echo "[Prumo Skill: lang-bash] Verification completed successfully."
exit 0
