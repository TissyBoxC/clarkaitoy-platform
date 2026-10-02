#!/usr/bin/env bash
# Verify that the cloud stack is healthy and reachable on loopback.
set -euo pipefail

cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
compose_file="$cloud_dir/docker-compose.yml"
env_file="$cloud_dir/.env"
endpoints_file="$cloud_dir/../public-endpoints.env"

cd "$cloud_dir"

if [ ! -f "$env_file" ]; then
  echo ".env is missing; copy .env.example to .env and configure it first" >&2
  exit 1
fi

if [ -f "$endpoints_file" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$endpoints_file"
  set +a
fi

"$cloud_dir/scripts/validate-cloud-env.sh" "$env_file"

docker compose -f "$compose_file" --env-file "$env_file" ps

check_service() {
  name="$1"
  service="$2"
  url="$3"
  for _ in $(seq 1 60); do
    if docker compose -f "$compose_file" --env-file "$env_file" exec -T "$service" \
      wget -q -T 5 -O /dev/null "$url"; then
      echo "ready: $name ($url)"
      return 0
    fi
    sleep 3
  done
  echo "not ready: $name ($url)" >&2
  return 1
}

check_service "device platform" "device_platform" "http://127.0.0.1:8081/readyz"
check_service "voice gateway" "voice_gateway" "http://127.0.0.1:8082/readyz"
check_service "sub2api" "sub2api" "http://127.0.0.1:8080/health"
check_service "admin console" "admin_web" "http://127.0.0.1:8080/healthz"
