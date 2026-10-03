#!/usr/bin/env bash
# Check one parent account and its linked AI account without exposing secrets.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
compose_file="$cloud_dir/docker-compose.yml"
env_file="$cloud_dir/.env"
phone=""

usage() {
  cat >&2 <<'EOF'
用法:
  ./scripts/check-family-ai-account.sh --phone 13800138000

脚本只输出脱敏手机号、账号状态、AI 账号状态、凭据是否就绪和余额，
不会输出邮箱、API Key、密文或其他内部凭据。
EOF
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

while [ "$#" -gt 0 ]; do
  case "$1" in
    --phone)
      phone="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "未知参数: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [ ! -f "$env_file" ]; then
  echo "缺少云端环境文件: $env_file" >&2
  exit 1
fi
if [ ! -f "$compose_file" ]; then
  echo "缺少 Compose 文件: $compose_file" >&2
  exit 1
fi
if ! printf '%s' "$phone" | grep -Eq '^1[0-9]{10}$'; then
  echo "手机号格式无效，请输入 11 位中国大陆手机号。" >&2
  exit 1
fi

postgres_id="$(docker compose -f "$compose_file" --env-file "$env_file" ps -q postgres)"
if [ -z "$postgres_id" ]; then
  echo "PostgreSQL 容器未运行，无法检查家长账号。" >&2
  exit 1
fi

postgres_user="$(read_env_value SPROUT_POSTGRES_USER)"
database_name="$(read_env_value SPROUT_DEVICE_PLATFORM_DATABASE_NAME)"
if [ -z "$postgres_user" ] || [ -z "$database_name" ]; then
  echo ".env 缺少 PostgreSQL 用户或设备平台数据库名。" >&2
  exit 1
fi

query='
SELECT
  CASE
    WHEN length(parent_accounts.phone) = 11
      THEN substr(parent_accounts.phone, 1, 3) || '****' || right(parent_accounts.phone, 4)
    ELSE '***'
  END,
  parent_accounts.status,
  COALESCE(ai_accounts.status, 'missing'),
  CASE
    WHEN ai_accounts.id IS NULL THEN 'no'
    WHEN octet_length(ai_accounts.credential_ciphertext) > 0
      AND octet_length(ai_accounts.credential_nonce) > 0
      AND length(btrim(ai_accounts.provider_account_id)) > 0
      THEN 'yes'
    ELSE 'no'
  END,
  COALESCE(ai_accounts.balance_usd, 0)::text
FROM parent_accounts
LEFT JOIN ai_accounts
  ON ai_accounts.parent_account_id = parent_accounts.id
WHERE parent_accounts.phone = :'phone'
  AND parent_accounts.role = 'parent'
ORDER BY parent_accounts.created_at DESC;
'

result="$(
  docker exec "$postgres_id" psql \
    -U "$postgres_user" \
    -d "$database_name" \
    -v ON_ERROR_STOP=1 \
    -v phone="$phone" \
    -At \
    -F '|' \
    -c "$query"
)"

if [ -z "$result" ]; then
  echo "未找到该家长账号。"
  exit 1
fi

IFS='|' read -r masked_phone parent_status ai_status credential_ready balance_usd <<< "$result"

parent_status_label="$parent_status"
case "$parent_status" in
  active) parent_status_label="正常" ;;
  disabled) parent_status_label="已停用" ;;
  pending_deletion) parent_status_label="待注销" ;;
esac

ai_status_label="$ai_status"
case "$ai_status" in
  active) ai_status_label="正常" ;;
  suspended) ai_status_label="已暂停" ;;
  closed) ai_status_label="已关闭" ;;
  missing) ai_status_label="未创建" ;;
esac

credential_label="$credential_ready"
case "$credential_ready" in
  yes) credential_label="已就绪" ;;
  no) credential_label="未就绪" ;;
esac

echo "家长账号: $masked_phone"
echo "家长状态: $parent_status_label"
echo "AI 账号: $ai_status_label"
echo "AI 凭据: $credential_label"
echo "AI 余额: $balance_usd USD"
