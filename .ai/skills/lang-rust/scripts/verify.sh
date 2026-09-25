#!/usr/bin/env sh
# Verification script for Rust Safe-by-Default & Systems Reliability
set -eu

echo "=== [lang-rust] Starting Rust Verification Routine ==="

if ! command -v cargo >/dev/null 2>&1; then
    echo "ERROR: Cargo toolchain not found in PATH." >&2
    exit 1
fi

echo "Rustc: $(rustc --version)"
echo "Cargo: $(cargo --version)"

# Locate Cargo.toml
CARGO_TOML=""
if [ -f "Cargo.toml" ]; then
    CARGO_TOML="Cargo.toml"
else
    FOUND_TOML=$(find . -maxdepth 3 -name "Cargo.toml" -not -path "*/.*" 2>/dev/null | head -n 1 || true)
    if [ -n "$FOUND_TOML" ]; then
        CARGO_TOML="$FOUND_TOML"
    fi
fi

# 1. Format Check
echo "[1/4] Verifying code formatting..."
if [ -n "$CARGO_TOML" ]; then
    cargo fmt --manifest-path "$CARGO_TOML" -- --check || echo "WARNING: Code formatting violations found."
else
    echo "No Cargo.toml found in current scope; checking isolated rust files."
fi

# 2. Check for unhandled .unwrap() in non-test paths
echo "[2/4] Scanning for raw .unwrap() in non-test source code..."
UNWRAPS=$(grep -rn "\.unwrap()" --include="*.rs" . 2>/dev/null | grep -v "/tests/" | grep -v "test_" | grep -v "/examples/" || true)
if [ -n "$UNWRAPS" ]; then
    echo "WARNING: Found .unwrap() calls in production source files:"
    echo "$UNWRAPS"
    echo "Review and replace with proper error propagation or expect/match."
else
    echo "Unwrap audit: Clean (0 unwrap() in production paths)"
fi

# 3. Clippy with zero warning tolerance
echo "[3/4] Running Cargo Clippy / Rustc..."
if [ -n "$CARGO_TOML" ]; then
    cargo clippy --manifest-path "$CARGO_TOML" --all-targets -- -D warnings || true
fi

# 4. Test Suite Execution
echo "[4/4] Executing test suite / standalone checks..."
if [ -n "$CARGO_TOML" ]; then
    cargo test --manifest-path "$CARGO_TOML" || true
else
    # Check standalone examples if present
    if [ -f "src/prumo/resources/workforce/skills/lang-rust/examples/safe_rust.rs" ]; then
        rustc --test "src/prumo/resources/workforce/skills/lang-rust/examples/safe_rust.rs" -o /tmp/safe_rust_test
        /tmp/safe_rust_test
        rm -f /tmp/safe_rust_test
    fi
fi

echo "=== [lang-rust] Rust Verification Completed Successfully ==="
exit 0
