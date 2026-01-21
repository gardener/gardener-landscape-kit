#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o pipefail

source $(dirname ${0})/common.sh

devDir="${REPO_ROOT}/dev"

ensure_gardener_dir() {
  gardenerVersion=$(go list -m -f "{{.Version}}" github.com/gardener/gardener)

  cd "${devDir}"
  existingVersion=""
  if [ -f gardener/local/VERSION ]; then
    existingVersion=$(cat gardener/local/VERSION || echo "")
  fi

  if [ "$existingVersion" == "$gardenerVersion" ]; then
    echo "✅ Gardener repository already at version: ${gardenerVersion}, skipping copying."
  else
    rm -rf gardener
    echo "💾 Copying Gardener version: ${gardenerVersion}"
    cp -r  $(go list -m -f "{{.Dir}}" github.com/gardener/gardener) gardener
    chmod -R u+w gardener
    chmod u+x gardener/hack/*.sh
    mkdir gardener/local
    echo "$gardenerVersion" > gardener/local/VERSION
  fi
}

skaffold_build_and_push_provider_local() {
  cd "${devDir}/gardener"
  BUILD_DATE=$(date '+%Y-%m-%dT%H:%M:%S%z' | sed 's/\([0-9][0-9]\)$$/:\1/g')
  export LD_FLAGS=$("${devDir}/gardener/hack/get-build-ld-flags.sh" k8s.io/component-base ${devDir}/gardener/VERSION Gardener $BUILD_DATE)
  # speed-up skaffold deployments by building all images concurrently
  export SKAFFOLD_BUILD_CONCURRENCY=0
  # build the images for the platform matching the nodes of the active kubernetes cluster, even in `skaffold build`, which doesn't enable this by default
  export SKAFFOLD_CHECK_CLUSTER_NODE_PLATFORMS=true
  export SKAFFOLD_DEFAULT_REPO=glk-registry.local.gardener.cloud:6001
  export SKAFFOLD_PUSH=true
  export SOURCE_DATE_EPOCH=$(date -d $BUILD_DATE +%s)
  export GARDENER_VERSION=$(cat VERSION)
  sed "s/- registry.local.gardener.cloud:5001/- glk-registry.local.gardener.cloud:6001/g" skaffold-operator.yaml > skaffold-operator-patched.yaml

  skaffold build -f skaffold-operator-patched.yaml -m provider-local --file-output=local/build-output.json
}

generate_extension_yaml() {
  tmpDir="${devDir}/gardener/local/provider-local"
  rm -rf "${tmpDir}"
  mkdir -p "${tmpDir}"
  cp -r "${devDir}/gardener/dev-setup/extensions/provider-local/components/extension/" "${tmpDir}"
  patch_file="${tmpDir}/extension/patch-extension-prow.yaml"
  cat <<EOF > "$patch_file"
  apiVersion: operator.gardener.cloud/v1alpha1
  kind: Extension
  metadata:
    name: provider-local
  spec:
    deployment:
      extension:
        values: {}
EOF
  kubectl kustomize "${tmpDir}/extension" > ${tmpDir}/extension.yaml

  declare -A dict
  dict["local-skaffold/gardener-extension-provider-local/charts/extension"]=":v0.0.0"
  dict["local-skaffold/gardener-extension-admission-local/charts/runtime"]=":v0.0.0"
  dict["local-skaffold/gardener-extension-admission-local/charts/application"]=":v0.0.0"
  dict["local-skaffold/machine-controller-manager-provider-local"]=""

  for v in "${!dict[@]}"
  do
    suffix=${dict[$v]}
    ref=$(yq -r ".builds[] | select(.imageName == \"$v\") | .tag" "${devDir}/gardener/local/build-output.json")
    yq eval --inplace "(.. | select(. == \"$v$suffix\")) = \"$ref\"" ${tmpDir}/extension.yaml
  done

  # patch extension.yaml for usage in kind cluster with different registry port
  sed -i "s/:6001/:5001/g" ${tmpDir}/extension.yaml
  yq eval --inplace '.spec.deployment.admission.values.image = (.spec.deployment.admission.runtimeCluster.helm.ociRepository.ref | sub("_charts_runtime","") | sub("@.+","")) ' ${tmpDir}/extension.yaml
  yq eval --inplace '.spec.deployment.extension.runtimeClusterValues.image = (.spec.deployment.extension.helm.ociRepository.ref | sub("_charts_extension","") | sub("@.+","")) ' ${tmpDir}/extension.yaml
  yq eval --inplace '.spec.deployment.extension.values.image = (.spec.deployment.extension.helm.ociRepository.ref | sub("_charts_extension","") | sub("@.+","")) ' ${tmpDir}/extension.yaml

  echo "✅ Generated extension.yaml for provider-local"
}

update_component() {
  componentDir="${devDir}/e2e/test-landscape/components/provider-local"
  mkdir -p "${componentDir}"
  cp "${devDir}/gardener/local/provider-local/extension.yaml" "${componentDir}"

  cat <<EOF > "${componentDir}/kustomization.yaml"
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- extension.yaml
EOF

  cat <<EOF > "${componentDir}/flux-kustomization.yaml"
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: provider-local
  namespace: garden
spec:
  interval: 30m
  path: components/provider-local
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
  dependsOn:
  - name: gardener-operator
    namespace: garden
EOF

  glk generate landscape -c "${WORK_DIR}/landscapekitconfiguration.yaml" "${WORK_DIR}/test-landscape"

  cd "${WORK_DIR}/test-landscape"
  git add components/provider-local components/kustomization.yaml
  git commit -m "Update provider-local" || echo "No changes to commit"
  git push
  echo "✅ Updated component provider-local"
}

ensure_gardener_dir
skaffold_build_and_push_provider_local
generate_extension_yaml
update_component
