#!/usr/bin/env bash
# Verification script for Release Engineering, Packaging & Rollback Readiness.
# Run from the repository root: bash src/prumo/resources/workforce/skills/release-engineering/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/release-engineering"

echo "===================================================="
echo " [Prumo Skill: release-engineering] Release Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/release-engineering-checklist.md" \
         "references/release-engineering-guide.md" \
         "templates/release-engineering-spec.md" "scripts/verify.sh"; do
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

if manifest.get("id") != "release-engineering":
    errors.append("manifest id must be 'release-engineering'")
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

# 3. Content invariants: versioning, signing, migration, rollback
echo "--- 3. Checking release-engineering content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "semver" "SKILL.md" "SKILL.md mandates SemVer versioning"
check_grep "checksum|sha256" "SKILL.md" "SKILL.md mandates checksum verification"
check_grep "migration" "SKILL.md" "SKILL.md covers migration rehearsal"
check_grep "rollback" "SKILL.md" "SKILL.md mandates rollback readiness"
check_grep "canary|staging" "checks/release-engineering-checklist.md" "Checklist covers staged rollout"
check_grep "downgrade" "checks/release-engineering-checklist.md" "Checklist covers downgrade rehearsal"
check_grep "halt" "checks/release-engineering-checklist.md" "Checklist covers halt conditions"
check_grep "expand.*contract|canary" "references/release-engineering-guide.md" "Reference guide covers rollout models"
check_grep "anti-pattern" "references/release-engineering-guide.md" "Reference guide covers anti-patterns"
check_grep "rollback|SHA256SUMS" "templates/release-engineering-spec.md" "Release template carries integrity and rollback"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] release-engineering verification passed cleanly."
  exit 0
else
  echo "[ERROR] release-engineering verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
