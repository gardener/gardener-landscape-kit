#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

gosec() {
  go tool -modfile "$HACK_DIR/tools/mod/go.mod" gosec "$@"
}
export -f gosec
