#!/usr/bin/env bash
# Verify the public read path without exposing SFTP credentials.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
env_file="$cloud_dir/.env"
endpoints_file="$cloud_dir/../public-endpoints.env"

if [ ! -f "$env_file" ]; then
  echo "Missing cloud environment file: $env_file" >&2
  exit 1
fi

read_value() {
  key="$1"
  awk -F= -v key="$key" '
    $1 == key {
      sub(/^[^=]*=/, "")
      sub(/\r$/, "")
      print
      exit
    }
  ' "$env_file"
}

public_base_url="$(read_value SPROUT_DOWNLOAD_PUBLIC_BASE_URL)"
if [ -z "$public_base_url" ] && [ -f "$endpoints_file" ]; then
  # shellcheck disable=SC1090
  . "$endpoints_file"
  public_base_url="${SPROUT_DOWNLOAD_PUBLIC_URL:-}"
fi

if [ -z "$public_base_url" ]; then
  echo "SPROUT_DOWNLOAD_PUBLIC_BASE_URL is not configured." >&2
  exit 1
fi

public_base_url="${public_base_url%/}"
index_payload="$(curl --fail --silent --show-error --retry 5 --retry-delay 2 "$public_base_url/index.json")"
printf '%s\n' "$index_payload" | grep -q '"schema_version"' || {
  echo "Download index is not valid JSON." >&2
  exit 1
}

status="$(curl --silent --show-error --output /dev/null --write-out '%{http_code}' --head "$public_base_url/")"
if [ "$status" != "404" ]; then
  echo "Expected / to return 404, got HTTP $status." >&2
  exit 1
fi

echo "Download service is healthy: $public_base_url/index.json"
