#!/usr/bin/env bash

export REPO_ROOT="$(readlink -f "$(dirname ${0})/../..")"

if [ -z "$GLK_KIND_CLUSTER_PREFIX" ]; then
  export GLK_KIND_CLUSTER_PREFIX=glk
fi

export GLK_CLUSTER_NAME=$GLK_KIND_CLUSTER_PREFIX-single
export GLK_KUBECONFIG=${REPO_ROOT}/dev/kind-$GLK_CLUSTER_NAME-kubeconfig.yaml
