#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"
CONTRACTS_DIR="$BACKEND_DIR/contracts"
cd "$BACKEND_DIR"

echo "Working directory: $BACKEND_DIR"
echo ""

gen_auth() {
    echo "→ Generating protobuf code for auth..."
    protoc \
      --proto_path="$CONTRACTS_DIR/auth" \
      --go_out="$CONTRACTS_DIR/auth/generated" \
      --go_opt=paths=source_relative \
      --go-grpc_out="$CONTRACTS_DIR/auth/generated" \
      --go-grpc_opt=paths=source_relative \
      "$CONTRACTS_DIR/auth/auth.proto"
    echo "✓ Auth done"
}

gen_company() {
    echo "→ Generating protobuf code for company..."
    protoc \
      --proto_path="$CONTRACTS_DIR/company" \
      --go_out="$CONTRACTS_DIR/company/generated" \
      --go_opt=paths=source_relative \
      --go-grpc_out="$CONTRACTS_DIR/company/generated" \
      --go-grpc_opt=paths=source_relative \
      "$CONTRACTS_DIR/company/company.proto"
    echo "✓ Company done"
}

gen_application() {
    echo "→ Generating protobuf code for application..."
    protoc \
      --proto_path="$CONTRACTS_DIR/application" \
      --go_out="$CONTRACTS_DIR/application/generated" \
      --go_opt=paths=source_relative \
      --go-grpc_out="$CONTRACTS_DIR/application/generated" \
      --go-grpc_opt=paths=source_relative \
      "$CONTRACTS_DIR/application/application.proto"
    echo "✓ Application done"
}

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
