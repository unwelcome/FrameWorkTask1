#!/bin/bash
find . -name "go.mod" -not -path "*/vendor/*" | while read f; do
    dir=$(dirname "$f")
    echo "→ tidy in $dir"
    (cd "$dir" && go mod tidy)
done