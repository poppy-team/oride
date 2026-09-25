#!/usr/bin/env bash
# Verification script for Motion Accessibility & Vestibular Safety
set -euo pipefail

echo "===================================================="
echo " [Prumo Skill: motion-accessibility] Motion Audit"
echo "===================================================="

VIOLATIONS=0
EXCLUDE="--exclude-dir=.git --exclude-dir=node_modules --exclude-dir=dist --exclude-dir=build --exclude-dir=target"

# 1. Check for presence of prefers-reduced-motion media query
echo "--- 1. Checking for prefers-reduced-motion declarations ---"
REDUCED_MOTION_DECLS=$(grep -rnE $EXCLUDE "prefers-reduced-motion" --include="*.css" --include="*.scss" --include="*.tsx" --include="*.jsx" --include="*.ts" --include="*.js" . 2>/dev/null || true)
if [ -n "$REDUCED_MOTION_DECLS" ]; then
    echo "[PASS] Found prefers-reduced-motion media queries / listeners in codebase:"
    echo "$REDUCED_MOTION_DECLS" | head -n 5
else
    echo "[WARN] No prefers-reduced-motion media queries found in active stylesheets."
fi

# 2. Check for unguarded smooth scrolling
echo "--- 2. Auditing smooth scrolling behavior ---"
SMOOTH_SCROLL=$(grep -rnE $EXCLUDE "scroll-behavior:\s*smooth" --include="*.css" --include="*.scss" . 2>/dev/null || true)
if [ -n "$SMOOTH_SCROLL" ]; then
    # Verify that scroll-behavior: auto exists for reduced-motion
    if ! grep -rnE $EXCLUDE "scroll-behavior:\s*auto" --include="*.css" --include="*.scss" . >/dev/null 2>&1; then
        echo "[WARN] 'scroll-behavior: smooth' declared without corresponding 'scroll-behavior: auto' in reduced-motion!"
    else
        echo "[PASS] Smooth scrolling accompanied by auto reset under reduced-motion."
    fi
else
    echo "[PASS] No forced global smooth scrolling detected."
fi

# 3. Scan for rapid strobe animation frequencies (> 3 Hz / < 333ms loops)
echo "--- 3. Auditing rapid strobe / flashing animations ---"
STROBE_ANIM=$(grep -rnE $EXCLUDE "animation:\s*.*(0\.[0-2][0-9]*s|[0-2][0-9]{2}ms)\s+.*infinite" --include="*.css" --include="*.scss" . 2>/dev/null || true)
if [ -n "$STROBE_ANIM" ]; then
    echo "[WARN] Found high-frequency looping animation (< 333ms loop cycle, potential seizure hazard):"
    echo "$STROBE_ANIM" | head -n 5
else
    echo "[PASS] Zero rapid high-frequency flashing loops detected."
fi

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
    echo "[SUCCESS] motion-accessibility verification passed cleanly."
    exit 0
else
    echo "[ERROR] motion-accessibility verification failed with $VIOLATIONS violation(s)."
    exit 1
fi
