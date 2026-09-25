#!/usr/bin/env bash
# Verification script for Color Contrast Verification (WCAG 2.2 & APCA)
set -euo pipefail

echo "===================================================="
echo " [Prumo Skill: contrast] Color Contrast Audit"
echo "===================================================="

VIOLATIONS=0

# 1. Run mathematical contrast checker
echo "--- 1. Executing Mathematical WCAG 2.2 Contrast Gauntlet ---"
EXAMPLE="src/prumo/resources/workforce/skills/contrast/examples/contrast_checker.py"
if [ -f "$EXAMPLE" ]; then
    python3 "$EXAMPLE"
    echo "[PASS] Mathematical contrast verification passed cleanly."
else
    echo "[FAIL] Missing contrast checker example script."
    VIOLATIONS=$((VIOLATIONS + 1))
fi

# 2. Check for dual-cue accessibility (color not used alone)
echo "--- 2. Auditing Non-Color Status Cues ---"
EXCLUDE="--exclude-dir=.git --exclude-dir=node_modules --exclude-dir=dist --exclude-dir=target"
COLOR_ONLY=$(grep -rnE $EXCLUDE "(color:\s*(red|#f00|#ff0000|green|#0f0))" --include="*.css" --include="*.tsx" . 2>/dev/null | grep -v "icon" | grep -v "svg" | grep -v "aria" | head -n 5 || true)
if [ -n "$COLOR_ONLY" ]; then
    echo "[INFO] Found direct color status styling. Ensure accompanying icons or text labels exist:"
    echo "$COLOR_ONLY"
fi

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
    echo "[SUCCESS] contrast verification passed cleanly."
    exit 0
else
    echo "[ERROR] contrast verification failed with $VIOLATIONS violation(s)."
    exit 1
fi
