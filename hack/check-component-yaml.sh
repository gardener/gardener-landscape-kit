#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

file="componentvector/components.yaml"

# Get total number of components
total=$(yq '.components | length' "$file")

# Check each component that it has a valid sourceRepository field
errors=0
for i in $(seq 0 $((total - 1))); do
    name=$(yq ".components[$i].name" "$file")
    sourceRepo=$(yq ".components[$i].sourceRepository" "$file")

    # Check if sourceRepository exists
    if [ "$sourceRepo" == "null" ]; then
        echo "Error: Component '$name' is missing sourceRepository field"
        ((errors++))
        continue
    fi

    # Check if it starts with https://
    if [[ ! "$sourceRepo" =~ ^https:// ]]; then
        echo "Error: Component '$name' has invalid sourceRepository: $sourceRepo"
        ((errors++))
    fi
done

if [ $errors -eq 0 ]; then
    echo "All $total components have valid sourceRepository fields"
    exit 0
else
    echo "Found $errors error(s)"
    exit 1
fi
