#!/usr/bin/env bash

# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

source "$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )/../env.sh"

WORK_DIR="$REPO_ROOT/dev/e2e"
mkdir -p "${WORK_DIR}"

glk() {
  gardener-landscape-kit "$@"
}

kubectl_glk() {
  kubectl --kubeconfig "$GLK_KUBECONFIG" "$@"
}

kubectl_glk_apply() {
  kubectl --kubeconfig "$GLK_KUBECONFIG" apply --server-side --force-conflicts "$@"
}
