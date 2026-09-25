#!/usr/bin/env bash
# Verification script for Prumo Canonical Navigation & Authority Resolution.
# Run from the repository root: bash src/prumo/resources/workforce/skills/prumo-navigation/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/prumo-navigation"

echo "===================================================="
echo " [Prumo Skill: prumo-navigation] Navigation Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/prumo-navigation-checklist.md" \
         "references/prumo-navigation-guide.md" \
         "templates/prumo-navigation-spec.md" "scripts/verify.sh"; do
  if [ -s "$SKILL_DIR/$f" ]; then
    pass "Found non-empty $f."
  else
    fail "Missing or empty $SKILL_DIR/$f."
  fi
done

# 2. Stale superseded artifacts must be gone
echo "--- 2. Checking for stale superseded artifacts ---"
for f in "templates/prumo-navigation-template.md"; do
  if [ -e "$SKILL_DIR/$f" ]; then
    fail "Stale artifact still present: $SKILL_DIR/$f."
  else
    pass "Stale artifact absent: $f."
  fi
done

# 3. Manifest validity and asset wiring
echo "--- 3. Validating manifest ---"
if command -v python3 >/dev/null 2>&1; then
  if ! python3 - "$SKILL_DIR" <<'EOF'
import json
import os
import sys

skill_dir = sys.argv[1]
errors = []

with open(os.path.join(skill_dir, "manifest.json"), encoding="utf-8") as fh:
    manifest = json.load(fh)

if manifest.get("id") != "prumo-navigation":
    errors.append("manifest id must be 'prumo-navigation'")
if manifest.get("version", 0) < 2:
    errors.append("manifest version must be >= 2")
if manifest.get("schema_version", 0) < 3:
    errors.append("manifest schema_version must be >= 3")

valid_modes = {"implementation", "review", "audit", "research",
               "design", "testing", "documentation", "release"}
for mode in manifest.get("modes", []):
    if mode not in valid_modes:
        errors.append(f"invalid mode {mode!r} in manifest")

for key in ("references", "templates", "checks", "scripts"):
    for rel in manifest.get(key, []):
        if not os.path.isfile(os.path.join(skill_dir, rel)):
            errors.append(f"manifest {key} entry missing on disk: {rel}")

if errors:
    for err in errors:
        print(f"[FAIL] {err}")
    sys.exit(1)
print("[PASS] Manifest is valid and assets are wired.")
EOF
  then
    fail "Manifest validation failed (see details above)."
  fi
else
  echo "[SKIP] python3 not found; skipping manifest validation."
fi

# 4. Content invariants
echo "--- 4. Checking navigation content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "authority" "SKILL.md" "SKILL.md enforces authority order"
check_grep "cite|citation" "SKILL.md" "SKILL.md mandates citations"
check_grep "drift" "SKILL.md" "SKILL.md handles canonical drift"
check_grep "progressive|smallest sufficient" "SKILL.md" "SKILL.md mandates progressive context"
check_grep "AUTHORITY_MAP" "references/prumo-navigation-guide.md" "Reference guide covers ownership map"
check_grep "Expansion Accounting|trail" "templates/prumo-navigation-spec.md" "Navigation template records the trail"
check_grep "Living Book" "checks/prumo-navigation-checklist.md" "Checklist bounds Living Book fetches"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] prumo-navigation verification passed cleanly."
  exit 0
else
  echo "[ERROR] prumo-navigation verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
