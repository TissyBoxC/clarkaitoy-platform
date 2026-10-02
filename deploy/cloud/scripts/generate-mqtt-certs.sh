#!/usr/bin/env sh
# Generate a private CA plus broker and shared device certificates.
#
# The shared device certificate is a migration aid for the current early
# stage. Each production device must use a unique certificate before release.
set -eu

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "usage: $0 <mqtt-hostname-or-public-ip> [--rotate-ca]" >&2
  exit 1
fi

mqtt_host="$1"
rotate_ca="${2:-}"
if [ -n "$rotate_ca" ] && [ "$rotate_ca" != "--rotate-ca" ]; then
  echo "unknown option: $rotate_ca" >&2
  exit 1
fi

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
cert_root="$script_dir/../mosquitto/certs"
broker_cert_dir="$cert_root/broker"
device_cert_dir="$cert_root/device"
env_file="$script_dir/../.env"
endpoints_file="$script_dir/../../public-endpoints.env"

if [ -f "$env_file" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$env_file"
  set +a
fi

if [ -f "$endpoints_file" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$endpoints_file"
  set +a
fi

# The official Mosquitto image runs as this fixed UID/GID.
mqtt_uid="${SPROUT_MQTT_UID:-1883}"
mqtt_gid="${SPROUT_MQTT_GID:-1883}"
device_uid="${SPROUT_DEVICE_PLATFORM_UID:-65532}"
device_gid="${SPROUT_DEVICE_PLATFORM_GID:-65532}"

case "$mqtt_host" in
  *[!0-9.]*)
    public_san="DNS:$mqtt_host"
    ;;
  *)
    public_san="IP:$mqtt_host"
    ;;
esac

mkdir -p "$broker_cert_dir" "$device_cert_dir"

if [ "$rotate_ca" = "--rotate-ca" ]; then
  rm -f "$broker_cert_dir/ca.crt" "$broker_cert_dir/ca.key"
fi

# Keep the existing trust root across permission-only migrations. Replacing it
# would invalidate every device that already trusts the current CA.
if [ ! -f "$broker_cert_dir/ca.crt" ] || [ ! -f "$broker_cert_dir/ca.key" ]; then
  if [ -f "$cert_root/ca.crt" ] && [ -f "$cert_root/ca.key" ]; then
    cp "$cert_root/ca.crt" "$broker_cert_dir/ca.crt"
    cp "$cert_root/ca.key" "$broker_cert_dir/ca.key"
  else
    openssl genrsa -out "$broker_cert_dir/ca.key" 4096
    openssl req -x509 -new -nodes -key "$broker_cert_dir/ca.key" \
      -sha256 -days 3650 -subj "/CN=Sprout Cloud MQTT CA" \
      -out "$broker_cert_dir/ca.crt"
  fi
fi

rm -f \
  "$broker_cert_dir/server.crt" "$broker_cert_dir/server.key" "$broker_cert_dir/server.csr" \
  "$broker_cert_dir/healthcheck.crt" "$broker_cert_dir/healthcheck.key" \
  "$broker_cert_dir/healthcheck.csr" \
  "$device_cert_dir/ca.crt" "$device_cert_dir/device.crt" \
  "$device_cert_dir/device.key" "$device_cert_dir/device.csr"

cp "$broker_cert_dir/ca.crt" "$device_cert_dir/ca.crt"

openssl genrsa -out "$broker_cert_dir/server.key" 2048
openssl req -new -key "$broker_cert_dir/server.key" \
  -subj "/CN=$mqtt_host" -out "$broker_cert_dir/server.csr"
cat > "$broker_cert_dir/server.ext" <<EOF
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:mqtt,IP:127.0.0.1,$public_san
EOF
openssl x509 -req -in "$broker_cert_dir/server.csr" \
  -CA "$broker_cert_dir/ca.crt" -CAkey "$broker_cert_dir/ca.key" \
  -CAcreateserial -out "$broker_cert_dir/server.crt" \
  -days 825 -sha256 -extfile "$broker_cert_dir/server.ext"

openssl genrsa -out "$broker_cert_dir/healthcheck.key" 2048
openssl req -new -key "$broker_cert_dir/healthcheck.key" \
  -subj "/CN=sprout-healthcheck" -out "$broker_cert_dir/healthcheck.csr"
cat > "$broker_cert_dir/healthcheck.ext" <<EOF
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=clientAuth
EOF
openssl x509 -req -in "$broker_cert_dir/healthcheck.csr" \
  -CA "$broker_cert_dir/ca.crt" -CAkey "$broker_cert_dir/ca.key" \
  -CAcreateserial -out "$broker_cert_dir/healthcheck.crt" \
  -days 825 -sha256 -extfile "$broker_cert_dir/healthcheck.ext"

openssl genrsa -out "$device_cert_dir/device.key" 2048
openssl req -new -key "$device_cert_dir/device.key" \
  -subj "/CN=sprout-device" -out "$device_cert_dir/device.csr"
cat > "$device_cert_dir/device.ext" <<EOF
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=clientAuth
EOF
openssl x509 -req -in "$device_cert_dir/device.csr" \
  -CA "$broker_cert_dir/ca.crt" -CAkey "$broker_cert_dir/ca.key" \
  -CAcreateserial -out "$device_cert_dir/device.crt" \
  -days 825 -sha256 -extfile "$device_cert_dir/device.ext"

rm -f "$broker_cert_dir"/*.csr "$broker_cert_dir"/*.ext "$broker_cert_dir"/*.srl
rm -f "$device_cert_dir"/*.csr "$device_cert_dir"/*.ext

# The broker and device platform run as different users. Keep each private
# key readable only by the process that actually needs it.
chown "$mqtt_uid:$mqtt_gid" \
  "$broker_cert_dir/server.key" "$broker_cert_dir/server.crt" \
  "$broker_cert_dir/healthcheck.key" "$broker_cert_dir/healthcheck.crt"
chown "$device_uid:$device_gid" \
  "$device_cert_dir/device.key" "$device_cert_dir/device.crt"
chmod 600 "$broker_cert_dir/ca.key"
chmod 640 "$broker_cert_dir/server.key" "$broker_cert_dir/healthcheck.key"
chmod 644 \
  "$broker_cert_dir/server.crt" "$broker_cert_dir/healthcheck.crt" \
  "$broker_cert_dir/ca.crt"
chmod 640 "$device_cert_dir/device.key"
chmod 644 "$device_cert_dir/device.crt" "$device_cert_dir/ca.crt"

echo "MQTT certificates generated for host: $mqtt_host"
