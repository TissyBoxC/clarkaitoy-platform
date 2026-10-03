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

# Compose bind-mounts these paths read-only. If the target is missing, Docker
# creates an empty directory instead of a file and sshd cannot start, so a
# stray directory from an earlier failed start must be cleared before key
# generation. Only empty directories are removed; unexpected content fails.
clear_non_file_path() {
  target_path="$1"
  if [ ! -e "$target_path" ] || [ -f "$target_path" ]; then
    return 0
  fi
  if [ -d "$target_path" ] && [ -z "$(ls -A "$target_path" 2>/dev/null)" ]; then
    rmdir "$target_path"
    return 0
  fi
  echo "路径已存在且不是普通文件，请先手工处理: $target_path" >&2
  return 1
}

clear_non_file_path "$key_dir/ssh_host_ed25519_key"
clear_non_file_path "$key_dir/ssh_host_rsa_key"

if [ ! -f "$key_dir/ssh_host_ed25519_key" ]; then
  ssh-keygen -q -t ed25519 -N '' -f "$key_dir/ssh_host_ed25519_key"
fi

if [ ! -f "$key_dir/ssh_host_rsa_key" ]; then
  ssh-keygen -q -t rsa -b 4096 -N '' -f "$key_dir/ssh_host_rsa_key"
fi

chmod 600 "$key_dir/ssh_host_ed25519_key" "$key_dir/ssh_host_rsa_key"
chmod 644 "$key_dir/ssh_host_ed25519_key.pub" "$key_dir/ssh_host_rsa_key.pub"

echo "SFTP host keys are ready in $key_dir"
