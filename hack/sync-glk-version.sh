#!/usr/bin/env bash
#
# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# This script syncs the gardener-landscape-kit component version in
# componentvector/components.yaml with the version specified in the VERSION file.
# If the component doesn't exist, it adds it as the first entry.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION_FILE="$REPO_ROOT/VERSION"
COMPONENTS_FILE="$REPO_ROOT/componentvector/components.yaml"

# Read the VERSION file
if [[ ! -f "$VERSION_FILE" ]]; then
  echo "ERROR: VERSION file not found at $VERSION_FILE" >&2
  exit 1
fi

VERSION=$(cat "$VERSION_FILE" | tr -d '[:space:]')

if [[ -z "$VERSION" ]]; then
  echo "ERROR: VERSION file is empty" >&2
  exit 1
fi

# Check if components.yaml exists
if [[ ! -f "$COMPONENTS_FILE" ]]; then
  echo "ERROR: components.yaml not found at $COMPONENTS_FILE" >&2
  exit 1
fi

GLK_COMPONENT_NAME="github.com/gardener/gardener-landscape-kit"
REPO_URL="https://$GLK_COMPONENT_NAME"

# Check if the GLK component already exists in the file
if yq eval ".components[] | select(.name == \"$GLK_COMPONENT_NAME\") | .name" "$COMPONENTS_FILE" | grep -q "$GLK_COMPONENT_NAME"; then
  # Component exists - update its version
  yq eval -i --indent 2 -c "(.components[] | select(.name == \"$GLK_COMPONENT_NAME\") | .version) = \"$VERSION\"" "$COMPONENTS_FILE"
  echo "Updated $GLK_COMPONENT_NAME component version to $VERSION"
else
  # Component doesn't exist - prepend it as the first entry
  yq eval -i --indent 2 -c ".components = [{\"name\": \"$GLK_COMPONENT_NAME\", \"sourceRepository\": \"$REPO_URL\", \"version\": \"$VERSION\"}] + .components" "$COMPONENTS_FILE"
  echo "Added $GLK_COMPONENT_NAME component with version $VERSION as first entry"
fi
