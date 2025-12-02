#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

alias flux=$FLUX_CLI

echo "> Generating Flux components"
flux install \
  --export \
  > pkg/components/flux/templates/landscape/gotk-components.yaml
flux create secret git flux-system \
  --url="https://github.com/<org>/<repo>" \
  --username="<username>" \
  --password="<git_token>" \
  --export \
  > pkg/components/flux/templates/landscape/git-sync-secret.yaml
flux create source git flux-system \
  --branch "<branch_name>" \
  --secret-ref flux-system --url "https://github.com/<org>/<repo>" \
  --export \
  > pkg/components/flux/templates/landscape/gotk-sync.yaml
flux create kustomization flux-system \
  --interval 10m \
  --path "<landscape_path_to_flux>" \
  --source GitRepository/flux-system \
  --export \
  >> pkg/components/flux/templates/landscape/gotk-sync.yaml
