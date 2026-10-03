#!/usr/bin/env bash
# Create stable SFTP host keys so CI and operators do not see a new
# fingerprint after every container recreation.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
key_dir="$cloud_dir/download/sftp"
keys_dir="$key_dir/authorized_keys"

mkdir -p "$key_dir" "$keys_dir"
chmod 700 "$key_dir"
chmod 755 "$keys_dir"

if [ ! -f "$key_dir/ssh_host_ed25519_key" ]; then
  ssh-keygen -q -t ed25519 -N '' -f "$key_dir/ssh_host_ed25519_key"
fi

if [ ! -f "$key_dir/ssh_host_rsa_key" ]; then
  ssh-keygen -q -t rsa -b 4096 -N '' -f "$key_dir/ssh_host_rsa_key"
fi

chmod 600 "$key_dir/ssh_host_ed25519_key" "$key_dir/ssh_host_rsa_key"
chmod 644 "$key_dir/ssh_host_ed25519_key.pub" "$key_dir/ssh_host_rsa_key.pub"

echo "SFTP host keys are ready in $key_dir"
