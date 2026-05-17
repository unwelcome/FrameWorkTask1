#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/.."

run_service() {
  local service=$1
  local dir="$BACKEND_DIR/$service"

  if [ ! -d "$dir" ]; then
    echo "ERROR: service '$service' not found at $dir"
    exit 1
  fi

  echo ""
  echo "=== Unit tests: $service ==="
  cd "$dir"
  go test -v ./internal/services/...
  echo "=== $service: OK ==="
}

ALL_SERVICES=("auth" "company" "application")

if [ -n "$1" ]; then
  run_service "$1"
else
  for service in "${ALL_SERVICES[@]}"; do
    run_service "$service"
  done
  echo ""
  echo "=== All unit tests passed ==="
fi
