#!/usr/bin/env bash
# Verification script for interaction-research (Interaction & User Experience Research)
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
SKILL_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
cd "${SKILL_DIR}"

echo "[Prumo Skill: interaction-research] Starting verification routine..."

# 1. Audit quantitative usability metrics in study example
echo "[1/3] Auditing quantitative research metrics (SUS, SEQ, Completion Rate)..."
python3 -c "
import os

study_path = 'examples/usability_test_study.md'
if not os.path.exists(study_path):
    raise FileNotFoundError('Missing ' + study_path)

with open(study_path, 'r', encoding='utf-8') as f:
    text = f.read()

required_metrics = ['SUS Score', 'SEQ', 'Completion', 'Finding']
for m in required_metrics:
    if m not in text:
        raise ValueError('Usability study example missing metric: ' + m)

print('Usability study metrics and findings validated.')
"

# 2. Mathematical validation of SUS calculation formula
echo "[2/3] Verifying SUS mathematical scoring calculation algorithm..."
python3 -c "
# Standard test array of 10 ratings (scale 1 to 5)
sample_responses = [4, 2, 4, 1, 5, 2, 4, 1, 5, 2] # Mostly positive experience

def calculate_sus(responses):
    assert len(responses) == 10
    total = 0
    for idx, r in enumerate(responses):
        if (idx + 1) % 2 != 0: # Odd
            total += (r - 1)
        else: # Even
            total += (5 - r)
    return total * 2.5

score = calculate_sus(sample_responses)
assert 0 <= score <= 100
assert score == 85.0, f'Expected 85.0, got {score}'
print('SUS calculation formula verified mathematically.')
"

# 3. Template verification
echo "[3/3] Checking template integrity..."
if [ ! -f "templates/user-research-protocol.md" ]; then
    echo "ERROR: Missing templates/user-research-protocol.md"
    exit 1
fi

echo "[Prumo Skill: interaction-research] Verification completed successfully."
exit 0
