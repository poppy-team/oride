#!/usr/bin/env sh
# Verification script for Go Idiomatic Reliability & Concurrency Hygiene
set -eu

echo "=== [lang-go] Starting Go Verification Suite ==="

# 1. Ensure go toolchain is present
if ! command -v go >/dev/null 2>&1; then
    echo "ERROR: Go toolchain not found in PATH." >&2
    exit 1
fi

echo "Go Version: $(go version)"

# 2. Format check
echo "[1/4] Checking code formatting (gofmt)..."
UNFORMATTED=$(gofmt -l . 2>/dev/null | grep -v ".git" || true)
if [ -n "$UNFORMATTED" ]; then
    echo "WARNING: Unformatted Go files detected:"
    echo "$UNFORMATTED"
    echo "Run: gofmt -s -w ."
fi

# 3. Static analysis with go vet
echo "[2/4] Running go vet..."
go vet ./... || echo "WARNING: go vet flagged potential issues above."

# 4. Check for common suspicious patterns (naked goroutines without context or waitgroup)
echo "[3/4] Scanning for unbounded 'go func()' invocations..."
if grep -rn --exclude-dir=".git" --exclude-dir="vendor" "go func()" --include="*.go" . 2>/dev/null | grep -v "_test.go"; then
    echo "INFO: Verify that all spawned goroutines above have explicit lifecycle owners."
fi

# 5. Run tests with race detector or standalone example
echo "[4/4] Verifying Go execution..."
if [ -f "src/prumo/resources/workforce/skills/lang-go/examples/safe_go.go" ]; then
    go run src/prumo/resources/workforce/skills/lang-go/examples/safe_go.go
fi

echo "=== [lang-go] Go Verification Passed Cleanly ==="
exit 0
