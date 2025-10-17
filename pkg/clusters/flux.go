// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	"path"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/meta"
)

const (
	// DirName is the directory name where the cluster instances are stored.
	DirName = "clusters"

	// RuntimeClusterDirName is the directory where the root Flux Kustomizations are stored.
	RuntimeClusterDirName = DirName + "/runtime"

	// FluxRuntimeClusterDirName is the directory name where the flux-system runtime cluster instance is stored.
	FluxRuntimeClusterDirName = RuntimeClusterDirName + "/flux-system"
)

// GenerateFluxSystemCluster generates the flux-system cluster instance in the given landscape directory.
func GenerateFluxSystemCluster(log logr.Logger, landscapeDir string, fs afero.Afero) error {
	generatedFluxInstance, err := generateFluxInstance()
	if err != nil {
		return err
	}
	fluxInstanceFilePath := path.Join(FluxRuntimeClusterDirName, FluxInstanceFileName)
	instanceFileExisted, err := fs.Exists(path.Join(landscapeDir, fluxInstanceFilePath))
	if err != nil {
		return err
	}
	if err := meta.CreateOrUpdateManifest(generatedFluxInstance, landscapeDir, fluxInstanceFilePath, fs); err != nil {
		return err
	}
	if !instanceFileExisted {
		logFluxFirstSteps(log, landscapeDir, fluxInstanceFilePath)
	}

	upgradeResource, err := generateFluxOperatorAutoUpgradeResource()
	if err != nil {
		return err
	}
	if err = meta.CreateOrUpdateManifest(upgradeResource, landscapeDir, path.Join(FluxRuntimeClusterDirName, FluxOperatorFileName), fs); err != nil {
		return err
	}

	return nil
}

func logFluxFirstSteps(log logr.Logger, landscapeDir, fluxInstanceFilePath string) {
	fluxInstanceFile := path.Join(landscapeDir, fluxInstanceFilePath)
	log.Info(`Generated FluxInstance manifest at ` + fluxInstanceFile + `

Next steps:
1. Ensure that the Flux Operator is installed in your cluster, e.g. via the following Helm command:

   $  helm install flux-operator oci://ghcr.io/controlplaneio-fluxcd/charts/flux-operator \
        --namespace flux-system \
        --create-namespace

   You can find more information about installing the Flux Operator at https://fluxcd.control-plane.io/operator/install/#install-methods

2. Create the sync repo pull secret in the flux-system namespace, e.g. via the following command:

   $  kubectl create secret generic \
        flux-system-git-auth \
        --namespace flux-system \
        --from-literal=username=git \
        --from-literal=password=$GIT_TOKEN

3. Adjust the generated manifest to your environment and apply it to your cluster:

   $  kubectl apply -f ` + fluxInstanceFile + `

4. Commit and push the changes in your landscape git repository.
`)
}
