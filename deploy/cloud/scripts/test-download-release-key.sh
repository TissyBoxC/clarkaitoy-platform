#!/usr/bin/env bash
# Verify the release key generator is idempotent and never overwrites a key.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
script="$cloud_dir/scripts/generate-download-release-key.sh"
test_root="$(mktemp -d)"
trap 'rm -rf "$test_root"' EXIT

mkdir -p "$test_root/cloud/download/sftp/authorized_keys" "$test_root/cloud/scripts"
cp "$script" "$test_root/cloud/scripts/generate-download-release-key.sh"
chmod +x "$test_root/cloud/scripts/generate-download-release-key.sh"

private_key="$test_root/cloud/migration-data/download-sftp/release_ed25519"
public_key="$test_root/cloud/download/sftp/authorized_keys/release.pub"

SPROUT_DOWNLOAD_SFTP_PRIVATE_KEY_FILE="$private_key" \
  SPROUT_DOWNLOAD_SFTP_AUTHORIZED_KEYS_DIR="$test_root/cloud/download/sftp/authorized_keys" \
  "$test_root/cloud/scripts/generate-download-release-key.sh" >/dev/null

test -s "$private_key"
test -s "$public_key"
first_private_digest="$(sha256sum "$private_key" | awk '{print $1}')"
first_public_digest="$(sha256sum "$public_key" | awk '{print $1}')"

SPROUT_DOWNLOAD_SFTP_PRIVATE_KEY_FILE="$private_key" \
  SPROUT_DOWNLOAD_SFTP_AUTHORIZED_KEYS_DIR="$test_root/cloud/download/sftp/authorized_keys" \
  "$test_root/cloud/scripts/generate-download-release-key.sh" >/dev/null

test "$first_private_digest" = "$(sha256sum "$private_key" | awk '{print $1}')"
test "$first_public_digest" = "$(sha256sum "$public_key" | awk '{print $1}')"

echo "download release key generator test passed"
