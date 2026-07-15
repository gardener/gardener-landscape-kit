#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

alias flux=$FLUX_CLI

echo "> Generating Flux components"
flux install \
  --export \
  > $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-components.yaml
flux create secret git flux-system \
  --url="https://github.com/<org>/<repo>" \
  --username="<username>" \
  --password="<git_token>" \
  --export \
  > $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/git-sync-secret.yaml
flux create source git flux-system \
  --branch "<branch>" \
  --secret-ref flux-system --url "https://github.com/<org>/<repo>" \
  --recurse-submodules \
  --export \
  > $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml
flux create kustomization flux-system \
  --interval 10m \
  --path "{{ .flux_path }}" \
  --source GitRepository/flux-system \
  --export \
  >> $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml

# Some post processing is needed because the flux CLI does not accept template variables everywhere.

## Replace URL placeholder with Helm template variables
sed -i 's|https://github.com/<org>/<repo>|{{ .repo_url }}|g' $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml

## Replace branch placeholder
sed -i 's|branch\:\s<branch>|{{ .repo_ref }}|g' $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml

## Add comment for recurseSubmodules
sed -i -e 's/recurseSubmodules: true/recurseSubmodules: true # required if Git submodules are used, e.g. to include the base repo in the landscape repo./g' \
  $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml

GOTK=$REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-components.yaml
COMPONENTS=$REPO_ROOT/componentvector/components.yaml
OCM_BASE_COMPONENT=$REPO_ROOT/.ocm/base-component.yaml

# Convert a hyphenated controller name to camelCase key, e.g. "source-controller" -> "sourceController"
to_camel_case() {
  echo "$1" | sed -E 's/-([a-z])/\U\1/g'
}

# Convert a camelCase key to kebab-case, e.g. "fluxCLI" -> "flux-cli", "sourceController" -> "source-controller"
to_kebab_case() {
  echo "$1" | sed 's/\([a-z]\)\([A-Z]\)/\1-\2/g; s/\([A-Z]\+\)\([A-Z][a-z]\)/\1-\2/g' | tr '[:upper:]' '[:lower:]'
}

GLK_INDEX=$(yq '.components | to_entries | .[] | select(.value.name == "github.com/gardener/gardener-landscape-kit") | .key' "$COMPONENTS")

## Extract all Flux controller images from gotk-components.yaml and update downstream files
echo "> Updating Flux controller images in componentvector/components.yaml and .ocm/base-component.yaml"

# Clear existing Flux controller resources so stale entries don't accumulate, preserving non-controller entries (e.g. fluxCLI).
GLK_INDEX=$GLK_INDEX yq -i '.components[env(GLK_INDEX)].resources = (.components[env(GLK_INDEX)].resources | with_entries(select(.key == "fluxCLI")))' "$COMPONENTS"

ocm_resources_file=$(mktemp)
trap "rm -f $ocm_resources_file" EXIT
echo "[]" > "$ocm_resources_file"

# Add Flux controller images to componentvector/components.yaml.
while IFS= read -r image; do
  # image is e.g. "ghcr.io/fluxcd/source-controller:v1.8.5"
  controller="${image#ghcr.io/fluxcd/}"   # "source-controller:v1.8.5"
  controller="${controller%%:*}"           # "source-controller"
  tag="${image##*:}"                       # "v1.8.5"
  repo="${image%:*}"                       # "ghcr.io/fluxcd/source-controller"
  resource_key=$(to_camel_case "$controller")

  GLK_INDEX=$GLK_INDEX resource_key=$resource_key repo=$repo tag=$tag \
    yq -i '.components[env(GLK_INDEX)].resources[env(resource_key)].ociImage.repository = env(repo) |
           .components[env(GLK_INDEX)].resources[env(resource_key)].ociImage.tag = env(tag)' "$COMPONENTS"

  sed -i "s|image: ${image}|image: {{ .resources.${resource_key}.ociImage.ref }}|g" "$GOTK"
done < <(grep -oP "(?<=image: )ghcr\.io/fluxcd/[^\s]+" "$GOTK" | sort -u)

# Add all gardener-landscape-kit resources from componentvector/components.yaml to .ocm/base-component.yaml.
while IFS=$'\t' read -r key repo tag; do
  name=$(to_kebab_case "$key")
  image="${repo}:${tag}"
  name=$name tag=$tag image=$image \
    yq -i '. += [{"name": env(name), "version": env(tag), "type": "ociImage", "relation": "external", "access": {"type": "ociRegistry", "imageReference": env(image)}}]' \
    "$ocm_resources_file"
done < <(GLK_INDEX=$GLK_INDEX yq '.components[env(GLK_INDEX)].resources | to_entries | .[] | .key + "\t" + .value.ociImage.repository + "\t" + .value.ociImage.tag' "$COMPONENTS")

resources_file=$ocm_resources_file yq -i '.resources = load(env(resources_file))' "$OCM_BASE_COMPONENT"
