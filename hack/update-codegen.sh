#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

CODE_GEN_DIR=$(go list -m -f '{{.Dir}}' k8s.io/code-generator)
source "${CODE_GEN_DIR}/kube_codegen.sh"

rm -f $GOPATH/bin/*-gen

PROJECT_ROOT=$(dirname $0)/..

echo "> Generating API groups"
kube::codegen::gen_helpers \
  --boilerplate "${PROJECT_ROOT}/hack/LICENSE_BOILERPLATE.txt" \
  "${PROJECT_ROOT}/pkg/apis/config"
