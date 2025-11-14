// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux

import (
	"path"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev2 "github.com/fluxcd/source-controller/api/v1"
	"github.com/go-logr/logr"
	"github.com/spf13/afero"
)

const (
	// DirName is the directory name where the cluster instances are stored.
	DirName = "flux"

	// FluxComponentsDirName is the directory name where the Flux cli generates the flux-system components into.
	FluxComponentsDirName = DirName + "/flux-system"

	// FluxSystemRepositoryName is the name of the flux system repository.
	FluxSystemRepositoryName = "flux-system"
)

// SourceRef is the reference to the repository containing the flux installation and manifests.
var SourceRef = kustomizev1.CrossNamespaceSourceReference{
	Kind:      sourcev2.GitRepositoryKind,
	Name:      FluxSystemRepositoryName,
	Namespace: FluxSystemNamespaceName,
}

// GenerateFluxSystemCluster generates the flux-system cluster instance in the given landscape directory.
func GenerateFluxSystemCluster(log logr.Logger, _, landscapeDir string, fs afero.Afero) error {
	instanceFileExisted, err := fs.DirExists(path.Join(landscapeDir, FluxComponentsDirName))
	if err != nil {
		return err
	}
	if !instanceFileExisted {
		logFluxFirstSteps(log, landscapeDir)
	}

	return nil
}

func logFluxFirstSteps(log logr.Logger, landscapeDir string) {
	fluxDir := path.Join(landscapeDir, DirName)
	log.Info(`Initialized the landscape for an expected Flux cluster at: ` + fluxDir + `

Next steps:
1. Adjust the generated manifests to your environment, especially the Git repository reference:

   # Directory with initial flux manifests: ` + fluxDir + `

2. Install the Flux CRDs initially:

   $  kubectl create -f ` + path.Join(landscapeDir, FluxComponentsDirName, "gotk-components.yaml") + `

3. You might want to consider creating the Git sync credentials manually and store them separately instead of checking them into Git:

   $  kubectl create -f ` + path.Join(landscapeDir, FluxComponentsDirName, "git-sync-secret.yaml") + `

2. Commit and push the changes in your landscape git repository.

3. Target the cluster to install Flux in:

  $  KUBECONFIG=...

4. Deploy Flux on the cluster:

  $  kubectl apply --force-conflicts -k ` + path.Join(landscapeDir, FluxComponentsDirName) + `
`)
}
