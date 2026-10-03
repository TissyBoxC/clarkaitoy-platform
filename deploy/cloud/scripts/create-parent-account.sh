#!/usr/bin/env bash
# Create one parent account through the maintenance-only platform command.
#
# Prefer SPROUT_PARENT_PASSWORD so the password is not stored in shell history.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
compose_file="$cloud_dir/docker-compose.yml"
env_file="$cloud_dir/.env"
phone=""
password="${SPROUT_PARENT_PASSWORD:-}"
guardian_family_name=""
child_nickname=""
child_birthday=""

usage() {
  cat >&2 <<'EOF'
用法:
  SPROUT_PARENT_PASSWORD='至少8位且包含字母和数字' \
    ./scripts/create-parent-account.sh \
      --phone 13800138000 \
      --guardian-family-name 王 \
      --child-nickname 小芽 \
      --child-birthday 2022-05-20

参数:
  --phone                 必填，11 位中国大陆手机号
  --guardian-family-name  可选，家长姓氏
  --child-nickname        可选，宝贝昵称
  --child-birthday        可选，格式 YYYY-MM-DD

密码优先从 SPROUT_PARENT_PASSWORD 读取；未设置时会在终端中隐藏输入。
也可以使用 --password，但不建议，因为参数可能进入 shell 历史。
EOF
}

require_option_value() {
  option="$1"
  value="${2:-}"
  if [ -z "$value" ]; then
    echo "$option 缺少参数值" >&2
    usage
    exit 1
  fi
  printf '%s' "$value"
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --phone)
      phone="$(require_option_value "$1" "${2:-}")"
      shift 2
      ;;
    --password)
      password="$(require_option_value "$1" "${2:-}")"
      shift 2
      ;;
    --guardian-family-name)
      guardian_family_name="$(require_option_value "$1" "${2:-}")"
      shift 2
      ;;
    --child-nickname)
      child_nickname="$(require_option_value "$1" "${2:-}")"
      shift 2
      ;;
    --child-birthday)
      child_birthday="$(require_option_value "$1" "${2:-}")"
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
if [ -n "$child_birthday" ]; then
  if ! printf '%s' "$child_birthday" | grep -Eq '^[0-9]{4}-[0-9]{2}-[0-9]{2}$' ||
    ! date -d "$child_birthday" +%F >/dev/null 2>&1; then
    echo "宝贝生日格式无效，请使用 YYYY-MM-DD。" >&2
    exit 1
  fi
fi

if [ -z "$password" ]; then
  if [ ! -t 0 ]; then
    echo "未提供密码。请在终端运行，或设置 SPROUT_PARENT_PASSWORD。" >&2
    exit 1
  fi
  printf '请输入家长账号密码（输入不会显示）: ' >&2
  IFS= read -r -s password
  printf '\n' >&2
fi

if [ "${#password}" -lt 8 ] || [ "${#password}" -gt 128 ]; then
  echo "密码长度必须为 8 到 128 个字符。" >&2
  exit 1
fi

device_platform_container_id="$(
  docker compose -f "$compose_file" --env-file "$env_file" ps -q device_platform
)"
if [ -z "$device_platform_container_id" ]; then
  echo "device_platform 容器未运行，请先启动云端服务。" >&2
  exit 1
fi

admin_arguments=(
  create-parent
  --phone "$phone"
)
if [ -n "$guardian_family_name" ]; then
  admin_arguments+=(--guardian-family-name "$guardian_family_name")
fi
if [ -n "$child_nickname" ]; then
  admin_arguments+=(--child-nickname "$child_nickname")
fi
if [ -n "$child_birthday" ]; then
  admin_arguments+=(--child-birthday "$child_birthday")
fi

# Pass the password through the process environment instead of command-line
# arguments, which may be visible through process inspection.
export DEVICE_PLATFORM_ADMIN_PASSWORD="$password"
trap 'unset DEVICE_PLATFORM_ADMIN_PASSWORD' EXIT

set +e
command_output="$(
  docker compose -f "$compose_file" --env-file "$env_file" exec -T \
    -e DEVICE_PLATFORM_ADMIN_PASSWORD \
    device_platform \
    /device-platform-admin "${admin_arguments[@]}" 2>&1
)"
command_status="$?"
set -e
unset DEVICE_PLATFORM_ADMIN_PASSWORD
trap - EXIT

if [ "$command_status" -ne 0 ]; then
  echo "家长账号创建失败。" >&2
  printf '%s\n' "$command_output" >&2
  exit "$command_status"
fi

masked_phone="${phone:0:3}****${phone:7:4}"
echo "家长账号已创建：$masked_phone"
if printf '%s' "$command_output" | grep -q 'AI 账户暂未开通'; then
  echo "家长 AI 账号暂未开通；登录家长端或再次执行检查脚本时会重试。"
  echo "检查命令: ./scripts/check-family-ai-account.sh --phone $phone"
else
  echo "家长 AI 账号已同步。"
fi
