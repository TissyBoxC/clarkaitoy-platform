#!/usr/bin/env bash
# Create the upload credential used by GitHub Actions when it is absent.
#
# The private key stays in ignored deployment data. The public key is placed
# in the authorized_keys directory so the SFTP entrypoint can initialize the
# release user without failing on an empty key directory.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
authorized_keys_dir="${SPROUT_DOWNLOAD_SFTP_AUTHORIZED_KEYS_DIR:-$cloud_dir/download/sftp/authorized_keys}"
private_key_path="${SPROUT_DOWNLOAD_SFTP_PRIVATE_KEY_FILE:-$cloud_dir/migration-data/download-sftp/release_ed25519}"
public_key_path="$authorized_keys_dir/release.pub"

mkdir -p "$authorized_keys_dir" "$(dirname -- "$private_key_path")"
chmod 700 "$authorized_keys_dir"

umask 077
if [ ! -s "$private_key_path" ]; then
  if [ -s "$public_key_path" ]; then
    echo "发布公钥已存在，但找不到对应私钥: $private_key_path" >&2
    echo "请恢复 GitHub Actions 使用的原私钥，或删除公钥后重新生成。" >&2
    exit 1
  fi
  ssh-keygen -q -t ed25519 -N '' -C 'sprout-release-upload' -f "$private_key_path"
fi

if [ ! -s "$public_key_path" ] ||
  ! cmp -s "$public_key_path" "$private_key_path.pub"; then
  ssh-keygen -y -f "$private_key_path" > "$public_key_path"
fi

chmod 600 "$private_key_path" "$private_key_path.pub"
chmod 644 "$public_key_path"

echo "SFTP 发布密钥已准备完成。"
echo "私钥路径: $private_key_path"
echo "公钥路径: $public_key_path"
echo "请在 GitHub 仓库 Secret 中设置 SPROUT_DOWNLOAD_SFTP_PRIVATE_KEY 为上述私钥内容。"
