#!/usr/bin/env sh
# Real complexity and function size analyzer for Clean Code audits
set -eu

TARGET_DIR="${1:-.}"
echo "=== [clean-code] Running Complexity & Size Audit on $TARGET_DIR ==="

echo "[1/3] Scanning for excessively long functions (> 40 lines)..."
# Scan Go, Rust, Python, and TS files
python3 -c "
import os, sys

target = sys.argv[1]
long_funcs = []

for root, _, files in os.walk(target):
    if any(p in root for p in ['.git', 'node_modules', 'vendor', 'target']):
        continue
    for f in files:
        if f.endswith(('.go', '.rs', '.py', '.ts', '.js')):
            path = os.path.join(root, f)
            try:
                with open(path, 'r', errors='ignore') as fp:
                    lines = fp.readlines()
                in_func = False
                func_name = ''
                func_lines = 0
                for idx, line in enumerate(lines):
                    sline = line.strip()
                    if sline.startswith(('func ', 'fn ', 'def ', 'function ')):
                        if in_func and func_lines > 40:
                            long_funcs.append((path, func_name, func_lines))
                        in_func = True
                        func_name = sline[:40]
                        func_lines = 0
                    elif in_func:
                        func_lines += 1
                        if sline == '}' and not line.startswith(' '):
                            if func_lines > 40:
                                long_funcs.append((path, func_name, func_lines))
                            in_func = False
            except Exception:
                pass

if long_funcs:
    print(f'Found {len(long_funcs)} functions exceeding 40 lines:')
    for p, name, count in long_funcs[:10]:
        print(f'  {p} -> {name}... ({count} lines)')
else:
    print('Clean: No oversized functions detected.')
" "$TARGET_DIR"

echo "[2/3] Scanning for deeply nested blocks (>= 4 indentation levels)..."
grep -rnE "^(    ){4,}|^\t{4,}" --include="*.go" --include="*.rs" --include="*.ts" --include="*.py" "$TARGET_DIR" 2>/dev/null | grep -v "/vendor/" | grep -v "node_modules" | head -n 10 || echo "Indentation nesting: Clean."

echo "[3/3] Checking for commented-out dead code..."
grep -rnE "(^\s*//\s*(var|let|const|func|def|import|if|for)\s+)|(^\s*#\s*(def|import|class|if)\s+)" "$TARGET_DIR" 2>/dev/null | grep -v "test" | head -n 10 || echo "Dead code scan: Clean."

echo "=== [clean-code] Complexity Audit Complete ==="
exit 0
