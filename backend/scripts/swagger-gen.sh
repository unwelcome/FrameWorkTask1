#!/bin/bash
set -e

# Determine backend root directory (parent of this script's folder)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
cd "$BACKEND_DIR/gateway"

echo "Working directory: $(pwd)"
echo ""
echo "→ Generating swagger docs for gateway..."

swag init \
  -o ./api/docs \
  --dir ./cmd,./internal/entities,./internal/errors,./internal/handlers

echo "✓ Swagger docs generated in gateway/api/docs"
