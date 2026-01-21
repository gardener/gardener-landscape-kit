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

  touch /data/.glk-initialisation-finished
}

if [ ! -f /data/.glk-initialisation-finished ]; then
  create_user &
else
  echo "glk initialisation skipped"
fi

/usr/bin/entrypoint
