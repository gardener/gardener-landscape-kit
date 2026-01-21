#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

export REPO_ROOT="$(readlink -f "$(dirname ${0})/../..")"

if [ -z "$GLK_KIND_CLUSTER_PREFIX" ]; then
  export GLK_KIND_CLUSTER_PREFIX=glk
fi

export GLK_CLUSTER_NAME=$GLK_KIND_CLUSTER_PREFIX-single
export GLK_KUBECONFIG=${REPO_ROOT}/dev/kind-$GLK_CLUSTER_NAME-kubeconfig.yaml

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