#!/usr/bin/env bash

# Script to find potentially long methods in Python files
# Usage: ./find_long_methods.sh <directory> <max_lines>

DIR=${1:-"."}
MAX_LINES=${2:-50}

echo "Scanning $DIR for functions longer than $MAX_LINES lines..."

find "$DIR" -name "*.py" -type f -print0 | while IFS= read -r -d '' file; do
    # Use awk to track function start and end
    awk -v max="$MAX_LINES" -v file="$file" '
        /^[ \t]*def / {
            if (in_func) {
                if (line_count > max) {
                    print file ": Function " func_name " is " line_count " lines long."
                }
            }
            in_func = 1
            func_name = $2
            sub(/\(.*/, "", func_name)
            line_count = 0
        }
        /^[ \t]*class / {
            if (in_func) {
                if (line_count > max) {
                    print file ": Function " func_name " is " line_count " lines long."
                }
            }
            in_func = 0
        }
        {
            if (in_func && $0 !~ /^[ \t]*#/ && $0 !~ /^[ \t]*$/) {
                line_count++
            }
        }
        END {
            if (in_func && line_count > max) {
                print file ": Function " func_name " is " line_count " lines long."
            }
        }
    ' "$file"
done
