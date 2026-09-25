#!/usr/bin/env bash
# Verification script for Git Workflow, Branching & Evidence-Backed Review.
# Run from the repository root: bash src/prumo/resources/workforce/skills/git-workflow/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/git-workflow"

echo "===================================================="
echo " [Prumo Skill: git-workflow] Git Workflow Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/git-workflow-checklist.md" \
         "references/git-workflow-guide.md" \
         "templates/git-workflow-spec.md" "scripts/verify.sh"; do
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

if manifest.get("id") != "git-workflow":
    errors.append("manifest id must be 'git-workflow'")
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

# 3. Content invariants: branches, commits, review evidence
echo "--- 3. Checking git-workflow content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "worktree|branch" "SKILL.md" "SKILL.md covers branch and worktree isolation"
check_grep "conventional commit" "SKILL.md" "SKILL.md mandates conventional commits"
check_grep "pull request" "SKILL.md" "SKILL.md covers evidence-backed pull requests"
check_grep "force-with-lease|force.push" "SKILL.md" "SKILL.md rules force-push scope"
check_grep "conventional" "checks/git-workflow-checklist.md" "Checklist covers commit format"
check_grep "CODEOWNERS|approv" "checks/git-workflow-checklist.md" "Checklist covers review approvals"
check_grep "secret" "checks/git-workflow-checklist.md" "Checklist covers secret hygiene"
check_grep "bisect|squash|rebase" "references/git-workflow-guide.md" "Reference guide covers history models"
check_grep "anti-pattern" "references/git-workflow-guide.md" "Reference guide covers anti-patterns"
check_grep "acceptance mapping|rollback" "templates/git-workflow-spec.md" "PR template carries acceptance mapping"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] git-workflow verification passed cleanly."
  exit 0
else
  echo "[ERROR] git-workflow verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
