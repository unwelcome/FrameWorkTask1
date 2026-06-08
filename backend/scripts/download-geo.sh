#!/usr/bin/env bash
# Download MaxMind GeoLite2 databases required by the gateway service.
#
# Usage:
#   MAXMIND_LICENSE_KEY=your_key ./backend/scripts/download-geo.sh
#
# Or set the key in your shell profile / CI secrets:
#   export MAXMIND_LICENSE_KEY=your_key
#
# Sign up and get a free license key at:
#   https://www.maxmind.com/en/geolite2/signup
#
# The databases are updated by MaxMind every Tuesday.
# Re-run this script periodically (or set up a cron job) to keep them current.

set -euo pipefail

# Load MAXMIND_LICENSE_KEY from root .env if not already set in the environment
if [[ -z "${MAXMIND_LICENSE_KEY:-}" ]]; then
  ROOT_ENV="$(cd "$(dirname "$0")/../.." && pwd)/.env"
  if [[ -f "$ROOT_ENV" ]]; then
    MAXMIND_LICENSE_KEY="$(grep -E '^MAXMIND_LICENSE_KEY=' "$ROOT_ENV" | cut -d '=' -f2- | tr -d '[:space:]')"
  fi
fi

LICENSE_KEY="${MAXMIND_LICENSE_KEY:-}"
if [[ -z "$LICENSE_KEY" ]]; then
  echo "Error: MAXMIND_LICENSE_KEY is not set." >&2
  echo "  Set it in the root .env file: MAXMIND_LICENSE_KEY=your_key" >&2
  echo "  Or export it: export MAXMIND_LICENSE_KEY=your_key" >&2
  exit 1
fi

DEST_DIR="$(cd "$(dirname "$0")/.." && pwd)/geo"
mkdir -p "$DEST_DIR"

BASE_URL="https://download.maxmind.com/app/geoip_download"

download_db() {
  local edition="$1"
  local filename="${edition}.mmdb"
  local archive="${edition}.tar.gz"
  local url="${BASE_URL}?edition_id=${edition}&license_key=${LICENSE_KEY}&suffix=tar.gz"

  echo "Downloading ${edition}..."
  curl -fsSL "$url" -o "/tmp/${archive}"

  echo "Extracting ${filename}..."
  tar -xzf "/tmp/${archive}" -C /tmp --wildcards "*/${filename}" --strip-components=1
  mv "/tmp/${filename}" "${DEST_DIR}/${filename}"
  rm -f "/tmp/${archive}"

  echo "Saved to ${DEST_DIR}/${filename}"
}

download_db "GeoLite2-City"
download_db "GeoLite2-ASN"

echo ""
echo "Done. Databases saved to backend/geo/"
