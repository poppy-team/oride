#!/usr/bin/env bash
# Verification script for Documentation Engineering & Canonical Markdown Publishing.
# Run from the repository root: bash src/prumo/resources/workforce/skills/documentation/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/documentation"

echo "===================================================="
echo " [Prumo Skill: documentation] Documentation Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/documentation-checklist.md" \
         "references/documentation-guide.md" \
         "templates/documentation-spec.md" "scripts/verify.sh"; do
  if [ -s "$SKILL_DIR/$f" ]; then
    pass "Found non-empty $f."
  else
    fail "Missing or empty $SKILL_DIR/$f."
  fi
done

# 2. Manifest validity and asset cross-references
echo "--- 2. Validating manifest and asset cross-references ---"
if command -v python3 >/dev/null 2>&1; then
  if ! python3 - "$SKILL_DIR" <<'EOF'
import json
import os
import sys

skill_dir = sys.argv[1]
errors = []

with open(os.path.join(skill_dir, "manifest.json"), encoding="utf-8") as fh:
    manifest = json.load(fh)

if manifest.get("id") != "documentation":
    errors.append("manifest id must be 'documentation'")
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
print("[PASS] Manifest is valid and all listed assets exist on disk.")
EOF
  then
    fail "Manifest validation failed (see details above)."
  fi
else
  echo "[SKIP] python3 not found; skipping manifest validation."
fi

# 3. Content invariants: authority, drift, links, publishing
echo "--- 3. Checking documentation content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "canonical|authority" "SKILL.md" "SKILL.md covers documentation authority"
check_grep "drift" "SKILL.md" "SKILL.md mandates drift checks against code"
check_grep "lychee|broken link" "SKILL.md" "SKILL.md mandates link integrity"
check_grep "owner" "SKILL.md" "SKILL.md mandates page ownership"
check_grep "drift" "checks/documentation-checklist.md" "Checklist covers drift"
check_grep "lychee" "checks/documentation-checklist.md" "Checklist covers link checking"
check_grep "review.*cadence|stale" "checks/documentation-checklist.md" "Checklist covers review cadence"
check_grep "canonical|projection|historical" "references/documentation-guide.md" "Reference guide covers authority roles"
check_grep "anti-pattern" "references/documentation-guide.md" "Reference guide covers anti-patterns"
check_grep "drift|lychee" "templates/documentation-spec.md" "Doc template carries drift and link evidence"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] documentation verification passed cleanly."
  exit 0
else
  echo "[ERROR] documentation verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
