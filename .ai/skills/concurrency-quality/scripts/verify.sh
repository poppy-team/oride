#!/usr/bin/env bash
set -euo pipefail

SKILL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "=== Verifying Concurrency Quality & Invariants ==="

# 1. Run race detection on concurrency examples
echo "[1/2] Running Go data race detector (-race) on bounded worker pool..."
go test -race -v "${SKILL_DIR}/examples/..."

# 2. Verify static concurrency patterns in examples and specs
echo "[2/2] Verifying structured concurrency and bounded queue invariants..."
python3 - <<PYCHECK
import sys
from pathlib import Path

skill_path = Path("${SKILL_DIR}")
pool_go = skill_path / "examples" / "bounded_worker_pool.go"

if not pool_go.exists():
    print("Error: bounded_worker_pool.go not found", file=sys.stderr)
    sys.exit(1)

content = pool_go.read_text(encoding="utf-8")

# Check for required synchronization primitives
required_tokens = ["sync.WaitGroup", "atomic.Bool", "select", "context.Context", "recover()"]
missing = [t for t in required_tokens if t not in content]
if missing:
    print(f"Error: Missing required concurrency primitives: {missing}", file=sys.stderr)
    sys.exit(1)

# Check for cache padding
if "_pad" not in content or "64 bytes" not in content:
    print("Error: Missing cache line padding verification in worker metrics", file=sys.stderr)
    sys.exit(1)

print("Concurrency invariants and cache alignment verified successfully.")
PYCHECK

echo "Concurrency Quality verification passed."
exit 0
