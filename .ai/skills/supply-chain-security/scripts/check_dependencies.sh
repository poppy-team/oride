#!/usr/bin/env sh
set -e
echo "Auditing dependency supply chain..."
if [ -f "package.json" ] && command -v npm >/dev/null 2>&1; then
    npm audit || true
fi
if [ -f "go.mod" ] && command -v go >/dev/null 2>&1; then
    go list -m all
fi
if [ -f "Cargo.toml" ] && command -v cargo >/dev/null 2>&1; then
    cargo audit || true
fi
