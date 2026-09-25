#!/usr/bin/env bash
set -euo pipefail

TARGET_FILE="${1:-src/prumo/resources/workforce/skills/wireframe-styleguide/examples/wireframe_styleguide_sample.md}"

if [ ! -f "$TARGET_FILE" ]; then
    echo "Target styleguide file not found: $TARGET_FILE" >&2
    exit 1
fi

echo "Verifying wireframe styleguide token ramp, contrast math, and annotation specifications: $TARGET_FILE"

python3 - <<PYCHECK
import sys
import re

path = "$TARGET_FILE"
with open(path, 'r', encoding='utf-8') as f:
    content = f.read()

# 1. Verify Token Ramp
required_tokens = [
    "--wire-canvas",
    "--wire-surface",
    "--wire-border-strong",
    "--wire-text-primary",
    "--wire-text-secondary",
    "--wire-focus",
    "--wire-error"
]
missing_tokens = [t for t in required_tokens if t not in content]
if missing_tokens:
    print(f"Error: Missing core graybox tokens: {missing_tokens}", file=sys.stderr)
    sys.exit(1)

# 2. Contrast Mathematics Check
def hex_to_luminance(hex_str):
    hex_clean = hex_str.lstrip('#')
    r, g, b = [int(hex_clean[i:i+2], 16) / 255.0 for i in (0, 2, 4)]
    def linearize(c):
        return c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4
    return 0.2126 * linearize(r) + 0.7152 * linearize(g) + 0.0722 * linearize(b)

def contrast_ratio(hex1, hex2):
    l1 = hex_to_luminance(hex1)
    l2 = hex_to_luminance(hex2)
    lighter = max(l1, l2)
    darker = min(l1, l2)
    return (lighter + 0.05) / (darker + 0.05)

# Extract surface, text-primary, text-secondary
surface_match = re.search(r'--wire-surface:\s*(#[0-9a-fA-F]{6})', content)
primary_match = re.search(r'--wire-text-primary:\s*(#[0-9a-fA-F]{6})', content)
secondary_match = re.search(r'--wire-text-secondary:\s*(#[0-9a-fA-F]{6})', content)
border_match = re.search(r'--wire-border-strong:\s*(#[0-9a-fA-F]{6})', content)

if surface_match and primary_match:
    cr_prim = contrast_ratio(surface_match.group(1), primary_match.group(1))
    if cr_prim < 7.0:
        print(f"Warning: Primary text contrast ratio {cr_prim:.2f}:1 is below 7:1 (AAA)", file=sys.stderr)
    if cr_prim < 4.5:
        print(f"Error: Primary text contrast ratio {cr_prim:.2f}:1 fails WCAG AA (4.5:1)", file=sys.stderr)
        sys.exit(1)
    print(f"Verified: Primary text contrast ratio is {cr_prim:.2f}:1 (WCAG AA/AAA Pass)")

if surface_match and secondary_match:
    cr_sec = contrast_ratio(surface_match.group(1), secondary_match.group(1))
    if cr_sec < 4.5:
        print(f"Error: Secondary text contrast ratio {cr_sec:.2f}:1 fails WCAG AA (4.5:1)", file=sys.stderr)
        sys.exit(1)
    print(f"Verified: Secondary text contrast ratio is {cr_sec:.2f}:1 (WCAG AA Pass)")

if surface_match and border_match:
    cr_border = contrast_ratio(surface_match.group(1), border_match.group(1))
    if cr_border < 3.0:
        print(f"Error: Strong border contrast ratio {cr_border:.2f}:1 fails WCAG AA non-text (3.0:1)", file=sys.stderr)
        sys.exit(1)
    print(f"Verified: Strong border contrast ratio is {cr_border:.2f}:1 (WCAG AA Non-Text Pass)")

# 3. Check for Component State Matrix
if "Component State Matrix" not in content and "Default" not in content and "Focus-Visible" not in content:
    print("Error: Missing Component State Matrix specification", file=sys.stderr)
    sys.exit(1)

# 4. Check for Annotation Ledger
if "Annotation Ledger" not in content:
    print("Error: Missing View Schematic Annotation Ledger", file=sys.stderr)
    sys.exit(1)

# 5. Check for Responsive Breakpoints
if "Responsive Layout" not in content and "Breakpoints" not in content:
    print("Error: Missing Responsive Layout specification", file=sys.stderr)
    sys.exit(1)

print("All wireframe styleguide token, contrast, state, and annotation rules validated successfully.")
PYCHECK

echo "Wireframe styleguide verification complete."
exit 0
