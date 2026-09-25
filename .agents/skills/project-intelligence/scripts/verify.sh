#!/usr/bin/env bash
# Verification script for Project Intelligence.
# Run from the repository root: bash src/prumo/resources/workforce/skills/project-intelligence/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/project-intelligence"

echo "===================================================="
echo " [Prumo Skill: project-intelligence] Project Intelligence Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/project-intelligence-checklist.md" \
         "references/project-intelligence-guide.md" \
         "templates/project-intelligence-spec.md" "scripts/verify.sh"; do
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

if manifest.get("id") != "project-intelligence":
    errors.append("manifest id must be 'project-intelligence'")
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

# 3. Content invariants: project-intelligence domain keywords
echo "--- 3. Checking project-intelligence content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "intelligence ledger|ledger" "SKILL.md" "SKILL.md mandates an intelligence ledger"
check_grep "variance" "SKILL.md" "SKILL.md mandates variance computation"
check_grep "calibrat" "SKILL.md" "SKILL.md mandates estimate calibration"
check_grep "estimate.*before|before.*estimate|precede" "SKILL.md" "SKILL.md requires estimates before work"
check_grep "reopen" "SKILL.md" "SKILL.md tracks reopen quality signals"
check_grep "tasks, not people|never individuals|individual attribution" "SKILL.md" "SKILL.md forbids individual attribution"
check_grep "backfill|timestamped" "checks/project-intelligence-checklist.md" "Checklist guards estimate integrity"
check_grep "variance|calibration factor" "references/project-intelligence-guide.md" "Reference guide covers variance and calibration"
check_grep "variance|calibration" "templates/project-intelligence-spec.md" "Spec template carries variance and calibration tables"
check_grep "jsonl|schema" "templates/project-intelligence-spec.md" "Spec template references the JSONL ledger schema"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] project-intelligence verification passed cleanly."
  exit 0
else
  echo "[ERROR] project-intelligence verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
