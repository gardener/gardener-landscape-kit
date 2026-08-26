#!/usr/bin/env bash

# SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

SCRIPT_DIR=$(dirname $0)

cd "$SCRIPT_DIR"
docker compose down
