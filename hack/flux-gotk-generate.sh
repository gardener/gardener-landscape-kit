#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

alias flux=$FLUX_CLI

echo "> Generating Flux components"
flux install \
  --export \
  > pkg/components/flux/templates/landscape/gotk-sync.yaml.tpl
flux create secret git flux-system \
  --url="https://github.com/<org>/<repo>" \
  --username="<username>" \
  --password="<git_token>" \
  --export \
  > pkg/components/flux/templates/landscape/git-sync-secret.yaml
flux create source git flux-system \
  --branch "<branch>" \
  --secret-ref flux-system --url "https://github.com/<org>/<repo>" \
  --export \
  > pkg/components/flux/templates/landscape/gotk-sync.yaml.tpl
flux create kustomization flux-system \
  --interval 10m \
  --path "{{ .flux_path }}" \
  --source GitRepository/flux-system \
  --export \
  >> pkg/components/flux/templates/landscape/gotk-sync.yaml.tpl

# Some post processing is needed because the flux CLI does not accept template variables everywhere.

## Replace URL placeholder with Helm template variables
sed -i 's|https://github.com/<org>/<repo>|{{ .repo_url }}|g' pkg/components/flux/templates/landscape/gotk-sync.yaml.tpl

## Replace branch placeholder
sed -i 's|branch\:\s<branch>|{{ .repo_ref }}|g' pkg/components/flux/templates/landscape/gotk-sync.yaml.tpl
