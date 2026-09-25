#!/usr/bin/env sh
# Verification script for Keyboard Navigation & Operability
set -eu

TARGET_DIR="${1:-.}"
echo "=== [keyboard-accessibility] Auditing Keyboard Navigation in $TARGET_DIR ==="

EXCLUDE_ARGS='--exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target --exclude-dir=archive'
INCLUDE_UI='--include=*.html --include=*.jsx --include=*.tsx --include=*.vue --include=*.svelte --include=*.js --include=*.ts'

# 1. Scan for forbidden positive tabindexes
echo "[1/3] Scanning for positive tabindexes (tabindex > 0)..."
POS_TAB=$(grep -rnE $EXCLUDE_ARGS $INCLUDE_UI 'tabindex=["'"'"'][1-9][0-9]*["'"'"']' "$TARGET_DIR" 2>/dev/null || true)
if [ -n "$POS_TAB" ]; then
    echo "ERROR: Forbidden positive tabindexes detected (breaks natural tab order):"
    echo "$POS_TAB"
    exit 1
else
    echo "Tabindex audit: PASS (0 positive tabindexes found in UI components)."
fi

# 2. Check for click handlers on non-interactive elements without role/tabindex/keydown
echo "[2/3] Scanning for unkeyed onClick handlers on divs/spans..."
DIV_CLICKS=$(grep -rnE $EXCLUDE_ARGS $INCLUDE_UI '<(div|span|section|p)[^>]*onClick' "$TARGET_DIR" 2>/dev/null | grep -v "role=" | grep -v "tabIndex" | head -n 5 || true)
if [ -n "$DIV_CLICKS" ]; then
    echo "WARNING: Potential mouse-only clickable elements detected:"
    echo "$DIV_CLICKS"
else
    echo "Clickable elements audit: PASS."
fi

# 3. Check for outline suppression
echo "[3/3] Checking for blind 'outline: none'..."
BLIND_OUTLINES=$(grep -rnE $EXCLUDE_ARGS $INCLUDE_UI 'outline:\s*(none|0)' "$TARGET_DIR" 2>/dev/null | grep -v "focus-visible" | head -n 5 || true)
if [ -n "$BLIND_OUTLINES" ]; then
    echo "WARNING: Potential focus outline suppressions detected without focus-visible:"
    echo "$BLIND_OUTLINES"
else
    echo "Focus outlines: PASS."
fi

echo "=== [keyboard-accessibility] Keyboard Navigation Audit Complete ==="
exit 0
