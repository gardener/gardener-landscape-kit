// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package kustomization

import (
	"maps"
	"slices"

	"github.com/spf13/afero"
	kustomize "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
)

const (
	// KustomizationFileName is the name of the component Kustomization file.
	KustomizationFileName = "kustomization.yaml"

	// FluxKustomizationFileName is the name of the Flux Kustomization file.
	FluxKustomizationFileName = "flux-kustomization.yaml"

	// FluxSystemRepositoryName is the name of the Flux system repository.
	FluxSystemRepositoryName = "flux-system"

	// OverrideDir is the directory referenced by the Flux Kustomization of a component.
	OverrideDir = "resources"
)

// NewKustomization creates a Kustomization with the given resources.
func NewKustomization(resources []string, patches []kustomize.Patch) *kustomize.Kustomization {
	return &kustomize.Kustomization{
		TypeMeta: kustomize.TypeMeta{
			APIVersion: kustomize.KustomizationVersion,
			Kind:       kustomize.KustomizationKind,
		},
		Resources: resources,
		Patches:   patches,
	}
}

// WriteKustomizationComponent writes the objects and a Kustomization file to the fs.
// The Kustomization file references all other objects.
// The objects map will be modified to include the Kustomization file.
func WriteKustomizationComponent(objects map[string][]byte, baseDir, componentDir string, fs afero.Afero) error {
	kustomization := NewKustomization(slices.Collect(maps.Keys(objects)), nil)
	content, err := yaml.Marshal(kustomization)
	if err != nil {
		return err
	}
	objects[KustomizationFileName] = content
	return files.WriteObjectsToFilesystem(objects, baseDir, componentDir, fs)
}
