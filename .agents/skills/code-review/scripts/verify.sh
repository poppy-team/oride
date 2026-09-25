#!/usr/bin/env bash
# Verification script for Independent Evidence-Based Code Review.
# Run from the repository root: bash src/prumo/resources/workforce/skills/code-review/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/code-review"

echo "===================================================="
echo " [Prumo Skill: code-review] Review Contract Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/code-review-checklist.md" \
         "references/code-review-guide.md" \
         "templates/code-review-spec.md" "scripts/verify.sh"; do
  if [ -s "$SKILL_DIR/$f" ]; then
    pass "Found non-empty $f."
  else
    fail "Missing or empty $SKILL_DIR/$f."
  fi
done

# 2. Stale superseded artifacts must be gone
echo "--- 2. Checking for stale superseded artifacts ---"
for f in "templates/code-review-template.md"; do
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

if manifest.get("id") != "code-review":
    errors.append("manifest id must be 'code-review'")
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
echo "--- 4. Checking review content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "blocker" "SKILL.md" "SKILL.md grades findings by severity"
check_grep "APPROVE|REQUEST CHANGES|ESCALATE" "SKILL.md" "SKILL.md defines deterministic verdicts"
check_grep "re-verif" "SKILL.md" "SKILL.md mandates fix re-verification"
check_grep "file:line" "SKILL.md" "SKILL.md requires file:line evidence"
check_grep "independence|independent" "SKILL.md" "SKILL.md requires reviewer independence"
check_grep "Silent-Failure" "checks/code-review-checklist.md" "Checklist covers silent failures"
check_grep "severity" "references/code-review-guide.md" "Reference guide defines severity model"
check_grep "Verdict" "templates/code-review-spec.md" "Review template records a verdict"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] code-review verification passed cleanly."
  exit 0
else
  echo "[ERROR] code-review verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
