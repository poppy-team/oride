#!/usr/bin/env sh
# Real module coupling and import dependency auditor
set -eu

TARGET_DIR="${1:-.}"
echo "=== [clean-code] Running Coupling & Cohesion Analysis on $TARGET_DIR ==="

echo "Auditing external dependencies and import coupling..."
python3 -c "
import os, sys

target = sys.argv[1]
imports_per_file = []

for root, _, files in os.walk(target):
    if any(p in root for p in ['.git', 'node_modules', 'vendor', 'target']):
        continue
    for f in files:
        if f.endswith(('.go', '.rs', '.py', '.ts')):
            path = os.path.join(root, f)
            count = 0
            try:
                with open(path, 'r', errors='ignore') as fp:
                    for line in fp:
                        s = line.strip()
                        if s.startswith(('import ', 'from ', 'use ')):
                            count += 1
                if count > 15:
                    imports_per_file.append((path, count))
            except Exception:
                pass

imports_per_file.sort(key=lambda x: x[1], reverse=True)
if imports_per_file:
    print('Files with high coupling (> 15 imports):')
    for p, count in imports_per_file[:10]:
        print(f'  {p}: {count} dependencies')
else:
    print('Coupling audit: Clean (all modules have <= 15 dependencies).')
" "$TARGET_DIR"

echo "=== [clean-code] Coupling Analysis Complete ==="
exit 0
