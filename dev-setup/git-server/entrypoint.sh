#!/bin/sh

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

if [ ! -d /data/gitea/conf ]; then
   mkdir -p /data/gitea/conf
fi

if [ ! -f /data/gitea/conf/app.ini ]; then
   cp /app.ini.sample /data/gitea/conf/app.ini
fi

create_user() {
  # Wait until the server is ready
  until curl -sf http://localhost:6080/ >/dev/null 2>&1; do
    echo "Waiting for Forgejo..."
    sleep 1
  done

  # Create default admin user
  su - git -c "/usr/local/bin/forgejo admin user create \
    --username gitops \
    --password testtest \
    --email gitops@local.gardener.cloud \
    --admin"
}

generate_runner_token() {
  # Generate runner token
  TOKEN=$(su - git -c "/usr/local/bin/forgejo forgejo-cli actions generate-secret")

  # Register the runner token
  su - git -c "/usr/local/bin/forgejo forgejo-cli actions register --secret $TOKEN --keep-labels"

  # Generate UUID from token (first 16 bytes)
  UUID=$(echo -n "$TOKEN" | head -c 16 | xxd -p -c 16 | sed -E 's/(.{8})(.{4})(.{4})(.{4})(.{12})/\1-\2-\3-\4-\5/')

  # Create runner directory and copy config
  mkdir -p /data/runner
  cp /runner-config.yaml /data/runner/config.yaml

  # Replace placeholders in config
  sed -i "s|<TOKEN>|$TOKEN|g" /data/runner/config.yaml
  sed -i "s|<UUID>|$UUID|g" /data/runner/config.yaml
}

if [ ! -f /data/.glk-initialisation-finished ]; then
  create_user && generate_runner_token && touch /data/.glk-initialisation-finished &
else
  echo "glk initialisation skipped"
fi

/usr/bin/entrypoint
