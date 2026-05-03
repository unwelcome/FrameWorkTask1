#!/bin/bash
set -e

# Determine backend root directory (parent of this script's folder)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
cd "$BACKEND_DIR"

echo "Working directory: $BACKEND_DIR"
echo ""

find . -name "go.mod" -not -path "*/vendor/*" | while read f; do
    dir=$(dirname "$f")
    echo "→ tidy in $dir"
    (cd "$dir" && go mod tidy)
done

echo ""
echo "All modules tidied"
