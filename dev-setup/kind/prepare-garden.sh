#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail

source $(dirname ${0})/common.sh

gardenComponentDir="${WORK_DIR}/test-landscape/components/gardener/garden"

patch_garden_yaml() {
  echo "🛠️ Patching garden.yaml"
  yq eval-all --inplace 'select(fileIndex == 0) * select(fileIndex == 1)' \
    "${gardenComponentDir}/garden.yaml" \
    <(cat <<EOF
spec:
  dns:
    providers:
    - name: primary
      type: local
      secretRef:
        name: garden-dns-local
  runtimeCluster:
    ingress:
      controller:
        kind: nginx
      domains:
      - name: ingress.runtime-garden.local.gardener.cloud
        provider: primary
    networking:
      ipFamilies:
      - IPv4
      pods:
      - 10.1.0.0/16
      nodes:
      - 172.18.0.0/24
      services:
      - 10.2.0.0/16
  virtualCluster:
    dns:
      domains:
      - name: virtual-garden.local.gardener.cloud
        provider: primary
    gardener:
      clusterIdentity: test-landscape-123456
EOF
)
  prettify "${gardenComponentDir}/garden.yaml"
}

patch_kustomization_yaml() {
  echo "🛠️ Patching kustomization.yaml"
  yq eval-all --inplace 'select(fileIndex == 0) *+ select(fileIndex == 1)' \
    "${gardenComponentDir}/kustomization.yaml" \
    <(cat <<EOF
resources:
- secret-garden-dns-local.yaml
EOF
)
  # Ensure resources are unique after patching using "*+" (deep merge) operator
  yq eval --inplace '.resources |= unique' "${gardenComponentDir}/kustomization.yaml"
  prettify "${gardenComponentDir}/kustomization.yaml"
}

write_secret_garden_dns_local_yaml() {
  echo "📝 Writing secret-garden-dns-local.yaml"
  cat > "${gardenComponentDir}/secret-garden-dns-local.yaml" <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: garden-dns-local
  namespace: garden
type: Opaque
data:
EOF
  prettify "${gardenComponentDir}/secret-garden-dns-local.yaml"
}

commit_changes() {
  echo "💾 Committing changes to garden component"
  cd "${gardenComponentDir}"
  git add garden.yaml kustomization.yaml secret-garden-dns-local.yaml
  git commit -m "Prepare garden component for local provider" || echo "No changes to commit"
  git push
}

patch_garden_yaml
write_secret_garden_dns_local_yaml
patch_kustomization_yaml
commit_changes
