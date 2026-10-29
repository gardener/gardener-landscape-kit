#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail
set -o errexit

glk_ensure_local_gardener_cloud_hosts() {
  if [ -n "${CI:-}" -a -n "${ARTIFACTS:-}" ]; then
    echo "> Adding registry entries to /etc/hosts..."
    printf "\n127.0.0.1 glk-registry.local.gardener.cloud\n" >> /etc/hosts
    printf "\n::1 glk-registry.local.gardener.cloud\n" >> /etc/hosts
    echo "> Content of '/etc/hosts' after adding local.gardener.cloud entries:\n$(cat /etc/hosts)"
  fi
}

# In CI, ensure the GLK registry CA is trusted by the system so that skaffold and helm can push to the
# HTTPS registry. The kind-create-cluster.sh normally handles this for local dev; in CI we run as root
# so the trust store update needs no sudo.
glk_ensure_registry_ca_trusted() {
  if [ -n "${CI:-}" -a -n "${ARTIFACTS:-}" ]; then
    "$(dirname "$0")/../dev-setup/infra.sh" ensure-registry-tls
  fi
}

export_forgejo_runner_logs() {
  docker logs glk-forgejo-runner > "${ARTIFACTS}/forgejo-runner.log" || true
}
