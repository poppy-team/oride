#!/usr/bin/env bash
# Verification script for Plugin Architecture.
# Run from the repository root: bash src/prumo/resources/workforce/skills/plugin-architecture/scripts/verify.sh
set -euo pipefail

SKILL_DIR="src/prumo/resources/workforce/skills/plugin-architecture"

echo "===================================================="
echo " [Prumo Skill: plugin-architecture] Plugin Architecture Audit"
echo "===================================================="

VIOLATIONS=0
fail() { echo "[FAIL] $1"; VIOLATIONS=$((VIOLATIONS + 1)); }
pass() { echo "[PASS] $1"; }

# 1. Required skill assets exist and are non-empty
echo "--- 1. Verifying required skill assets ---"
for f in "SKILL.md" "manifest.json" "checks/plugin-architecture-checklist.md" \
         "references/plugin-architecture-guide.md" \
         "templates/plugin-architecture-spec.md" "scripts/verify.sh"; do
  if [ -s "$SKILL_DIR/$f" ]; then
    pass "Found non-empty $f."
  else
    fail "Missing or empty $SKILL_DIR/$f."
  fi
done

# 2. Manifest validity and asset cross-references
echo "--- 2. Validating manifest and asset cross-references ---"
if command -v python3 >/dev/null 2>&1; then
  if ! python3 - "$SKILL_DIR" "plugin-architecture" <<'EOF'
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

# 3. Content invariants: sandbox, permissions, versioning, lifecycle
echo "--- 3. Checking plugin-architecture content invariants ---"
check_grep() {
  if grep -qiE "$1" "$SKILL_DIR/$2" 2>/dev/null; then
    pass "$3."
  else
    fail "$3."
  fi
}
check_grep "sandbox|WASM|isolated" "SKILL.md" "SKILL.md mandates sandboxed execution"
check_grep "permission|manifest|capabilit" "SKILL.md" "SKILL.md mandates permission manifests"
check_grep "apiVersion|semver|deprecat" "SKILL.md" "SKILL.md mandates versioned contracts"
check_grep "hot-reload|hot.reload|signature|Sigstore" "SKILL.md" "SKILL.md mandates lifecycle and signatures"
check_grep "quota|VFS|traversal" "checks/plugin-architecture-checklist.md" "Checklist covers quotas and traversal"
check_grep "compat|deprecat" "checks/plugin-architecture-checklist.md" "Checklist covers compatibility"
check_grep "ambient authority|capabilit" "references/plugin-architecture-guide.md" "Reference guide covers capability design"
check_grep "wasmtime|isolated-vm|ShadowRealm|seccomp" "references/plugin-architecture-guide.md" "Reference guide compares sandbox mechanisms"
check_grep "apiVersion|quota|compat" "templates/plugin-architecture-spec.md" "Plugin spec template records contract and quotas"
check_grep "hostile|fuzz" "templates/plugin-architecture-spec.md" "Plugin spec template records adversarial evidence"

echo "===================================================="
if [ "$VIOLATIONS" -eq 0 ]; then
  echo "[SUCCESS] plugin-architecture verification passed cleanly."
  exit 0
else
  echo "[ERROR] plugin-architecture verification failed with $VIOLATIONS violation(s)."
  exit 1
fi
