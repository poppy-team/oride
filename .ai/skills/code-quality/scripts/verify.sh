#!/usr/bin/env bash
# Verification script for Code Quality Auditing, Complexity & Maintainability.
# Run from the repository root: bash src/prumo/resources/workforce/skills/code-quality/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/code-quality"

echo "===================================================="
echo " [Prumo Skill: code-quality] Code Quality Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/code-quality-checklist.md" \
         "references/code-quality-guide.md" \
         "templates/code-quality-spec.md" "scripts/verify.sh"; do
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

if manifest.get("id") != "code-quality":
    errors.append("manifest id must be 'code-quality'")
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

# 3. Content invariants: complexity, duplication, and gate mechanics
echo "--- 3. Checking code-quality content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "cyclomatic" "SKILL.md" "SKILL.md enforces cyclomatic complexity ceilings"
check_grep "duplication" "SKILL.md" "SKILL.md enforces duplication thresholds"
check_grep "maintainability" "SKILL.md" "SKILL.md covers maintainability metrics"
check_grep "lint" "SKILL.md" "SKILL.md mandates lint gates"
check_grep "complexity" "checks/code-quality-checklist.md" "Checklist covers complexity ceilings"
check_grep "duplication" "checks/code-quality-checklist.md" "Checklist covers duplication invariants"
check_grep "coverage" "checks/code-quality-checklist.md" "Checklist covers test coverage"
check_grep "cyclomatic|maintainability index" "references/code-quality-guide.md" "Reference guide covers complexity models"
check_grep "anti-pattern" "references/code-quality-guide.md" "Reference guide covers anti-patterns"
check_grep "threshold|before.*after" "templates/code-quality-spec.md" "Audit template carries thresholds and before/after evidence"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] code-quality verification passed cleanly."
  exit 0
else
  echo "[ERROR] code-quality verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
