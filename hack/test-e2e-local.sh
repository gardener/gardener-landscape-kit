#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

source "$REPO_ROOT/dev-setup/kind/common.sh"
export KUBECONFIG="$GLK_KUBECONFIG"

bash "${GARDENER_HACK_DIR}/test-e2e-local.sh" "$@"
