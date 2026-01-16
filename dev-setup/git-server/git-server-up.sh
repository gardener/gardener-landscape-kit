#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

SCRIPT_DIR=$(dirname $0)

REGISTRY_HOSTNAME="glk-registry.local.gardener.cloud"

GIT_SERVER_HOSTNAME="git.local.gardener.cloud"
GIT_SERVER_URL="http://$GIT_SERVER_HOSTNAME:6080"
USER="test:testtest"
REPO_NAMES="base test-landscape"

check_local_dns_record() {
  local local_registry_ip_address=""
  local hostname="$1"

  if [[ "$OSTYPE" == "darwin"* ]]; then
    # Suppress exit code using "|| true"
    local_registry_ip_address=$(dscacheutil -q host -a name $hostname | grep "ip_address" | head -n 1| cut -d' ' -f2 || true)
  elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Suppress exit code using "|| true"
    local_registry_ip_address="$(getent ahosts $hostname || true)"
  else
    echo "Warning: Unknown OS. Make sure $hostname resolves to 127.0.0.1"
    return 0
  fi

  if ! echo "$local_registry_ip_address" | grep -q "127.0.0.1" ; then
      echo "Error: $hostname does not resolve to 127.0.0.1. Please add a line for it in /etc/hosts"
      echo "Command output: $local_registry_ip_address"
      echo "Content of '/etc/hosts':"
      cat /etc/hosts
      exit 1
  fi
}

check_local_dns_records() {
  check_local_dns_record $GIT_SERVER_HOSTNAME
  check_local_dns_record $REGISTRY_HOSTNAME
}

create_repo_call() {
  repo="$1"
  opt="${2:-}"
  curl $opt -H "Content-Type: application/json" \
    -d "{\"name\":\"$repo\"}" \
    -u $USER \
    --fail-with-body \
    -X POST \
    $GIT_SERVER_URL/api/v1/user/repos
}

create_repo() {
  repo="$1"
  if [[ $(curl --silent -u $USER --fail-with-body ${GIT_SERVER_URL}/api/v1/user/repos | yq ".[].name | select(. == \"$repo\")") == "" ]]; then
    echo "Create git repository '$repo'"
    created="false"
    for i in {1..10}; do
      if create_repo_call $repo >/dev/null 2>&1; then
        created="true"
        break
      fi
      sleep 1
    done

    if [[ "$created" != "true" ]]; then
      create_repo_call $repo -v
    fi
  else
    echo "Git repository '$repo' already exists"
  fi
}

check_local_dns_records

cd "$SCRIPT_DIR"
docker compose up -d --build

for i in {1..150}; do
  if curl -u $USER --fail-with-body $GIT_SERVER_URL/api/v1/user >/dev/null 2>&1; then
    break
  fi
  if (( $i % 10 == 9 )); then
    echo "waiting for Forjo startup..."
  fi
  sleep 0.1
done

echo "🚀 git-server is up and running at $GIT_SERVER_URL"

for repo_name in ${REPO_NAMES:-}; do
  create_repo $repo_name
done
