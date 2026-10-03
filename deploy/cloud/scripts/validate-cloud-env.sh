#!/usr/bin/env bash
# Reject incomplete or malformed cloud values before Compose starts containers.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
env_file="${1:-$cloud_dir/.env}"

if [ ! -f "$env_file" ]; then
  echo "Cloud environment file not found: $env_file" >&2
  echo "Copy .env.example to .env and configure every required value first." >&2
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

is_hex_64() {
  value="$1"
  [ "${#value}" -eq 64 ] && printf '%s' "$value" | grep -Eq '^[0-9a-fA-F]{64}$'
}

validation_failed=0

require_value() {
  key="$1"
  value="$(read_value "$key")"
  if [ -z "$value" ]; then
    echo "Invalid cloud environment: $key is missing or empty." >&2
    validation_failed=1
    return
  fi
  case "$value" in
    replace-with-*|change-me*|your-*)
      echo "Invalid cloud environment: $key still contains an example value." >&2
      validation_failed=1
      ;;
  esac
}

required_keys="
SPROUT_POSTGRES_PASSWORD
SPROUT_REDIS_PASSWORD
SPROUT_INTERNAL_SERVICE_TOKEN
SPROUT_AUTH_ACCESS_TOKEN_SECRET
SPROUT_AI_CREDENTIAL_KEY
SPROUT_MFA_CREDENTIAL_KEY
SPROUT_SUB2API_API_KEY
SPROUT_SUB2API_ADMIN_EMAIL
SPROUT_SUB2API_ADMIN_PASSWORD
SPROUT_SUB2API_JWT_SECRET
SPROUT_SUB2API_TOTP_ENCRYPTION_KEY
SPROUT_DOWNLOAD_SFTP_PORT
SPROUT_DOWNLOAD_HTTP_PORT
SPROUT_DOWNLOAD_SFTP_USER
SPROUT_DOWNLOAD_SFTP_GID
SPROUT_DOWNLOAD_PUBLIC_BASE_URL
"

for key in $required_keys; do
  require_value "$key"
done

totp_key="$(read_value SPROUT_SUB2API_TOTP_ENCRYPTION_KEY)"
if [ -n "$totp_key" ] && ! is_hex_64 "$totp_key"; then
  echo "Invalid cloud environment: SPROUT_SUB2API_TOTP_ENCRYPTION_KEY must be exactly 64 hexadecimal characters." >&2
  echo "Generate one with: openssl rand -hex 32" >&2
  validation_failed=1
fi

admin_email="$(read_value SPROUT_SUB2API_ADMIN_EMAIL)"
if [ -n "$admin_email" ] && ! printf '%s' "$admin_email" | grep -Eq '^[^@[:space:]]+@[^@[:space:]]+\.[^@[:space:]]+$'; then
  echo "Invalid cloud environment: SPROUT_SUB2API_ADMIN_EMAIL must be a valid email address." >&2
  validation_failed=1
fi

if [ "$validation_failed" -ne 0 ]; then
  exit 1
fi

echo "Cloud environment validation passed: $env_file"
