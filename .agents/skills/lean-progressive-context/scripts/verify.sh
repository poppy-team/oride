#!/usr/bin/env bash
# Verification script for Lean Progressive Context.
# Run from the repository root: bash src/prumo/resources/workforce/skills/lean-progressive-context/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/lean-progressive-context"

echo "===================================================="
echo " [Prumo Skill: lean-progressive-context] Lean Progressive Context Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/lean-progressive-context-checklist.md" \
         "references/lean-progressive-context-guide.md" \
         "templates/lean-progressive-context-spec.md" "scripts/verify.sh"; do
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

if manifest.get("id") != "lean-progressive-context":
    errors.append("manifest id must be 'lean-progressive-context'")
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
print("[PASS] Manifest is valid and all listed assets exist.")
EOF
  then
    fail "Manifest validation failed (see details above)."
  fi
else
  echo "[SKIP] python3 not found; skipping manifest validation."
fi

# 3. Content invariants: lean-progressive-context domain keywords
echo "--- 3. Checking lean-progressive-context content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "token budget|budget envelope" "SKILL.md" "SKILL.md mandates a declared token budget"
check_grep "staged|expansion|one layer" "SKILL.md" "SKILL.md mandates staged expansion"
check_grep "sufficiency" "SKILL.md" "SKILL.md mandates sufficiency testing"
check_grep "capsule" "SKILL.md" "SKILL.md mandates Working Context Capsules"
check_grep "ledger" "SKILL.md" "SKILL.md mandates a running token ledger"
check_grep "hypothesis" "SKILL.md" "SKILL.md mandates hypothesis-driven expansion"
check_grep "capsule" "checks/lean-progressive-context-checklist.md" "Checklist covers capsules"
check_grep "ledger|reconcile" "checks/lean-progressive-context-checklist.md" "Checklist covers ledger reconciliation"
check_grep "capsule|sufficiency|expansion" "references/lean-progressive-context-guide.md" "Reference guide covers capsules, sufficiency, and expansion"
check_grep "ledger|running total" "templates/lean-progressive-context-spec.md" "Spec template carries an expansion ledger"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] lean-progressive-context verification passed cleanly."
  exit 0
else
  echo "[ERROR] lean-progressive-context verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
