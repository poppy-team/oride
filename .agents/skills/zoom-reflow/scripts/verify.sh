#!/usr/bin/env sh
# Verification script for zoom-reflow (Zoom & Reflow Testing - WCAG 2.2 SC 1.4.4 & SC 1.4.10)
set -e

echo "[Prumo Skill: zoom-reflow] Starting verification routine..."

# 1. Viewport Meta Scalability Audit (Prohibit user-scalable=no or maximum-scale restrictions)
echo "[1/3] Scanning for viewport zoom restrictions in HTML/templates..."
RESTRICTED_VIEWPORTS=$(grep -rnE 'user-scalable\s*=\s*(no|0)|maximum-scale\s*=\s*1(\.0)?' . \
  --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target --exclude-dir=build --exclude-dir=dist \
  --include="*.html" --include="*.jsx" --include="*.tsx" --include="*.vue" --include="*.svelte" 2>/dev/null || true)

if [ -n "$RESTRICTED_VIEWPORTS" ]; then
    echo "ERROR: Discovered viewport definitions blocking user zoom (violates WCAG 1.4.4):"
    echo "$RESTRICTED_VIEWPORTS"
    exit 1
fi

# 2. Check for rigid viewport widths exceeding 320px in CSS stylesheets without media/container query guard
echo "[2/3] Scanning for fixed rigid widths (>320px) in stylesheets..."
RIGID_WIDTHS=$(grep -rnE '^\s*width:\s*([4-9][0-9]{2}|[1-9][0-9]{3,})px' . \
  --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target --exclude-dir=build --exclude-dir=dist \
  --include="*.css" --include="*.scss" --include="*.less" 2>/dev/null || true)

if [ -n "$RIGID_WIDTHS" ]; then
    echo "WARNING: Hardcoded fixed widths detected. Ensure elements reflow at 320px width:"
    echo "$RIGID_WIDTHS" | head -n 5
fi

# 3. Example CSS validation
echo "[3/3] Validating example stylesheet syntax..."
if [ -f "examples/fluid-reflow.css" ]; then
    # Ensure clamp and minmax patterns are present
    grep -q "clamp(" "examples/fluid-reflow.css" || {
        echo "ERROR: examples/fluid-reflow.css missing fluid clamp() definitions"
        exit 1
    }
    grep -q "repeat(auto-fit" "examples/fluid-reflow.css" || {
        echo "ERROR: examples/fluid-reflow.css missing auto-fit responsive grid"
        exit 1
    }
fi

echo "[Prumo Skill: zoom-reflow] Verification completed successfully."
exit 0
