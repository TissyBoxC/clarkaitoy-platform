#!/usr/bin/env bash
# Upgrade the cloud stack with a verified pre-upgrade database backup.
#
# Database restore is intentionally not automatic: schema migrations can make
# a blind rollback destroy valid data written after the upgrade.
set -euo pipefail
set -E

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
compose_file="$cloud_dir/docker-compose.yml"
env_file="$cloud_dir/.env"
backup_root="${SPROUT_CLOUD_BACKUP_DIR:-$cloud_dir/migration-data/backups}"
backup_dir=""
previous_platform_version=""
previous_sub2api_version=""
platform_version=""
sub2api_version=""

usage() {
  cat >&2 <<'EOF'
用法: upgrade-cloud.sh <平台版本> [Sub2API 版本]

示例:
  ./scripts/upgrade-cloud.sh 0.10.0 0.2.16
  ./scripts/upgrade-cloud.sh 0.10.0

省略 Sub2API 版本时保留当前 .env 中的版本。
EOF
}

normalize_version() {
  raw_version="${1#v}"
  if ! printf '%s' "$raw_version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
    echo "版本格式无效: $1；必须使用 X.Y.Z 或 vX.Y.Z" >&2
    exit 1
  fi
  printf '%s' "$raw_version"
}

read_env_value() {
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

require_env_key() {
  key="$1"
  if ! grep -q "^${key}=" "$env_file"; then
    echo ".env 缺少必要变量: $key" >&2
    exit 1
  fi
}

# merge_missing_env_keys appends keys that a previous release did not have.
#
# Older deployments predate the release download service, so a strict
# validator would otherwise refuse to upgrade them. Existing values are never
# overwritten: only keys that are completely absent are appended.
merge_missing_env_keys() {
  template_file="$cloud_dir/.env.example"
  if [ ! -f "$template_file" ]; then
    return 0
  fi

  added_keys=""
  while IFS= read -r line; do
    case "$line" in
      ''|\#*)
        continue
        ;;
    esac
    key="${line%%=*}"
    case "$key" in
      *[!A-Z0-9_]*|'')
        continue
        ;;
    esac
    if grep -q "^${key}=" "$env_file"; then
      continue
    fi
    printf '%s\n' "$line" >> "$env_file"
    added_keys="$added_keys $key"
  done < "$template_file"

  if [ -n "$added_keys" ]; then
    echo "已从 .env.example 补齐缺失变量:$added_keys" >&2
    echo "带有 example 占位值的变量仍需手工填写后再升级。" >&2
  fi
}

# ensure_download_credentials creates the stable SFTP host keys and upload
# keypair when absent. Compose bind-mounts the host keys as files; a missing
# file makes Docker create a directory and sshd cannot start. An empty
# authorized_keys directory also makes atmoz/sftp initialization fail.
#
# Both scripts are idempotent and keep existing credentials intact.
ensure_download_credentials() {
  host_key_script="$cloud_dir/scripts/generate-download-ssh-host-keys.sh"
  release_key_script="$cloud_dir/scripts/generate-download-release-key.sh"
  key_dir="$cloud_dir/download/sftp"

  if [ ! -f "$key_dir/ssh_host_ed25519_key" ] ||
    [ ! -f "$key_dir/ssh_host_rsa_key" ]; then
    if [ ! -x "$host_key_script" ]; then
      echo "缺少发布文件服务主机密钥，且无法自动生成: $host_key_script" >&2
      return 1
    fi
    "$host_key_script"
  fi

  download_credentials_recreated="false"
  if [ ! -s "$key_dir/authorized_keys/release.pub" ]; then
    if [ ! -x "$release_key_script" ]; then
      echo "缺少发布文件上传密钥，且无法自动生成: $release_key_script" >&2
      return 1
    fi
    "$release_key_script"
    download_credentials_recreated="true"
  fi
}

write_env_versions() {
  temp_file="$(mktemp "${env_file}.tmp.XXXXXX")"
  if ! awk \
    -v platform_key="SPROUT_PLATFORM_VERSION" \
    -v platform_version="$platform_version" \
    -v sub2api_key="SPROUT_SUB2API_VERSION" \
    -v sub2api_version="$sub2api_version" '
      BEGIN {
        platform_found = 0
        sub2api_found = 0
      }
      index($0, platform_key "=") == 1 {
        print platform_key "=" platform_version
        platform_found = 1
        next
      }
      index($0, sub2api_key "=") == 1 {
        print sub2api_key "=" sub2api_version
        sub2api_found = 1
        next
      }
      {
        print
      }
      END {
        if (!platform_found || !sub2api_found) {
          exit 1
        }
      }
    ' "$env_file" > "$temp_file"; then
    rm -f "$temp_file"
    echo ".env 版本写入失败；文件未被替换。" >&2
    exit 1
  fi

  # Preserve the restrictive mode expected for production secrets.
  chmod 600 "$temp_file"
  mv "$temp_file" "$env_file"
}

ensure_postgres_ready() {
  postgres_id="$(docker compose -f "$compose_file" --env-file "$env_file" ps -q postgres)"
  if [ -z "$postgres_id" ]; then
    echo "PostgreSQL 未运行；先启动它以完成升级前备份。" >&2
    docker compose -f "$compose_file" --env-file "$env_file" up -d postgres >/dev/null
  fi

  for _ in $(seq 1 60); do
    postgres_id="$(docker compose -f "$compose_file" --env-file "$env_file" ps -q postgres)"
    if [ -n "$postgres_id" ] &&
      docker exec "$postgres_id" pg_isready \
        -U "$postgres_user" -d "$postgres_database" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done

  echo "PostgreSQL 未在预期时间内就绪，升级已停止。" >&2
  return 1
}

backup_databases() {
  backup_dir="$backup_root/$(date +%Y%m%d-%H%M%S)"
  if ! mkdir -p "$backup_dir"; then
    backup_dir=""
    return 1
  fi
  chmod 700 "$backup_dir"

  {
    printf 'platform_version=%s\n' "$previous_platform_version"
    printf 'sub2api_version=%s\n' "$previous_sub2api_version"
    printf 'created_at=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  } > "$backup_dir/versions.before-upgrade"

  if ! docker exec "$postgres_id" pg_dump \
    -U "$postgres_user" \
    --format=custom \
    --no-owner \
    --no-privileges \
    "$device_platform_database" > "$backup_dir/sprout_device_platform.dump"; then
    rm -f "$backup_dir/sprout_device_platform.dump"
    return 1
  fi

  if ! docker exec "$postgres_id" pg_dump \
    -U "$postgres_user" \
    --format=custom \
    --no-owner \
    --no-privileges \
    "$sub2api_database" > "$backup_dir/sprout_sub2api.dump"; then
    rm -f "$backup_dir/sprout_sub2api.dump"
    return 1
  fi

  (
    cd "$backup_dir"
    sha256sum \
      versions.before-upgrade \
      sprout_device_platform.dump \
      sprout_sub2api.dump > SHA256SUMS
  )
  chmod 600 \
    "$backup_dir/versions.before-upgrade" \
    "$backup_dir/sprout_device_platform.dump" \
    "$backup_dir/sprout_sub2api.dump" \
    "$backup_dir/SHA256SUMS"
}

print_rollback_instructions() {
  status="$1"
  echo "" >&2
  echo "升级失败，未自动回滚数据库。" >&2
  echo "失败退出码: $status" >&2
  if [ -n "$backup_dir" ]; then
    echo "升级前备份: $backup_dir" >&2
  else
    echo "升级前备份尚未完成；请先检查 PostgreSQL 状态。" >&2
  fi
  echo "" >&2
  echo "容器回滚步骤：" >&2
  echo "1. 将 $env_file 中的版本恢复为：" >&2
  echo "   SPROUT_PLATFORM_VERSION=$previous_platform_version" >&2
  echo "   SPROUT_SUB2API_VERSION=$previous_sub2api_version" >&2
  echo "2. cd $cloud_dir" >&2
  echo "3. docker compose -f docker-compose.yml --env-file .env pull device_platform voice_gateway admin_web upgrade_worker sub2api" >&2
  echo "4. docker compose -f docker-compose.yml --env-file .env up -d --remove-orphans" >&2
  echo "5. ./scripts/check-stack.sh" >&2
  echo "" >&2
  echo "只有确认数据库损坏且备份校验通过后，才手工恢复数据库。" >&2
  echo "禁止执行 docker compose down -v；该命令会删除所有数据卷。" >&2
  exit "$status"
}

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  usage
  exit 1
fi

platform_version="$(normalize_version "$1")"

if [ ! -f "$env_file" ]; then
  echo "缺少云端环境文件: $env_file" >&2
  exit 1
fi
if [ ! -f "$compose_file" ]; then
  echo "缺少 Compose 文件: $compose_file" >&2
  exit 1
fi

require_env_key "SPROUT_PLATFORM_VERSION"
require_env_key "SPROUT_SUB2API_VERSION"
previous_platform_version="$(read_env_value SPROUT_PLATFORM_VERSION)"
previous_sub2api_version="$(read_env_value SPROUT_SUB2API_VERSION)"

if [ "$#" -eq 2 ]; then
  sub2api_version="$(normalize_version "$2")"
else
  sub2api_version="$previous_sub2api_version"
fi

if ! printf '%s' "$previous_sub2api_version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo ".env 中的 SPROUT_SUB2API_VERSION 格式无效: $previous_sub2api_version" >&2
  exit 1
fi

if command -v git >/dev/null 2>&1 &&
  git -C "$cloud_dir" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  # The cloud bundle is deployed as a subset of the platform repository.
  # Only reject local edits inside this deployment directory; unrelated source
  # work in the parent repository must not block an infrastructure upgrade.
  git_status="$(
    git -C "$cloud_dir" status --porcelain -- . |
      grep -vE '^(\?\?|!!) (migration-data|mosquitto/certs)/' || true
  )"
  if [ -n "$git_status" ]; then
    echo "云端部署目录存在未提交变更，拒绝升级。请先提交或清理：" >&2
    printf '%s\n' "$git_status" >&2
    exit 1
  fi
else
  echo "未检测到 Git 工作区；跳过工作区检查。" >&2
fi

merge_missing_env_keys
ensure_download_credentials

"$cloud_dir/scripts/validate-cloud-env.sh" "$env_file"

postgres_user="$(read_env_value SPROUT_POSTGRES_USER)"
postgres_database="$(read_env_value SPROUT_POSTGRES_DB)"
device_platform_database="$(read_env_value SPROUT_DEVICE_PLATFORM_DATABASE_NAME)"
sub2api_database="$(read_env_value SPROUT_SUB2API_DATABASE_NAME)"

if [ -z "$postgres_user" ] ||
  [ -z "$postgres_database" ] ||
  [ -z "$device_platform_database" ] ||
  [ -z "$sub2api_database" ]; then
  echo ".env 缺少 PostgreSQL 用户或数据库名，无法备份。" >&2
  exit 1
fi

ensure_postgres_ready
backup_databases

echo "升级前备份完成: $backup_dir"
# Only arm rollback guidance once a mutation is possible; pre-flight failures
# change nothing and must not look like a failed upgrade.
trap 'print_rollback_instructions "$?"' ERR
write_env_versions
echo "版本已原子切换到 platform=$platform_version sub2api=$sub2api_version"

docker compose -f "$compose_file" --env-file "$env_file" pull \
  device_platform voice_gateway admin_web upgrade_worker sub2api
if [ "$download_credentials_recreated" = "true" ]; then
  # A running container that failed the first user initialization stores its
  # marker in the container filesystem. Recreate the stateless download
  # services so the newly generated public key is imported.
  docker compose -f "$compose_file" --env-file "$env_file" \
    up -d --force-recreate download_init download_ftp download_http
fi
docker compose -f "$compose_file" --env-file "$env_file" up -d --remove-orphans
"$cloud_dir/scripts/check-stack.sh"

echo ""
echo "升级完成。"
echo "平台版本: $platform_version"
echo "Sub2API 版本: $sub2api_version"
echo "升级前备份: $backup_dir"
