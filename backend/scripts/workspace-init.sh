#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/.."

cd "$BACKEND_DIR"

if [ -f go.work ]; then
  echo "go.work already exists, skipping"
  exit 0
fi

go work init
go work use ./application ./auth ./company ./contracts ./e2e ./gateway ./shared

echo "go.work created successfully"
