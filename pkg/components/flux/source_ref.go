// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux

import (
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev2 "github.com/fluxcd/source-controller/api/v1"
)

const (
	// FluxSystemRepositoryName is the name of the Flux system repository.
	FluxSystemRepositoryName = "flux-system"

	// FluxSystemNamespaceName is the name of the namespace used by Flux components.
	FluxSystemNamespaceName = "flux-system"
)

// SourceRef is the reference to the repository containing the Flux installation and manifests.
var SourceRef = kustomizev1.CrossNamespaceSourceReference{
	Kind:      sourcev2.GitRepositoryKind,
	Name:      FluxSystemRepositoryName,
	Namespace: FluxSystemNamespaceName,
}
