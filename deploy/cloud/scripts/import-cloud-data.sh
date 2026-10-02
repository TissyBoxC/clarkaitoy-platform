#!/usr/bin/env bash
# Restore database dumps and Sub2API application data into a running stack.
#
# Run this from deploy/cloud after PostgreSQL, Redis, and MQTT are running.
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <migration-data-directory>" >&2
  exit 1
fi

data_dir="$(CDPATH= cd -- "$1" && pwd)"
cloud_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
compose_file="$cloud_dir/docker-compose.yml"
env_file="$cloud_dir/.env"

cd "$cloud_dir"

if [ ! -f "$env_file" ]; then
  echo ".env is missing; copy .env.example to .env and configure it first" >&2
  exit 1
fi

set -a
. "$env_file"
set +a

if command -v sha256sum >/dev/null 2>&1 && [ -f "$data_dir/SHA256SUMS" ]; then
  (cd "$data_dir" && sha256sum -c SHA256SUMS)
fi

postgres_id="$(docker compose -f "$compose_file" ps -q postgres)"

if [ -z "$postgres_id" ]; then
  echo "postgres container is not running" >&2
  exit 1
fi

sub2api_volume="$(
  docker volume ls --format '{{.Name}}' |
    awk '/_sprout_sub2api_data$/{print; exit}'
)"
if [ -z "$sub2api_volume" ]; then
  # Create the volume without starting Sub2API before the dump is restored.
  docker compose -f "$compose_file" create --no-deps sub2api >/dev/null
  sub2api_volume="$(
    docker volume ls --format '{{.Name}}' |
      awk '/_sprout_sub2api_data$/{print; exit}'
  )"
fi
if [ -z "$sub2api_volume" ]; then
  echo "Sub2API data volume could not be created" >&2
  exit 1
fi

# pg_restore --clean cannot reliably replace schemas while Sub2API holds
# connections, so keep the application stopped until both dumps are restored.
docker compose -f "$compose_file" stop sub2api >/dev/null 2>&1 || true

echo "Restoring sprout_device_platform"
docker exec -i "$postgres_id" pg_restore -U "$SPROUT_POSTGRES_USER" \
  -d "$SPROUT_DEVICE_PLATFORM_DATABASE_NAME" \
  --clean --if-exists --no-owner < "$data_dir/sprout_device_platform.dump"

echo "Restoring sprout_sub2api"
docker exec -i "$postgres_id" pg_restore -U "$SPROUT_POSTGRES_USER" \
  -d "$SPROUT_SUB2API_DATABASE_NAME" \
  --clean --if-exists --no-owner < "$data_dir/sprout_sub2api.dump"

echo "Restoring Sub2API non-secret application data"
docker run --rm -i -v "$sub2api_volume:/target" alpine:3.20 \
  sh -c 'tar -xzf - -C /target' \
  < "$data_dir/sprout_sub2api_data.tar.gz"

echo "Data restore complete"
