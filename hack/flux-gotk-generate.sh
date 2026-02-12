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
