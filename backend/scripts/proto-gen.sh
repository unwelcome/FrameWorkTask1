#!/bin/bash
set -e

# Determine backend root directory (parent of this script's folder)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
cd "$BACKEND_DIR"

echo "Working directory: $BACKEND_DIR"
echo ""

gen_auth() {
    echo "→ Generating protobuf code for auth service..."
    protoc \
      --proto_path=auth/api \
      --go_out=auth/api/generated \
      --go_opt=paths=source_relative \
      --go-grpc_out=auth/api/generated \
      --go-grpc_opt=paths=source_relative \
      auth.proto
    echo "✓ Auth service done"
}

gen_company() {
    echo "→ Generating protobuf code for company service..."
    protoc \
      --proto_path=company/api \
      --go_out=company/api/generated \
      --go_opt=paths=source_relative \
      --go-grpc_out=company/api/generated \
      --go-grpc_opt=paths=source_relative \
      company.proto
    echo "✓ Company service done"
}

gen_application() {
    echo "→ Generating protobuf code for application service..."
    protoc \
      --proto_path=application/api \
      --go_out=application/api/generated \
      --go_opt=paths=source_relative \
      --go-grpc_out=application/api/generated \
      --go-grpc_opt=paths=source_relative \
      application.proto
    echo "✓ Application service done"
}

# If argument provided — generate only that service
case "$1" in
    auth)        gen_auth ;;
    company)     gen_company ;;
    application) gen_application ;;
    "")
        gen_auth
        gen_company
        gen_application
        ;;
    *)
        echo "Unknown service: $1"
        echo "Usage: $0 [auth|company|application]"
        exit 1
        ;;
esac

echo ""
echo "Done"
