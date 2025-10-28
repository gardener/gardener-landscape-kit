// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	"path"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/components/kustomization"
)

const (
	// DirName is the directory name where the cluster instances are stored.
	DirName = "flux"

	// FluxComponentsDirName is the directory name where the flux cli generates the flux-system components into.
	FluxComponentsDirName = DirName + "/flux-system"
)

// GenerateFluxSystemCluster generates the flux-system cluster instance in the given landscape directory.
func GenerateFluxSystemCluster(log logr.Logger, baseDir, landscapeDir string, fs afero.Afero) error {
	instanceFileExisted, err := fs.DirExists(path.Join(landscapeDir, FluxComponentsDirName))
	if err != nil {
		return err
	}
	if !instanceFileExisted {
		logFluxFirstSteps(log, baseDir, landscapeDir)
	}

	return nil
}

func logFluxFirstSteps(log logr.Logger, baseDir, landscapeDir string) {
	fluxDir := path.Join(landscapeDir, DirName)
	landscapePath := kustomization.ComputeBasePath(landscapeDir, baseDir)
	log.Info(`Initialized the landscape for an expected flux cluster at: ` + fluxDir + `

Next steps:
1. Adjust the generated manifests to your environment:

   $  # Directory with initial flux manifests: ` + fluxDir + `

2. Commit and push the changes in your landscape git repository.

3. Install the Flux CLI in your local environment by following the instructions at https://fluxcd.io/flux/installation/#install-the-flux-cli

  $  brew install fluxcd/tap/flux

4  Target the cluster to install flux in:

  $  KUBECONFIG=...

5. Bootstrap the cluster with Flux

  $  flux bootstrap git \
       --url=https://<host>/<org>/<repository> \
       --path=` + path.Join(landscapePath, DirName) + ` \
       --username=<my-username> \
       --password=<my-password> \
       --token-auth=true
`)
}
