#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

source $HACK_DIR/tools/mod/aliases.sh;

bash $GARDENER_HACK_DIR/sast.sh $@
