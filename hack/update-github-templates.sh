#!/bin/bash
#
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

mkdir -p "$(dirname $0)/../.github" "$(dirname $0)/../.github/ISSUE_TEMPLATE"

for file in `find "${GARDENER_HACK_DIR}"/../.github -name '*.md'`; do
  cat "$file" |\
    sed 's/operating Gardener/working with Gardener-Landscape-Kit/g' |\
    sed 's/to the Gardener project/for Gardener-Landscape-Kit/g' |\
    sed 's/to Gardener/to Gardener-Landscape-Kit/g' |\
    sed 's/- Gardener version:/- Gardener version (if relevant):\n- Gardener-Landscape-Kit version:/g' \
  > "$(dirname $0)/../.github/${file#*.github/}"
done
