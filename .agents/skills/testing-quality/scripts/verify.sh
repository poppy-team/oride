#!/usr/bin/env bash
set -euo pipefail

SKILL_ID="testing-quality"
SKILL_DIR="src/prumo/resources/workforce/skills/$SKILL_ID"
VIOLATIONS=0

pass() { printf '[PASS] %s\n' "$1"; }
fail() { printf '[FAIL] %s\n' "$1"; VIOLATIONS=$((VIOLATIONS + 1)); }

for asset in SKILL.md manifest.json; do
  if [[ -s "$SKILL_DIR/$asset" ]]; then pass "Found non-empty $asset."; else fail "Missing or empty $asset."; fi
done

if command -v python3 >/dev/null 2>&1; then
  if python3 - "$SKILL_DIR" "$SKILL_ID" <<'PY'
from pathlib import Path
import json
import sys

skill_dir = Path(sys.argv[1])
skill_id = sys.argv[2]
errors = []
try:
    manifest = json.loads((skill_dir / "manifest.json").read_text(encoding="utf-8"))
except Exception as exc:
    print(f"[FAIL] Manifest JSON: {exc}")
    raise SystemExit(1)
if manifest.get("id") != skill_id:
    errors.append("manifest id does not match directory")
if not isinstance(manifest.get("version"), int) or manifest.get("version", 0) < 2:
    errors.append("manifest version must be at least 2")
if not isinstance(manifest.get("schema_version"), int) or manifest.get("schema_version", 0) < 3:
    errors.append("manifest schema_version must be at least 3")
valid_modes = {"implementation", "review", "audit", "research", "design", "testing", "documentation", "release"}
modes = manifest.get("modes")
if not isinstance(modes, list) or not modes:
    errors.append("manifest modes must be a non-empty list")
else:
    for mode in modes:
        if mode not in valid_modes:
            errors.append(f"invalid mode: {mode}")
for key in ("references", "templates", "checks", "scripts"):
    paths = manifest.get(key)
    if not isinstance(paths, list) or not paths:
        errors.append(f"manifest {key} must be a non-empty list")
        continue
    for rel in paths:
        path = Path(rel) if isinstance(rel, str) else Path(".")
        if not isinstance(rel, str) or path.is_absolute() or ".." in path.parts:
            errors.append(f"unsafe {key} path: {rel}")
            continue
        asset = skill_dir / path
        if not asset.is_file() or asset.stat().st_size == 0:
            errors.append(f"missing or empty {key} asset: {rel}")
for rel in manifest.get("examples", []):
    path = Path(rel) if isinstance(rel, str) else Path(".")
    if not isinstance(rel, str) or path.is_absolute() or ".." in path.parts:
        errors.append(f"unsafe examples path: {rel}")
    elif not (skill_dir / path).is_file() or (skill_dir / path).stat().st_size == 0:
        errors.append(f"missing or empty example asset: {rel}")
provenance = manifest.get("provenance", {})
if provenance.get("version") != "0.4.0":
    errors.append("provenance version must be 0.4.0")
if provenance.get("modified") != "2026-09-23":
    errors.append("provenance modified must be 2026-09-23")
if errors:
    for error in errors:
        print(f"[FAIL] {error}")
    raise SystemExit(1)
print("[PASS] Manifest JSON and wired assets are valid.")
PY
  then pass "Manifest and asset validation completed."; else fail "Manifest or asset validation failed."; fi
else
  fail "python3 is required to validate the manifest."
fi

check_grep() {
  if grep -qiE -- "$1" "$SKILL_DIR/$2" 2>/dev/null; then pass "$3."; else fail "$3."; fi
}
check_grep "testing|test" "SKILL.md" "SKILL.md names testing quality"
check_grep "deterministic|isolat" "SKILL.md" "SKILL.md covers deterministic isolation"
check_grep "regression|mutation|coverage" "SKILL.md" "SKILL.md covers regression and quality signals"
check_grep "unit tests|integration|e2e" "references/testing-pyramid.md" "Reference guide explains the testing pyramid"
check_grep "- \[ \]" "checks/test-quality-checklist.md" "Checklist contains verification items"
check_grep "pyramid|regression case|evidence" "templates/testing-quality-template.md" "Template contains concrete evidence fields"

if [[ "$VIOLATIONS" -eq 0 ]]; then
  printf '[SUCCESS] %s verification passed cleanly.\n' "$SKILL_ID"
  exit 0
fi
printf '[ERROR] %s verification failed with %s violation(s).\n' "$SKILL_ID" "$VIOLATIONS"
exit 1
