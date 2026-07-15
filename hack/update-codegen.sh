#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

MODFILE="$(go list -m -f '{{.Dir}}' github.com/gardener/gardener/hack/tools)/go.mod"
GOWORK=off go mod download -modfile "${MODFILE}" k8s.io/code-generator
CODE_GEN_DIR=$(GOWORK=off go list -m -modfile "${MODFILE}" -f '{{.Dir}}' k8s.io/code-generator)
source "${CODE_GEN_DIR}/kube_codegen.sh"

PROJECT_ROOT=$(dirname $0)/..

echo "> Generating API groups"
kube::codegen::gen_helpers \
  --boilerplate "${PROJECT_ROOT}/hack/LICENSE_BOILERPLATE.txt" \
  "${PROJECT_ROOT}/pkg/apis/config"
