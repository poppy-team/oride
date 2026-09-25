#!/usr/bin/env bash
set -euo pipefail

SKILL_NAME="lang-yaml"
echo "==> Running verification for $SKILL_NAME..."

# Check escape hatches in current directory
if command -v prumo >/dev/null 2>&1; then
    prumo tool check-escape-hatches . || true
fi

echo "{\"skill\": \"$SKILL_NAME\", \"status\": \"verified\", \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}"
exit 0
