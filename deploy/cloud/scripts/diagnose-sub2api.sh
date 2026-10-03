#!/usr/bin/env bash
# Collect Sub2API startup diagnostics without changing application data.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
compose_file="$cloud_dir/docker-compose.yml"
env_file="$cloud_dir/.env"
project_name="$(awk '/^name:[[:space:]]*/{print $2; exit}' "$compose_file")"
project_name="${project_name:-sprout}"
container_name="${project_name}-sub2api-1"

cd "$cloud_dir"

echo "== environment validation =="
if ! "$cloud_dir/scripts/validate-cloud-env.sh" "$env_file"; then
  echo "Fix the environment validation errors above before diagnosing the service." >&2
fi

echo
echo "== compose status =="
docker compose --env-file "$env_file" ps sub2api postgres redis

echo
echo "== container state =="
docker inspect "$container_name" \
  --format 'status={{.State.Status}} exit={{.State.ExitCode}} error={{.State.Error}} health={{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}'

echo
echo "== container environment (secrets redacted) =="
docker inspect "$container_name" \
  --format '{{range .Config.Env}}{{println .}}{{end}}' |
  sed -E 's/^((DATABASE|REDIS)_[A-Z_]*(PASSWORD|SECRET|TOKEN)|ADMIN_PASSWORD|JWT_SECRET|TOTP_ENCRYPTION_KEY|SPROUT_INTERNAL_API_TOKEN)=.*/\1=<redacted>/'

echo
echo "== logs =="
docker logs --tail=250 "$container_name" 2>&1 || true

echo
echo "== data volume =="
volume_name="$(
  docker volume ls --format '{{.Name}}' |
    awk -v prefix="${project_name}_" '$0 ~ "^" prefix ".*sprout_sub2api_data$" {print; exit}'
)"
if [ -z "$volume_name" ]; then
  echo "Sub2API data volume not found"
else
  echo "volume: $volume_name"
  docker run --rm -v "$volume_name:/target:ro" alpine:3.20 \
    sh -c 'ls -la /target; find /target -maxdepth 2 -type f -printf "%M %u:%g %p\n" 2>/dev/null | sort'
fi

echo
echo "== connectivity =="
docker run --rm --network "${project_name}_default" alpine:3.20 \
  sh -c 'nc -zv postgres 5432; nc -zv redis 6379'
