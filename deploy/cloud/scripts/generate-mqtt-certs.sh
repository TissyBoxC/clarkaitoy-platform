#!/usr/bin/env sh
# Generate a private CA plus broker and shared device certificates.
#
# The shared device certificate is a migration aid for the current early
# stage. Each production device must use a unique certificate before release.
set -eu

if [ "$#" -ne 1 ]; then
  echo "usage: $0 <mqtt-hostname-or-public-ip>" >&2
  exit 1
fi

mqtt_host="$1"
script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
cert_dir="$script_dir/../mosquitto/certs"

case "$mqtt_host" in
  *[!0-9.]*)
    public_san="DNS:$mqtt_host"
    ;;
  *)
    public_san="IP:$mqtt_host"
    ;;
esac

mkdir -p "$cert_dir"
rm -f \
  "$cert_dir/ca.crt" "$cert_dir/ca.key" \
  "$cert_dir/server.crt" "$cert_dir/server.key" "$cert_dir/server.csr" \
  "$cert_dir/device.crt" "$cert_dir/device.key" "$cert_dir/device.csr"

openssl genrsa -out "$cert_dir/ca.key" 4096
openssl req -x509 -new -nodes -key "$cert_dir/ca.key" -sha256 -days 3650 \
  -subj "/CN=Sprout Cloud MQTT CA" -out "$cert_dir/ca.crt"

openssl genrsa -out "$cert_dir/server.key" 2048
openssl req -new -key "$cert_dir/server.key" \
  -subj "/CN=$mqtt_host" -out "$cert_dir/server.csr"
cat > "$cert_dir/server.ext" <<EOF
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:mqtt,IP:127.0.0.1,$public_san
EOF
openssl x509 -req -in "$cert_dir/server.csr" -CA "$cert_dir/ca.crt" \
  -CAkey "$cert_dir/ca.key" -CAcreateserial -out "$cert_dir/server.crt" \
  -days 825 -sha256 -extfile "$cert_dir/server.ext"

openssl genrsa -out "$cert_dir/device.key" 2048
openssl req -new -key "$cert_dir/device.key" \
  -subj "/CN=sprout-device" -out "$cert_dir/device.csr"
cat > "$cert_dir/device.ext" <<EOF
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=clientAuth
EOF
openssl x509 -req -in "$cert_dir/device.csr" -CA "$cert_dir/ca.crt" \
  -CAkey "$cert_dir/ca.key" -CAcreateserial -out "$cert_dir/device.crt" \
  -days 825 -sha256 -extfile "$cert_dir/device.ext"

rm -f "$cert_dir"/*.csr "$cert_dir"/*.ext "$cert_dir"/*.srl
chmod 600 "$cert_dir"/*.key
chmod 644 "$cert_dir"/*.crt

echo "MQTT certificates generated for host: $mqtt_host"
