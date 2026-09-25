#!/usr/bin/env sh
# Verification script for UI Component Implementation
set -eu

TARGET_DIR="${1:-.}"
echo "=== [ui-implementation] Auditing UI Component Standards in $TARGET_DIR ==="

EXCLUDE_ARGS='--exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target --exclude-dir=archive'

# 1. Check for raw hardcoded color hex values (#fff, #0070f3, etc)
echo "[1/3] Scanning for hardcoded color hex values (violates design tokens)..."
HEX_MATCHES=$(grep -rnE $EXCLUDE_ARGS "#[0-9a-fA-F]{3,8}" "$TARGET_DIR" 2>/dev/null | grep -E "\.(tsx|jsx|vue|svelte|css|html)$" | head -n 10 || true)
if [ -n "$HEX_MATCHES" ]; then
    echo "WARNING: Hardcoded hex colors detected. Replace with design tokens (var(--color-*)):"
    echo "$HEX_MATCHES"
else
    echo "Design tokens check: Clean."
fi

# 2. Check for interactive elements without keyboard access or focus outline suppression
echo "[2/3] Checking for 'outline: none' or 'outline: 0' without focus replacement..."
OUTLINE_NONE=$(grep -rnE $EXCLUDE_ARGS "outline:\s*(none|0)" "$TARGET_DIR" 2>/dev/null | grep -v "focus-visible" | head -n 5 || true)
if [ -n "$OUTLINE_NONE" ]; then
    echo "WARNING: Potential focus indicator suppression detected:"
    echo "$OUTLINE_NONE"
else
    echo "Focus indicator check: Clean."
fi

# 3. Check for component tests
echo "[3/3] Scanning for component test suites..."
TEST_COUNT=$(find "$TARGET_DIR" -not -path "*/.*" -name "*.test.tsx" -o -name "*.spec.tsx" -o -name "*.stories.tsx" 2>/dev/null | wc -l)
echo "Found $TEST_COUNT component test/story files."

echo "=== [ui-implementation] UI Component Audit Complete ==="
exit 0
