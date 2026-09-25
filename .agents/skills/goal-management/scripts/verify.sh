#!/usr/bin/env bash
# Verification script for Prumo Goal Authoring, Phasing & Lock Management.
# Run from the repository root: bash src/prumo/resources/workforce/skills/goal-management/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/goal-management"

echo "===================================================="
echo " [Prumo Skill: goal-management] Goal Management Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/goal-management-checklist.md" \
         "references/goal-management-guide.md" \
         "templates/goal-management-spec.md" "scripts/verify.sh"; do
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

if manifest.get("id") != "goal-management":
    errors.append("manifest id must be 'goal-management'")
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

# 3. Content invariants: criteria, phases, locks, evidence
echo "--- 3. Checking goal-management content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "acceptance criteria" "SKILL.md" "SKILL.md mandates acceptance criteria"
check_grep "phase" "SKILL.md" "SKILL.md covers phased execution"
check_grep "lock" "SKILL.md" "SKILL.md mandates lock discipline"
check_grep "re-scope|rescope" "SKILL.md" "SKILL.md governs re-scope"
check_grep "predicate|measurable" "checks/goal-management-checklist.md" "Checklist covers criterion quality"
check_grep "lock" "checks/goal-management-checklist.md" "Checklist covers lock discipline"
check_grep "stale" "checks/goal-management-checklist.md" "Checklist covers staleness"
check_grep "predicate|stale" "references/goal-management-guide.md" "Reference guide models criteria and staleness"
check_grep "anti-pattern" "references/goal-management-guide.md" "Reference guide covers anti-patterns"
check_grep "evidence ledger|closure" "templates/goal-management-spec.md" "Goal template carries evidence ledger"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] goal-management verification passed cleanly."
  exit 0
else
  echo "[ERROR] goal-management verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
