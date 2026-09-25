#!/usr/bin/env sh
# Verification script for Native Memory Management
set -eu

echo "=== [memory-management] Starting Memory Architecture Audit ==="

# 1. Check for raw naked allocations in high-frequency loops
echo "[1/3] Scanning for naked dynamic allocations in source code..."
MALLOC_CALLS=$(grep -rn "malloc(" src/ 2>/dev/null | grep -v "allocator" | grep -v "test" || true)
if [ -n "$MALLOC_CALLS" ]; then
    echo "WARNING: Raw malloc calls detected outside dedicated allocator modules:"
    echo "$MALLOC_CALLS"
    echo "Review and migrate to Arena or Pool allocation."
else
    echo "Naked malloc scan: Clean."
fi

# 2. Check for false sharing risks in concurrent structures
echo "[2/3] Checking for multithreaded struct alignment..."
if grep -rn "pthread_create" src/ 2>/dev/null || grep -rn "std::thread" src/ 2>/dev/null; then
    if ! grep -rq "alignas(64)" src/ 2>/dev/null && !grep -rq "repr(align(64))" src/ 2>/dev/null; then
        echo "INFO: Multithreaded execution detected. Verify that thread-local stats structs are cache-line aligned (64 bytes) to avoid false sharing."
    fi
fi

# 3. Compile and run AddressSanitizer if C/C++ build system exists
echo "[3/3] Checking AddressSanitizer availability..."
if [ -f "CMakeLists.txt" ] || [ -f "Makefile" ]; then
    echo "INFO: Build system detected. Ensure ASan flags (-fsanitize=address,undefined) are active in debug/test builds."
fi

echo "=== [memory-management] Memory Audit Completed ==="
exit 0
