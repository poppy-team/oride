#!/usr/bin/env sh
# Verification script for Screen Reader Compatibility
set -eu

TARGET_DIR="${1:-.}"
echo "=== [screen-reader] Auditing Accessible Names & Semantics in $TARGET_DIR ==="

EXCLUDE_ARGS='--exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target --exclude-dir=archive'
INCLUDE_UI='--include=*.html --include=*.jsx --include=*.tsx --include=*.vue --include=*.svelte --include=*.js --include=*.ts'

# 1. Check for icon-only buttons missing aria-label
echo "[1/3] Scanning for buttons lacking visible text or aria-label..."
EMPTY_BUTTONS=$(grep -rnE $EXCLUDE_ARGS $INCLUDE_UI '<button[^>]*>\s*<(svg|i|span class=)[^>]*>\s*</button>' "$TARGET_DIR" 2>/dev/null | grep -v "aria-label" || true)
if [ -n "$EMPTY_BUTTONS" ]; then
    echo "WARNING: Potential empty accessible name detected on button:"
    echo "$EMPTY_BUTTONS"
else
    echo "Button naming audit: PASS."
fi

# 2. Check for images missing alt attribute
echo "[2/3] Checking for <img> tags without alt attribute..."
MISSING_ALT=$(grep -rnE $EXCLUDE_ARGS $INCLUDE_UI '<img ' "$TARGET_DIR" 2>/dev/null | grep -v 'alt=' || true)
if [ -n "$MISSING_ALT" ]; then
    echo "WARNING: Images missing alt attribute:"
    echo "$MISSING_ALT"
else
    echo "Image alt audit: PASS."
fi

# 3. Check for form inputs lacking id for label association
echo "[3/3] Scanning inputs for id attributes..."
UNLABELED_INPUTS=$(grep -rnE $EXCLUDE_ARGS $INCLUDE_UI '<input ' "$TARGET_DIR" 2>/dev/null | grep -v 'id=' | grep -v 'aria-label' | head -n 5 || true)
if [ -n "$UNLABELED_INPUTS" ]; then
    echo "INFO: Verify that inputs without direct id are wrapped by <label> tags:"
    echo "$UNLABELED_INPUTS"
fi

echo "=== [screen-reader] Screen Reader Audit Complete ==="
exit 0
