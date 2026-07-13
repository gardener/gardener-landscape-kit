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

## Extract controller images from gotk-components.yaml and update componentvector/components.yaml
echo "> Updating Flux controller images in componentvector/components.yaml"

GOTK=$REPO_ROOT/pkg/components/flux/templates/landscape/flux-system/gotk-components.yaml
COMPONENTS=$REPO_ROOT/componentvector/components.yaml

extract_image() {
  grep -oP "(?<=image: )ghcr\.io/fluxcd/${1}:[^\s]+" "$GOTK" | head -1
}

SOURCE_IMAGE=$(extract_image "source-controller")
KUSTOMIZE_IMAGE=$(extract_image "kustomize-controller")
HELM_IMAGE=$(extract_image "helm-controller")
NOTIFICATION_IMAGE=$(extract_image "notification-controller")

GLK_INDEX=$(yq '.components | to_entries | .[] | select(.value.name == "github.com/gardener/gardener-landscape-kit") | .key' "$COMPONENTS")

update_resource() {
  local resource_key=$1
  local image=$2
  local repo="${image%:*}"
  local tag="${image##*:}"
  yq -i ".components[${GLK_INDEX}].resources.${resource_key}.ociImage.repository = \"${repo}\"" "$COMPONENTS"
  yq -i ".components[${GLK_INDEX}].resources.${resource_key}.ociImage.tag = \"${tag}\"" "$COMPONENTS"
}

update_resource "sourceController" "$SOURCE_IMAGE"
update_resource "kustomizeController" "$KUSTOMIZE_IMAGE"
update_resource "helmController" "$HELM_IMAGE"
update_resource "notificationController" "$NOTIFICATION_IMAGE"

## Update Flux controller images in .ocm/base-component.yaml
echo "> Updating Flux controller images in .ocm/base-component.yaml"

BASE_COMPONENT=$REPO_ROOT/.ocm/base-component.yaml

update_ocm_resource() {
  local name=$1
  local image=$2
  local tag="${image##*:}"
  yq -i "(.resources[] | select(.name == \"${name}\") | .version) = \"${tag}\"" "$BASE_COMPONENT"
  yq -i "(.resources[] | select(.name == \"${name}\") | .access.imageReference) = \"${image}\"" "$BASE_COMPONENT"
}

update_ocm_resource "source-controller" "$SOURCE_IMAGE"
update_ocm_resource "kustomize-controller" "$KUSTOMIZE_IMAGE"
update_ocm_resource "helm-controller" "$HELM_IMAGE"
update_ocm_resource "notification-controller" "$NOTIFICATION_IMAGE"

## Replace actual images in gotk-components.yaml with template placeholders
echo "> Replacing images in gotk-components.yaml with template placeholders"

sed -i "s|image: ${SOURCE_IMAGE}|image: {{ .resources.sourceController.ociImage.ref }}|g" "$GOTK"
sed -i "s|image: ${KUSTOMIZE_IMAGE}|image: {{ .resources.kustomizeController.ociImage.ref }}|g" "$GOTK"
sed -i "s|image: ${HELM_IMAGE}|image: {{ .resources.helmController.ociImage.ref }}|g" "$GOTK"
sed -i "s|image: ${NOTIFICATION_IMAGE}|image: {{ .resources.notificationController.ociImage.ref }}|g" "$GOTK"
