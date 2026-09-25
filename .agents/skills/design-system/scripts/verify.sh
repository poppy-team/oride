#!/usr/bin/env sh
# Verification script for design-system (Design System Engineering)
set -e

echo "[Prumo Skill: design-system] Starting verification routine..."

# 1. Validate design token JSON schema structure
echo "[1/3] Validating design token specification JSON..."
if [ -f "examples/design_token_spec.json" ]; then
    python3 -c "
import json
with open('examples/design_token_spec.json', 'r') as f:
    tokens = json.load(f)
required_sections = ['brand', 'semantic', 'spacing', 'typography']
for section in required_sections:
    if section not in tokens:
        raise ValueError(f'Missing required token category: {section}')
print('Design token specification valid and well-formed.')
"
fi

# 2. Syntax check on JavaScript extraction script
echo "[2/3] Checking Node.js token extraction tool syntax..."
if [ -f "scripts/extract_design_tokens.js" ] && command -v node >/dev/null 2>&1; then
    node -c scripts/extract_design_tokens.js
    echo "extract_design_tokens.js syntax is valid."
fi

# 3. Component static token audit (warn on raw hex colors in TSX/JSX components)
echo "[3/3] Checking for hardcoded color anti-patterns in component implementations..."
HARDCODED_HEX=$(grep -rnE '#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})\b' . \
  --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target --exclude-dir=build --exclude-dir=dist \
  --exclude="*.json" --exclude="*.md" --exclude="*.js" \
  --include="*.tsx" --include="*.jsx" 2>/dev/null || true)

if [ -n "$HARDCODED_HEX" ]; then
    echo "NOTICE: Raw hex colors found in components (ensure these are fallbacks only):"
    echo "$HARDCODED_HEX" | head -n 5
fi

echo "[Prumo Skill: design-system] Verification completed successfully."
exit 0
