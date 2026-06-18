#!/bin/sh

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -eu

create_user() {
  # Wait until the server is ready
  until curl -sf http://localhost:6080/ >/dev/null 2>&1; do
    echo "Waiting for Forgejo..."
    sleep 1
  done

  if su - git -c "/usr/local/bin/forgejo admin user list" | grep -q 'gitops'; then
    echo "Admin user already exists, skipping creation."
    return
  fi

  # Create default admin user
  su - git -c "/usr/local/bin/forgejo admin user create \
    --username $GIT_USER_NAME \
    --password $GIT_USER_PASSWORD \
    --email gitops@local.gardener.cloud \
    --admin"
}

generate_runner_token() {
  # Check if config already exists with configured TOKEN and UUID
  EXISTING_TOKEN=""
  if [ -f /runner/config.yaml ]; then
    # Extract existing TOKEN from config (if not placeholders)
    EXISTING_TOKEN=$(sed -n 's/^[[:space:]]*token:[[:space:]]*\(.*\)/\1/p' /runner/config.yaml || true)

    # Reset if they are still placeholders
    [ "$EXISTING_TOKEN" = "<TOKEN>" ] && EXISTING_TOKEN=""
  fi

  # Determine TOKEN to use
  if [ -n "$EXISTING_TOKEN" ]; then
    TOKEN="$EXISTING_TOKEN"
    echo "Reusing existing runner token."
  else
    # Generate new runner token
    TOKEN=$(su - git -c "/usr/local/bin/forgejo forgejo-cli actions generate-secret")

    # Register the runner token
    su - git -c "/usr/local/bin/forgejo forgejo-cli actions register --secret $TOKEN --keep-labels"
  fi

  UUID=$(echo -n "$TOKEN" | head -c 16 | xxd -p -c 16 | sed -E 's/(.{8})(.{4})(.{4})(.{4})(.{12})/\1-\2-\3-\4-\5/')
  cp /runner-config.yaml /runner/config.yaml

  sed -i "s|<TOKEN>|$TOKEN|g" /runner/config.yaml
  sed -i "s|<UUID>|$UUID|g" /runner/config.yaml
}

if [ ! -d /data/gitea/conf ]; then
   mkdir -p /data/gitea/conf
fi

if [ ! -f /data/gitea/conf/app.ini ]; then
   cp /app.ini.sample /data/gitea/conf/app.ini
fi

create_user && generate_runner_token &

/usr/bin/entrypoint
