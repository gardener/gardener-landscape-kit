#!/usr/bin/env bash
# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail

COMMAND="${1:-up}"
VALID_COMMANDS=("up" "down" "cleanup-registry-ca" "ensure-registry-tls")

INFRA_COMPOSE_FILE="$(dirname "$0")/infra/docker-compose.yaml"
DIR_REGISTRY_TLS="$(dirname "$0")/infra/registry/tls"
REGISTRY_HOST="glk-registry.local.gardener.cloud"
REGISTRY_CA_CN="GLK Local Registry CA"
REGISTRY_CA_FILE_BASE="glk-local-registry-ca"

SUDO=""
if [[ "$(id -u)" != "0" ]]; then
  SUDO="sudo "
fi

cleanup_registry_ca() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    echo "> Removing registry CA from macOS login keychain..."
    security delete-certificate -c "$REGISTRY_CA_CN" ~/Library/Keychains/login.keychain-db 2>/dev/null || true
  elif [[ "$(uname -s)" == "Linux" ]]; then
    if command -v update-ca-certificates >/dev/null 2>&1; then
      echo "> Removing registry CA from /usr/local/share/ca-certificates/..."
      ${SUDO}rm -f /usr/local/share/ca-certificates/${REGISTRY_CA_FILE_BASE}.crt
      ${SUDO}update-ca-certificates --fresh >/dev/null 2>&1 || true
    elif command -v update-ca-trust >/dev/null 2>&1; then
      echo "> Removing registry CA from /etc/pki/ca-trust/source/anchors/..."
      ${SUDO}rm -f /etc/pki/ca-trust/source/anchors/${REGISTRY_CA_FILE_BASE}.pem
      ${SUDO}update-ca-trust extract >/dev/null 2>&1 || true
    fi
  fi
}

# The local registry is served over HTTPS with a self-signed CA (generated once and reused across restarts).
# The Go-based tools that push to it (skaffold, helm) verify the server certificate, so the CA must be
# installed into the OS trust store.
ensure_registry_tls() {
  local ca_crt="$DIR_REGISTRY_TLS/ca.crt"
  local ca_key="$DIR_REGISTRY_TLS/ca.key"
  local tls_crt="$DIR_REGISTRY_TLS/registry.crt"
  local tls_key="$DIR_REGISTRY_TLS/registry.key"

  mkdir -p "$DIR_REGISTRY_TLS"

  # Regenerate if any certificate is expired or expiring within 30 days (certs are valid for 180 days).
  for cert_file in "$ca_crt" "$tls_crt"; do
    if [[ -f "$cert_file" ]] && ! openssl x509 -in "$cert_file" -noout -checkend 2592000 2>/dev/null; then
      echo "> Registry certificate $cert_file is expired or expiring soon, regenerating..."
      rm -f "$ca_crt" "$ca_key" "$tls_crt" "$tls_key"
      cleanup_registry_ca
      break
    fi
  done

  if [[ ! -f "$ca_crt" || ! -f "$tls_crt" || ! -f "$tls_key" ]]; then
    echo "> Generating self-signed CA and server certificate for $REGISTRY_HOST..."
    local cnf
    cnf="$(mktemp)"
    cat > "$cnf" <<EOF
[req]
distinguished_name = dn
[dn]
[v3_ca]
basicConstraints = critical,CA:TRUE
subjectKeyIdentifier = hash
keyUsage = critical,keyCertSign,cRLSign
nameConstraints = critical,permitted;DNS:$REGISTRY_HOST,permitted;DNS:localhost,permitted;IP:127.0.0.1/255.255.255.255,permitted;IP:::1/ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff
[v3_server]
basicConstraints = CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
keyUsage = critical,digitalSignature,keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = $REGISTRY_HOST
DNS.2 = localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

    openssl req -x509 -newkey rsa:4096 -sha256 -days 180 -nodes \
      -set_serial 1 \
      -keyout "$ca_key" -out "$ca_crt" \
      -subj "/CN=$REGISTRY_CA_CN" \
      -config "$cnf" -extensions v3_ca 2>/dev/null

    local csr
    csr="$(mktemp)"
    openssl req -newkey rsa:4096 -sha256 -nodes \
      -keyout "$tls_key" -out "$csr" \
      -subj "/CN=$REGISTRY_HOST" -config "$cnf" 2>/dev/null
    openssl x509 -req -in "$csr" -CA "$ca_crt" -CAkey "$ca_key" \
      -set_serial 1 \
      -out "$tls_crt" -days 180 -sha256 \
      -extfile "$cnf" -extensions v3_server 2>/dev/null

    rm -f "$cnf" "$csr"

    # Restart the registry container if it is already running so it picks up the new cert immediately.
    docker compose -f "$INFRA_COMPOSE_FILE" restart glk-registry 2>/dev/null || true
  else
    echo "> Reusing existing registry TLS certificates in $DIR_REGISTRY_TLS."
  fi

  # Add the CA to the OS trust store so Go tools (skaffold, helm) trust the registry.
  if [[ "$(uname -s)" == "Darwin" ]]; then
    if ! security find-certificate -c "$REGISTRY_CA_CN" ~/Library/Keychains/login.keychain-db >/dev/null 2>&1; then
      echo "> Adding registry CA to macOS login keychain (run 'make cleanup-local-registry-ca' to undo); you may be prompted for your login password..."
      security add-trusted-cert -d -r trustRoot \
        -k ~/Library/Keychains/login.keychain-db "$ca_crt"
    fi
  elif [[ "$(uname -s)" == "Linux" ]]; then
    if [[ -n "$SUDO" ]]; then
      echo "> Installing the registry CA into the system trust store may prompt for your sudo password... (run 'make cleanup-local-registry-ca' to undo)"
    fi
    if command -v update-ca-certificates >/dev/null 2>&1; then
      local trust_anchor="/usr/local/share/ca-certificates/${REGISTRY_CA_FILE_BASE}.crt"
      if ! cmp -s "$ca_crt" "$trust_anchor" 2>/dev/null; then
        echo "> Installing registry CA into $trust_anchor..."
        ${SUDO}cp "$ca_crt" "$trust_anchor"
        ${SUDO}update-ca-certificates >/dev/null
      fi
    elif command -v update-ca-trust >/dev/null 2>&1; then
      local trust_anchor="/etc/pki/ca-trust/source/anchors/${REGISTRY_CA_FILE_BASE}.pem"
      if ! cmp -s "$ca_crt" "$trust_anchor" 2>/dev/null; then
        echo "> Installing registry CA into $trust_anchor..."
        ${SUDO}cp "$ca_crt" "$trust_anchor"
        ${SUDO}update-ca-trust extract >/dev/null
      fi
    else
      echo "> WARNING: no supported CA trust tool found; skaffold/helm may not trust the registry."
    fi
  fi
  # Install the CA into any running kind nodes so containerd can pull from the HTTPS registry.
  # kind nodes are Debian-based, so /usr/local/share/ca-certificates + update-ca-certificates is the right path.
  for node in $(docker ps --filter "label=io.x-k8s.kind.cluster" --format "{{.Names}}" 2>/dev/null); do
    if ! docker exec "$node" grep -q "$REGISTRY_CA_CN" /usr/local/share/ca-certificates/${REGISTRY_CA_FILE_BASE}.crt 2>/dev/null; then
      echo "> Installing registry CA into kind node $node..."
      docker cp "$ca_crt" "$node:/usr/local/share/ca-certificates/${REGISTRY_CA_FILE_BASE}.crt"
      docker exec "$node" update-ca-certificates >/dev/null
      docker exec "$node" systemctl restart containerd 2>/dev/null || true
    fi
  done
}

case "$COMMAND" in
  up)
    ensure_registry_tls
    docker compose -f "$INFRA_COMPOSE_FILE" up -d
    ;;
  down)
    docker compose -f "$INFRA_COMPOSE_FILE" down
    ;;
  cleanup-registry-ca)
    cleanup_registry_ca
    ;;
  ensure-registry-tls)
    ensure_registry_tls
    ;;
  *)
    echo "Unknown command: $COMMAND. Valid commands: ${VALID_COMMANDS[*]}" >&2
    exit 1
    ;;
esac
