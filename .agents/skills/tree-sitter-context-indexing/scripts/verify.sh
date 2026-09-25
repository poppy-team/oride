#!/usr/bin/env bash
# Verification script for Tree-Sitter Context Indexing.
# Run from the repository root: bash src/prumo/resources/workforce/skills/tree-sitter-context-indexing/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/tree-sitter-context-indexing"
SKILL_ID="tree-sitter-context-indexing"

echo "===================================================="
echo " [Prumo Skill: $SKILL_ID] Tree-Sitter Indexing Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/tree-sitter-context-indexing-checklist.md" \
         "references/tree-sitter-context-indexing-guide.md" \
         "templates/tree-sitter-context-indexing-spec.md" "scripts/verify.sh"; do
  if [ -s "$SKILL_DIR/$f" ]; then
    pass "Found non-empty $f."
  else
    fail "Missing or empty $SKILL_DIR/$f."
  fi
done
if [ -x "$SKILL_DIR/scripts/verify.sh" ]; then
  pass "verify.sh is executable."
else
  fail "verify.sh is not executable."
fi

# 2. Stale superseded artifacts must be gone
echo "--- 2. Checking for stale superseded artifacts ---"
for f in "templates/tree-sitter-context-indexing-template.md"; do
  if [ -e "$SKILL_DIR/$f" ]; then
    fail "Stale artifact still present: $SKILL_DIR/$f."
  else
    pass "Stale artifact absent: $f."
  fi
done

# 3. Manifest validity and asset cross-references
echo "--- 3. Validating manifest and asset wiring ---"
if command -v python3 >/dev/null 2>&1; then
  if ! python3 - "$SKILL_DIR" "$SKILL_ID" <<'EOF'
import json
import os
import sys

skill_dir, skill_id = sys.argv[1], sys.argv[2]
errors = []

with open(os.path.join(skill_dir, "manifest.json"), encoding="utf-8") as fh:
    manifest = json.load(fh)

if manifest.get("id") != skill_id:
    errors.append(f"manifest id must be {skill_id!r}")
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
print("[PASS] Manifest is valid and all wired assets exist.")
EOF
  then
    fail "Manifest validation failed (see details above)."
  fi
else
  echo "[SKIP] python3 not found; skipping manifest validation."
fi

# 4. Content invariants for tree-sitter context indexing
echo "--- 4. Checking tree-sitter-context-indexing content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "grammar.*pin|pinned.*grammar" "SKILL.md" "SKILL.md mandates pinned grammar revisions"
check_grep "\.scm|declarative quer" "SKILL.md" "SKILL.md mandates declarative .scm queries"
check_grep "old_tree|old tree|incremental" "SKILL.md" "SKILL.md mandates incremental re-parsing"
check_grep "chunk|byte" "SKILL.md" "SKILL.md mandates syntax-aligned chunking"
check_grep "ERROR node|error flag" "SKILL.md" "SKILL.md mandates parse-error flagging"
check_grep "captures|@function|#match" "references/tree-sitter-context-indexing-guide.md" "Reference guide covers captures and predicates"
check_grep "throughput|MB per second|recall" "templates/tree-sitter-context-indexing-spec.md" "Spec template records throughput and recall"
check_grep "p95|recall" "checks/tree-sitter-context-indexing-checklist.md" "Checklist covers recall and latency gates"
check_grep "tree-sitter (parse|query|CLI|cli)" "references/tree-sitter-context-indexing-guide.md" "Reference guide covers CLI workflow"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] $SKILL_ID verification passed cleanly."
  exit 0
else
  echo "[ERROR] $SKILL_ID verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
