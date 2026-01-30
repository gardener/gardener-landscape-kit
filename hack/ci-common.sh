#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail
set -o errexit

glk_ensure_local_gardener_cloud_hosts() {
  if [ -n "${CI:-}" -a -n "${ARTIFACTS:-}" ]; then
    echo "> Adding local.gardener.cloud entries to /etc/hosts..."
    printf "\n127.0.0.1 glk-registry.local.gardener.cloud\n" >> /etc/hosts
    printf "\n::1 glk-registry.local.gardener.cloud\n" >> /etc/hosts
    printf "\n127.0.0.1 git.local.gardener.cloud\n" >> /etc/hosts
    printf "\n::1 git.local.gardener.cloud\n" >> /etc/hosts
    echo "> Content of '/etc/hosts' after adding local.gardener.cloud entries:\n$(cat /etc/hosts)"
  fi
}
