#!/usr/bin/env bash
# Verification script for Native Performance Optimization.
# Run from the repository root: bash src/prumo/resources/workforce/skills/performance-native/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/performance-native"

echo "===================================================="
echo " [Prumo Skill: performance-native] Native Performance Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/performance-native-checklist.md" \
         "references/performance-native-guide.md" \
         "templates/performance-native-spec.md" "scripts/verify.sh"; do
  if [ -s "$SKILL_DIR/$f" ]; then
    pass "Found non-empty $f."
  else
    fail "Missing or empty $SKILL_DIR/$f."
  fi
done

# 2. Manifest validity and asset cross-references
echo "--- 2. Validating manifest and asset cross-references ---"
if command -v python3 >/dev/null 2>&1; then
  if ! python3 - "$SKILL_DIR" "performance-native" <<'EOF'
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
print("[PASS] Manifest is valid and all listed assets exist on disk.")
EOF
  then
    fail "Manifest validation failed (see details above)."
  fi
else
  echo "[SKIP] python3 not found; skipping manifest validation."
fi

# 3. Content invariants: profiling, cache behavior, allocation, GPU
echo "--- 3. Checking performance-native content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "profiler|perf record|tracy" "SKILL.md" "SKILL.md mandates profiler-driven work"
check_grep "cache.miss|L1-D" "SKILL.md" "SKILL.md mandates cache-miss analysis"
check_grep "arena|allocation churn|malloc" "SKILL.md" "SKILL.md mandates allocation discipline"
check_grep "hitch|p99" "checks/performance-native-checklist.md" "Checklist covers p99 and hitch gates"
check_grep "heaptrack|sanitizer" "checks/performance-native-checklist.md" "Checklist covers memory verification"
check_grep "false sharing|alignas" "references/performance-native-guide.md" "Reference guide covers false sharing"
check_grep "SIMD|vectoriz" "references/performance-native-guide.md" "Reference guide covers vectorization"
check_grep "SoA|Top-Down|top-down" "references/performance-native-guide.md" "Reference guide covers layout and top-down analysis"
check_grep "RenderDoc|draw call" "templates/performance-native-spec.md" "Perf spec template covers GPU evidence"
check_grep "p99|RSS" "templates/performance-native-spec.md" "Perf spec template records p99 and RSS"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] performance-native verification passed cleanly."
  exit 0
else
  echo "[ERROR] performance-native verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
