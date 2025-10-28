// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package kustomization

import (
	"maps"
	"path"
	"slices"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/spf13/afero"
	kustomize "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/meta"
)

const (
	// KustomizationFileName is the name of the component kustomization file.
	KustomizationFileName = "kustomization.yaml"
	// FluxSystemRepositoryName is the name of the flux system repository.
	FluxSystemRepositoryName = "flux-system"
)

// WriteKustomizationComponent writes the objects and a Kustomization file to the fs.
func WriteKustomizationComponent(objects map[string]interface{}, baseDir, componentDir string, fs afero.Afero) error {
	objects[KustomizationFileName] = &kustomize.Kustomization{
		TypeMeta: kustomize.TypeMeta{
			APIVersion: kustomize.KustomizationVersion,
			Kind:       kustomize.KustomizationKind,
		},
		Resources: slices.Collect(maps.Keys(objects)),
	}
	return WriteObjectsToFilesystem(objects, baseDir, componentDir, fs)
}

// WriteFluxKustomization writes the flux kustomization for components to the fs.
func WriteFluxKustomization(kustomization *kustomizev1.Kustomization, fileName, baseDir, componentDir string, fs afero.Afero) error {
	objects := map[string]any{fileName: kustomization}

	return WriteObjectsToFilesystem(objects, baseDir, componentDir, fs)
}

// WriteObjectsToFilesystem writes the given objects to the filesystem at the specified baseDir and filePathDir.
func WriteObjectsToFilesystem(objects map[string]interface{}, baseDir, filePathDir string, fs afero.Afero) error {
	if err := fs.MkdirAll(path.Join(baseDir, filePathDir), 0700); err != nil {
		return err
	}

	for fileName, object := range objects {
		filePath := path.Join(filePathDir, fileName)
		switch object := object.(type) {
		case []byte:
			if err := meta.CreateOrUpdateManifest(object, baseDir, filePath, fs); err != nil {
				return err
			}
		default:
			file, err := yaml.Marshal(object)
			if err != nil {
				return err
			}
			if err := meta.CreateOrUpdateManifest(file, baseDir, filePath, fs); err != nil {
				return err
			}
		}
	}

	return nil
}

// ComputeBasePath determines the correct base directory reference for same-repo setups.
func ComputeBasePath(baseDir, landscapeDir string) string {
	landscapePrefix, _ := path.Split(landscapeDir)
	basePrefix, shortBaseDir := path.Split(baseDir)
	if landscapePrefix == basePrefix {
		return shortBaseDir
	}
	return baseDir
}
