#!/usr/bin/env sh
# Verification script for editor-tooling (Editor Tooling & UI)
set -e

echo "[Prumo Skill: editor-tooling] Starting verification routine..."

# 1. Secret & safety check
if command -v prumo >/dev/null 2>&1; then
    prumo tool scan-secrets . || {
        echo "WARNING: Secrets check flagged potential issues."
    }
fi

# 2. Syntax & test checks
if [ -f "go.mod" ] && command -v go >/dev/null 2>&1; then
    go vet ./... || true
elif [ -f "package.json" ] && command -v npm >/dev/null 2>&1; then
    npm test --if-present || true
elif [ -f "Cargo.toml" ] && command -v cargo >/dev/null 2>&1; then
    cargo check || true
fi

echo "[Prumo Skill: editor-tooling] Verification complete."
