#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail

clusterNameSuffix=$1

source $(dirname ${0})/common.sh

kind delete cluster --name $GLK_CLUSTER_NAME

rm -f $GLK_KUBECONFIG
