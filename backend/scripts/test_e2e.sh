#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/.."

# Services that have e2e tests
E2E_SERVICES=("auth")

run_e2e() {
  local service=$1
  local dir="$BACKEND_DIR/e2e"

  # Check that this service actually has e2e tests
  local supported=false
  for s in "${E2E_SERVICES[@]}"; do
    [ "$s" = "$service" ] && supported=true && break
  done

  if [ "$supported" = false ]; then
    echo "ERROR: no e2e tests for service '$service' (available: ${E2E_SERVICES[*]})"
    exit 1
  fi

  echo ""
  echo "=== E2E tests: $service ==="
  cd "$dir"
  go test -v -timeout 10m -run "$(get_pattern "$service")"
  echo "=== $service e2e: OK ==="
}

get_pattern() {
  case "$1" in
    auth) echo "Test" ;;  # all tests in e2e package are auth tests for now
    *)    echo "Test" ;;
  esac
}

if [ -n "$1" ]; then
  run_e2e "$1"
else
  for service in "${E2E_SERVICES[@]}"; do
    run_e2e "$service"
  done
  echo ""
  echo "=== All e2e tests passed ==="
fi
