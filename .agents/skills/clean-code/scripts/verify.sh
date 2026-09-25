#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
MANIFEST_PATH="${SKILL_DIR}/manifest.json"

cd "${SKILL_DIR}"

VIOLATIONS=0

fail() {
    printf '[FAIL] %s\n' "$1" >&2
    VIOLATIONS=$((VIOLATIONS + 1))
}

required_assets=(
    "SKILL.md"
    "manifest.json"
    "references/clean-code-principles.md"
    "references/martin-fowler-refactoring.md"
    "templates/clean-code-review.md"
    "checks/cleanliness-checklist.md"
    "scripts/analyze_complexity.sh"
    "scripts/analyze-cohesion.sh"
    "scripts/verify.sh"
    "examples/before_after_clean_code.md"
)

for asset in "${required_assets[@]}"; do
    if [[ ! -s "${SKILL_DIR}/${asset}" ]]; then
        fail "Required asset is missing or empty: ${asset}"
    fi
done

MANIFEST_VIOLATIONS="$(
    python3 - "${MANIFEST_PATH}" "${SKILL_DIR}" <<'PY'
import json
import os
from pathlib import Path
import sys

manifest_path = Path(sys.argv[1])
skill_dir = Path(sys.argv[2]).resolve()
violations = []
valid_modes = {
    "implementation",
    "review",
    "audit",
    "research",
    "design",
    "testing",
    "release",
    "documentation",
}
asset_keys = ("references", "templates", "checks", "scripts", "examples")

try:
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
except (OSError, json.JSONDecodeError) as error:
    print(f"Manifest is not valid JSON: {error}")
    raise SystemExit(0)

if manifest.get("id") != "clean-code":
    violations.append("Manifest id must be clean-code")

version = manifest.get("version")
if type(version) is not int or version < 2:
    violations.append("Manifest version must be an integer >= 2")

schema_version = manifest.get("schema_version")
if type(schema_version) is not int or schema_version < 3:
    violations.append("Manifest schema_version must be an integer >= 3")

modes = manifest.get("modes")
if not isinstance(modes, list) or not modes:
    violations.append("Manifest modes must be a non-empty array")
else:
    invalid_modes = sorted({mode for mode in modes if mode not in valid_modes})
    if invalid_modes:
        violations.append(f"Manifest contains invalid modes: {', '.join(invalid_modes)}")

provenance = manifest.get("provenance")
if not isinstance(provenance, dict):
    violations.append("Manifest provenance must be an object")
else:
    if provenance.get("version") != "0.4.0":
        violations.append("Manifest provenance.version must be 0.4.0")
    if provenance.get("modified") != "2026-09-23":
        violations.append("Manifest provenance.modified must be 2026-09-23")

for key in asset_keys:
    assets = manifest.get(key)
    if not isinstance(assets, list) or not assets:
        violations.append(f"Manifest {key} must be a non-empty array")
        continue

    for asset in assets:
        if not isinstance(asset, str) or not asset.strip():
            violations.append(f"Manifest {key} contains an invalid asset path")
            continue

        relative_path = Path(asset)
        if relative_path.is_absolute() or ".." in relative_path.parts:
            violations.append(f"Manifest {key} asset escapes the skill directory: {asset}")
            continue

        resolved_path = (skill_dir / relative_path).resolve()
        try:
            resolved_path.relative_to(skill_dir)
        except ValueError:
            violations.append(f"Manifest {key} asset escapes the skill directory: {asset}")
            continue

        if not resolved_path.is_file() or resolved_path.stat().st_size == 0:
            violations.append(f"Manifest {key} asset is missing or empty: {asset}")
        elif key == "scripts" and not os.access(resolved_path, os.X_OK):
            violations.append(f"Manifest script is not executable: {asset}")

if "scripts/verify.sh" not in manifest.get("scripts", []):
    violations.append("Manifest scripts must include scripts/verify.sh")

for violation in violations:
    print(violation)
PY
)"

if [[ -n "${MANIFEST_VIOLATIONS}" ]]; then
    while IFS= read -r violation; do
        if [[ -n "${violation}" ]]; then
            fail "${violation}"
        fi
    done <<< "${MANIFEST_VIOLATIONS}"
fi

check_content() {
    local file="$1"
    local pattern="$2"
    local description="$3"

    if ! grep -qiE "${pattern}" "${SKILL_DIR}/${file}"; then
        fail "${description}"
    fi
}

check_content "SKILL.md" "domain[- ]expressive naming|domain naming" "SKILL.md must define domain-expressive naming"
check_content "SKILL.md" "single level of abstraction|SLAP" "SKILL.md must define SLAP"
check_content "SKILL.md" "rule of three" "SKILL.md must enforce the Rule of Three"
check_content "SKILL.md" "composition over inheritance" "SKILL.md must prefer composition over inheritance"
check_content "SKILL.md" "pure core, imperative shell" "SKILL.md must isolate pure logic from side effects"
check_content "SKILL.md" "command-query separation|CQS" "SKILL.md must define command-query separation"
check_content "SKILL.md" "cyclomatic complexity" "SKILL.md must define a complexity target"
check_content "SKILL.md" "preserve behavior" "SKILL.md must require behavior preservation"
check_content "checks/cleanliness-checklist.md" "no single-implementation interfaces" "Checklist must reject unnecessary interfaces"
check_content "references/clean-code-principles.md" "afferent.*coupling|efferent.*coupling" "Principles must cover coupling metrics"
check_content "references/martin-fowler-refactoring.md" "extract function / method" "Reference must cover function extraction"

if (( VIOLATIONS > 0 )); then
    printf '[ERROR] clean-code verification failed with %d violation(s).\n' "${VIOLATIONS}" >&2
    exit 1
fi

printf '[Prumo Skill: clean-code] Verification completed successfully.\n'
