#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_DIR="$SCRIPT_DIR/../keys"
TEST_KEYS_DIR="$SCRIPT_DIR/../keys/test"

usage() {
  echo "Usage: $0 [prod|test|all]"
  echo ""
  echo "  $0 prod  — generate production key pair to backend/keys/"
  echo "  $0 test  — generate test key pair to backend/keys/test/"
  echo "  $0 all   — generate both"
  echo ""
  echo "  prod keys are gitignored and used in docker-compose.yml"
  echo "  test keys are committed and used in docker-compose.test.yml"
}

gen_pair() {
  local dir=$1
  local label=$2

  mkdir -p "$dir"

  if [ -f "$dir/private.pem" ] || [ -f "$dir/public.pem" ]; then
    echo "WARNING: keys already exist in $dir"
    read -r -p "Overwrite? [y/N] " confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
      echo "Skipped $label keys."
      return
    fi
  fi

  echo "Generating $label EC P-256 key pair in $dir ..."
  openssl ecparam -name prime256v1 -genkey -noout -out "$dir/private.pem" 2>/dev/null
  openssl ec -in "$dir/private.pem" -pubout -out "$dir/public.pem" 2>/dev/null

  chmod 600 "$dir/private.pem"
  chmod 644 "$dir/public.pem"

  echo "  private.pem — $(wc -c < "$dir/private.pem") bytes (keep secret)"
  echo "  public.pem  — $(wc -c < "$dir/public.pem") bytes (safe to share)"
  echo "Done: $label keys generated."
}

case "${1:-all}" in
  prod) gen_pair "$KEYS_DIR"      "production" ;;
  test) gen_pair "$TEST_KEYS_DIR" "test" ;;
  all)
    gen_pair "$KEYS_DIR"      "production"
    gen_pair "$TEST_KEYS_DIR" "test"
    ;;
  -h|--help) usage ;;
  *) echo "Unknown option: $1"; usage; exit 1 ;;
esac
