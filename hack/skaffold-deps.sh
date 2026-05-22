#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

operation="${1:-check}"

echo "> ${operation} Skaffold Dependencies"

success=true
repo_root="$(git rev-parse --show-toplevel)"

bash "${GARDENER_HACK_DIR}"/check-skaffold-deps-for-binary.sh "$operation" --skaffold-file "dev-setup/skaffold.yaml" --binary "gardener-landscape-kit" --skaffold-config gardener-landscape-kit
