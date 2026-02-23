#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

if [[ $# -lt 2 ]]; then
  cat <<USAGE
Usage: $0 <componentName> <componentVersion>

Description:
  Resolves an OCM component descriptor from the public Gardener releases repository
  and outputs the extracted components.yaml file.

Arguments:
  <componentName>     The name of the OCM component to resolve
                      (e.g., github.com/gardener/gardener-extension-provider-aws)
  <componentVersion>  The version of the component to resolve
                      (e.g., v1.68.3)

Output:
  Prints the components.yaml file extracted from the public OCM component descriptor
  to stdout. This file contains the resolved component references and dependencies.

Example:
  $0 github.com/gardener/gardener-extension-provider-aws v1.68.3

Notes:
  - Fetches component descriptors from: oci://europe-docker.pkg.dev/gardener-project/releases
  - Uses temporary directory for intermediate files (automatically cleaned up)
  - Runs 'resolve ocm' command with --ignore-missing-components flag
USAGE
  exit 1
fi

componentName="$1"
componentVersion="$2"

# make temp dir for the resolved components
tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

cat <<EOF > "$tmp_dir/landscapekitconfiguration.yaml"
apiVersion: landscape.config.gardener.cloud/v1alpha1
kind: LandscapeKitConfiguration
ocm:
  repositories:
    - oci://europe-docker.pkg.dev/gardener-project/releases
  rootComponent:
    name: $componentName
    version: $componentVersion
EOF

# Fake component as custom OCM component, to be resolved by the landscapekit and written to the components vector file.
echo $componentName > "$tmp_dir/ocm-component-name"

go run ./cmd/gardener-landscape-kit resolve ocm -c "$tmp_dir/landscapekitconfiguration.yaml" -d "$tmp_dir" --ignore-missing-components > /dev/null

cat "$tmp_dir/components.yaml"
