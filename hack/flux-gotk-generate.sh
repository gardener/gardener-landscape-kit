#!/usr/bin/env bash

# SPDX-FileCopyrightText: Contributors to the Gardener project
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
echo "{{- if eq .sourceKind \"GitRepository\" }}" > $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml
flux create source git flux-system \
  --tag "<ref>" \
  --secret-ref flux-system --url "https://github.com/<org>/<repo>" \
  --recurse-submodules \
  --export \
  >> $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml
echo "{{- else if eq .sourceKind \"OCIRepository\" }}" >> $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml
flux create source oci flux-system \
  --tag "<ref>" \
  --secret-ref flux-system --url "https://github.com/<org>/<repo>" \
  --interval=5m \
  --export \
  >> $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml
echo "{{- end }}" >> $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml
flux create kustomization flux-system \
  --interval 10m \
  --path "{{ .flux_path }}" \
  --source GitRepository/flux-system \
  --export \
  | sed -e 's|kind: GitRepository|kind: {{ .sourceKind }}|g' \
  >> $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml \

# Some post processing is needed because the flux CLI does not accept template variables everywhere.

## Replace URL placeholder with Helm template variables
sed -i 's|https://github.com/<org>/<repo>|{{ .source_url }}|g' $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml

## Replace ref placeholder
sed -i 's|tag\:\s<ref>|{{ .source_ref }}|g' $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml

## Add comment for recurseSubmodules
sed -i -e 's/recurseSubmodules: true/recurseSubmodules: true # required if Git submodules are used, e.g. to include the base repo in the landscape repo./g' \
  $REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-sync.yaml

GOTK=$REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-components.yaml
COMPONENTS=$REPO_ROOT/componentvector/components.yaml
OCM_BASE_COMPONENT=$REPO_ROOT/.ocm/base-component.yaml

GLK_INDEX=$(yq '.components | to_entries | .[] | select(.value.name == "github.com/gardener/gardener-landscape-kit") | .key' "$COMPONENTS")

# Convert a hyphenated controller name to camelCase key, e.g. "source-controller" -> "sourceController"
to_camel_case() {
  echo "$1" | sed -E 's/-([a-z])/\U\1/g'
}

## Extract all Flux controller images from gotk-components.yaml and update downstream files
echo "> Updating Flux controller images in componentvector/components.yaml and .ocm/base-component.yaml"

# TODO(LucaBernstein): Remove this pinning guard (and the associative array below) once security-pinned image builds
# (e.g. v1.6.5-27add72a) are no longer needed.
# Snapshot existing pinned tags before the clear step wipes them.
declare -A pinned_tags
while IFS=$'\t' read -r name tag; do
  pinned_tags["$name"]="$tag"
done < <(GLK_INDEX=$GLK_INDEX yq '.components[env(GLK_INDEX)].imageVectorOverwrite.images[] | select(.name != "flux-cli") | .name + "\t" + .tag' "$COMPONENTS")

# Clear existing Flux controller images so stale entries don't accumulate, preserving non-controller entries (e.g. flux-cli).
GLK_INDEX=$GLK_INDEX yq -i '.components[env(GLK_INDEX)].imageVectorOverwrite.images = (.components[env(GLK_INDEX)].imageVectorOverwrite.images | map(select(.name == "flux-cli")))' "$COMPONENTS"

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

  # Always keep the existing pinned tag; warn if the upstream base version changed so a human
  # can manually update the pinned tag in components.yaml and re-run.
  upstream_tag="$tag"
  existing_tag="${pinned_tags[$controller]}"
  if [[ -n "$existing_tag" ]]; then
    existing_base="${existing_tag%%-*}"
    new_base="${tag%%-*}"
    if [[ "$existing_base" != "$new_base" ]]; then
      echo "  ERROR: upstream base version for ${controller} changed (${existing_base} -> ${new_base})." \
           "Update the pinned tag in components.yaml manually, then re-run."
      exit 1
    fi
    tag="$existing_tag"
  fi

  GLK_INDEX=$GLK_INDEX controller=$controller repo=$repo tag=$tag \
    yq -i '.components[env(GLK_INDEX)].imageVectorOverwrite.images += [{"name": env(controller), "repository": env(repo), "tag": env(tag)}]' "$COMPONENTS"

  # Match the plain upstream tag in gotk-components.yaml; replace with the (potentially pinned) tag.
  sed -i "s|image: ${repo}:${upstream_tag}|image: {{ .images.${resource_key}.ref }}|g" "$GOTK"
done < <(grep -oP "(?<=image: )ghcr\.io/fluxcd/[^\s]+" "$GOTK" | sort -u)

# Add all gardener-landscape-kit resources from componentvector/components.yaml to .ocm/base-component.yaml.
while IFS=$'\t' read -r name repo tag; do
  image="${repo}:${tag}"
  name=$name tag=$tag image=$image repo=$repo \
    yq -i '. += [{"name": env(name), "version": env(tag), "type": "ociImage", "relation": "external", "access": {"type": "ociRegistry", "imageReference": env(image)}, "labels": [{"name": "imagevector.gardener.cloud/name", "value": env(name)}, {"name": "imagevector.gardener.cloud/repository", "value": env(repo)}]}]' \
    "$ocm_resources_file"
done < <(GLK_INDEX=$GLK_INDEX yq '.components[env(GLK_INDEX)].imageVectorOverwrite.images[] | .name + "\t" + .repository + "\t" + .tag' "$COMPONENTS")

resources_file=$ocm_resources_file yq -i '.resources = load(env(resources_file))' "$OCM_BASE_COMPONENT"

# Sync flux-cli version to GitHub Action definition in dev-setup, replacing whatever version is currently pinned there.
FLUX_CLI_VERSION=$(yq -e '.components[0].imageVectorOverwrite.images[] | select(.name == "flux-cli") | .tag' $COMPONENTS)
WORKFLOW_FILE="$REPO_ROOT/dev-setup/git-repos/workflow-push-oci.yaml"
sed -i -E "s|(fluxcd/flux2/action@).+|\1$FLUX_CLI_VERSION|g" "$WORKFLOW_FILE"
