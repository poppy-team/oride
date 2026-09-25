#!/usr/bin/env sh
set -e
echo "Checking for cyclic imports and package boundaries..."
if command -v go >/dev/null 2>&1 && [ -f "go.mod" ]; then
    exec go vet ./...
elif command -v npm >/dev/null 2>&1 && [ -f "package.json" ]; then
    exec npx madge --circular .
elif command -v cargo >/dev/null 2>&1 && [ -f "Cargo.toml" ]; then
    exec cargo check
else
    echo "No standard package manifest detected; scan complete."
fi
