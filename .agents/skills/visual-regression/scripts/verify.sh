#!/usr/bin/env bash
# Verification script for visual-regression (Automated Visual Regression Testing)
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
cd "${SKILL_DIR}"

echo "[Prumo Skill: visual-regression] Starting verification routine..."

# 1. Syntax check on Playwright config and comparison scripts
echo "[1/3] Checking Node.js script syntax..."
if command -v node >/dev/null 2>&1; then
    if [ -f "examples/playwright_vrt_config.js" ]; then
        node -c examples/playwright_vrt_config.js
        echo "playwright_vrt_config.js syntax valid."
    fi
    if [ -f "scripts/compare_screenshots.js" ]; then
        node -c scripts/compare_screenshots.js
        echo "compare_screenshots.js syntax valid."
    fi
else
    echo "NOTICE: node not found; skipping JS syntax checks."
fi

# 2. Threshold verification in configuration
echo "[2/3] Checking threshold configuration in Playwright config..."
if [ -f "examples/playwright_vrt_config.js" ]; then
    grep -q "threshold" "examples/playwright_vrt_config.js" || {
        echo "ERROR: Missing threshold configuration in playwright_vrt_config.js"
        exit 1
    }
fi

# 3. Template verification
echo "[3/3] Checking visual regression report template..."
if [ ! -f "templates/visual-regression-report.md" ]; then
    echo "ERROR: Missing templates/visual-regression-report.md"
    exit 1
fi

echo "[Prumo Skill: visual-regression] Verification completed successfully."
exit 0
