#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o nounset
set -o pipefail
set -o errexit

source $GARDENER_HACK_DIR/ci-common.sh

clamp_mss_to_pmtu

ensure_local_gardener_cloud_hosts

# test setup
make kind-up e2e-prepare

# export all container logs and events after test execution
trap "
  ( export_artifacts "glk-single" )
  ( make kind-down )
" EXIT

make test-e2e-local
