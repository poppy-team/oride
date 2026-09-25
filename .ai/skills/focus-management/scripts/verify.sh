#!/usr/bin/env sh
# Verification script for Focus Management & Transition Architecture
set -eu

TARGET_DIR="${1:-.}"
echo "=== [focus-management] Auditing Focus Traps & Restoration in $TARGET_DIR ==="

EXCLUDE_ARGS='--exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target --exclude-dir=archive'
INCLUDE_UI='--include=*.html --include=*.jsx --include=*.tsx --include=*.vue --include=*.svelte --include=*.js --include=*.ts'

# 1. Scan for custom modal dialogs
echo "[1/3] Scanning for modal dialog components..."
MODALS=$(grep -rnE $EXCLUDE_ARGS $INCLUDE_UI 'role=["'"'"']dialog["'"'"']' "$TARGET_DIR" 2>/dev/null || true)
if [ -n "$MODALS" ]; then
    echo "Found modal dialog definitions:"
    echo "$MODALS"
fi

# 2. Check for presence of focus restoration logic
echo "[2/3] Verifying focus restoration hooks (activeElement caching)..."
if grep -rq $EXCLUDE_ARGS $INCLUDE_UI "document.activeElement" "$TARGET_DIR" 2>/dev/null; then
    echo "activeElement caching: PASS."
else
    echo "WARNING: No references to 'document.activeElement' found in UI code. Ensure modals restore focus upon close."
fi

# 3. Check for Escape keydown listener
echo "[3/3] Checking for Escape key dismissal..."
if grep -rq $EXCLUDE_ARGS $INCLUDE_UI "key === 'Escape'" "$TARGET_DIR" 2>/dev/null || grep -rq $EXCLUDE_ARGS $INCLUDE_UI 'key === "Escape"' "$TARGET_DIR" 2>/dev/null; then
    echo "Escape key dismissal: PASS."
else
    echo "WARNING: No Escape key listeners found on dialogs/menus."
fi

echo "=== [focus-management] Focus Audit Complete ==="
exit 0
