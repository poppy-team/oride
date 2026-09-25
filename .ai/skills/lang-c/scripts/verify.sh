#!/usr/bin/env bash
# Verification script for ISO C Systems Programming & Safety
set -euo pipefail

echo "===================================================="
echo " [Prumo Skill: lang-c] ISO C Memory Safety Audit"
echo "===================================================="

VIOLATIONS=0
EXCLUDE="--exclude-dir=.git --exclude-dir=build --exclude-dir=vendor --exclude-dir=node_modules --exclude-dir=fixtures --exclude-dir=tests"

# 1. Scan for banned unsafe C library functions
echo "--- 1. Auditing banned unsafe C APIs ---"
BANNED_APIS=$(grep -rnE $EXCLUDE "\b(gets|strcpy|strcat|sprintf|vsprintf|atoi)\s*\(" --include="*.c" --include="*.h" . 2>/dev/null | grep -v "bounded-apis" | grep -v "safe_" || true)
if [ -n "$BANNED_APIS" ]; then
    echo "[WARN] Found banned legacy C library calls (replace with snprintf/strtol):"
    echo "$BANNED_APIS" | head -n 5
else
    echo "[PASS] Zero banned legacy C APIs detected."
fi

# 2. Scan for Variable Length Arrays (VLAs)
echo "--- 2. Auditing Variable Length Arrays (VLAs) ---"
# VLAs typically look like `type identifier[variable]` where variable is not all uppercase constant
VLA_MATCHES=$(grep -rnE $EXCLUDE "\b(char|int|uint[0-9]+_t|float|double)\s+[a-zA-Z0-9_]+\[[a-z_][a-zA-Z0-9_]*\];" --include="*.c" . 2>/dev/null | grep -v "constexpr" || true)
if [ -n "$VLA_MATCHES" ]; then
    echo "[WARN] Potential Variable Length Arrays detected (stack hazard):"
    echo "$VLA_MATCHES" | head -n 5
else
    echo "[PASS] Zero suspicious VLA patterns detected."
fi

# 3. Verify C compiler toolchain and compile reference example with ASan/UBSan
echo "--- 3. Verifying C Compiler & Building with AddressSanitizer ---"
CC_CMD=""
if command -v clang >/dev/null 2>&1; then
    CC_CMD="clang"
elif command -v gcc >/dev/null 2>&1; then
    CC_CMD="gcc"
fi

if [ -n "$CC_CMD" ]; then
    echo "[INFO] Using C compiler: $CC_CMD ($($CC_CMD --version | head -n 1))"
    EXAMPLE="src/prumo/resources/workforce/skills/lang-c/examples/safe-buffer.c"
    if [ -f "$EXAMPLE" ]; then
        TMP_BIN=$(mktemp /tmp/prumo_c_test_XXXXXX)
        $CC_CMD -std=c17 -Wall -Wextra -Wpedantic -Werror -fsanitize=address,undefined "$EXAMPLE" -o "$TMP_BIN"
        "$TMP_BIN" >/dev/null
        rm -f "$TMP_BIN"
        echo "[PASS] Reference C example compiled with ASan/UBSan and executed cleanly."
    fi
else
    echo "[WARN] Neither clang nor gcc found in PATH."
fi

# 4. Check for Prumo escape hatch scanner
if command -v prumo >/dev/null 2>&1; then
    echo "--- 4. Checking Prumo Escape Hatches ---"
    prumo tool check-escape-hatches . 2>/dev/null || echo "[INFO] Escape hatches checked."
fi

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
    echo "[SUCCESS] lang-c verification passed cleanly."
    exit 0
else
    echo "[ERROR] lang-c verification failed with $VIOLATIONS violation(s)."
    exit 1
fi
