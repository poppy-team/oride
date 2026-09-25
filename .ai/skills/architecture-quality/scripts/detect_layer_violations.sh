#!/usr/bin/env sh
set -e
TARGET="."
echo "Auditing layer boundary integrity (Clean Architecture)..."
grep -rnE "(import|require).*(\/adapters\/|\/controllers\/|\/infra\/)" "/domain" 2>/dev/null && {
    echo "ERROR: Clean Architecture boundary violation: Domain imports external adapter/infra!"
    exit 1
} || {
    echo "Architecture boundaries compliant: Domain layer is self-contained."
    exit 0
}
