#!/usr/bin/env bash
# Verification script for CI/CD Pipeline Design, Hardening & Reproducibility.
# Run from the repository root: bash src/prumo/resources/workforce/skills/ci-cd/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/ci-cd"

echo "===================================================="
echo " [Prumo Skill: ci-cd] CI/CD Pipeline Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/ci-cd-checklist.md" \
         "references/ci-cd-guide.md" \
         "templates/ci-cd-spec.md" "scripts/verify.sh"; do
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

if manifest.get("id") != "ci-cd":
    errors.append("manifest id must be 'ci-cd'")
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

# 3. Content invariants: reproducibility, caching, secrets, gates
echo "--- 3. Checking ci-cd content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "clean checkout|reproducib" "SKILL.md" "SKILL.md mandates clean-checkout reproducibility"
check_grep "fail.?fast" "SKILL.md" "SKILL.md mandates fail-fast gate ordering"
check_grep "cache" "SKILL.md" "SKILL.md covers dependency caching"
check_grep "secret" "SKILL.md" "SKILL.md covers secret hardening"
check_grep "reproducib" "checks/ci-cd-checklist.md" "Checklist covers reproducibility"
check_grep "flake" "checks/ci-cd-checklist.md" "Checklist covers flake rate"
check_grep "secret" "checks/ci-cd-checklist.md" "Checklist covers secrets"
check_grep "oidc|provenance|pin" "references/ci-cd-guide.md" "Reference guide covers supply-chain hardening"
check_grep "anti-pattern" "references/ci-cd-guide.md" "Reference guide covers anti-patterns"
check_grep "stage|timing" "templates/ci-cd-spec.md" "Pipeline template carries stage timings"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] ci-cd verification passed cleanly."
  exit 0
else
  echo "[ERROR] ci-cd verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
