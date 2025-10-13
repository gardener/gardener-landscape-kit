// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	_ "embed"

	v1 "github.com/controlplaneio-fluxcd/flux-operator/api/v1"
	"sigs.k8s.io/yaml"
)

const (
	// FluxOperatorFileName is the name of the flux operator resource set manifest file.
	FluxOperatorFileName = "flux-operator.yaml"
)

var (
	//go:embed manifests/flux-operator.yaml
	fluxOperatorResourceSet []byte
)

func generateFluxOperatorAutoUpgradeResource() (*v1.ResourceSet, error) {
	var resourceSet *v1.ResourceSet
	err := yaml.Unmarshal(fluxOperatorResourceSet, &resourceSet)
	return resourceSet, err
}
