#!/usr/bin/env sh
# Verification script for project-documentation-architect
set -e

echo "[Prumo Skill: project-documentation-architect] Starting verification routine..."

SKILL_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

# 1. Secret & safety check
if command -v prumo >/dev/null 2>&1; then
    prumo tool scan-secrets "$SKILL_DIR" || {
        echo "WARNING: Secrets check flagged potential issues."
    }
fi

# 2. Manifest is valid JSON with required fields
export SKILL_DIR
if command -v python3 >/dev/null 2>&1; then
    python3 - "$SKILL_DIR/manifest.json" <<'EOF'
import json, re, sys
m = json.load(open(sys.argv[1], encoding="utf-8"))
assert re.match(r"^[A-Za-z0-9._-]+$", m["id"]), "bad id"
assert m.get("name") and m.get("purpose"), "missing name/purpose"
for rel in m.get("references", []) + m.get("templates", []) + m.get("checks", []) + m.get("scripts", []) + m.get("examples", []):
    import os
    assert os.path.isfile(os.path.join(os.environ.get("SKILL_DIR", "."), rel)), "missing: " + rel
print("manifest OK:", m["id"])
EOF
else
    echo "NOTE: python3 not available; skipped manifest validation."
fi

# 3. SKILL.md carries the required house sections
for section in "## Purpose" "## Use when" "## Do not use when" "## Required context" "## Procedure" "## Decision rules" "## Evidence required" "## Output contract"; do
    grep -qF "$section" "$SKILL_DIR/SKILL.md" || {
        echo "FAIL: SKILL.md missing section: $section"
        exit 1
    }
done

echo "[Prumo Skill: project-documentation-architect] Verification complete."
