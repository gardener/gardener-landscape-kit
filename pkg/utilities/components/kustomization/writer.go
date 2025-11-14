// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package kustomization

import (
	"maps"
	"os"
	"path"
	"slices"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/spf13/afero"
	kustomize "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/components/meta"
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

	// GLKSystemDirName is the name of the directory that contains system files for gardener-landscape-kit.
	GLKSystemDirName = ".glk"

	// DefaultDirName is the name of the directory within the GLK system directory that contains the default generated configuration files.
	DefaultDirName = "defaults"
)

// CreateKustomization creates a Kustomization with the given resources.
func CreateKustomization(resources []string, patches []kustomize.Patch) *kustomize.Kustomization {
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
	kustomization := CreateKustomization(slices.Collect(maps.Keys(objects)), nil)
	content, err := yaml.Marshal(kustomization)
	if err != nil {
		return err
	}
	objects[KustomizationFileName] = content
	return WriteObjectsToFilesystem(objects, baseDir, componentDir, fs)
}

// WriteFluxKustomization writes the Flux Kustomization for components to the fs.
func WriteFluxKustomization(kustomization *kustomizev1.Kustomization, fileName, baseDir, componentDir string, fs afero.Afero) error {
	content, err := yaml.Marshal(kustomization)
	if err != nil {
		return err
	}
	objects := map[string][]byte{fileName: content}

	return WriteObjectsToFilesystem(objects, baseDir, componentDir, fs)
}

// WriteObjectsToFilesystem writes the given objects to the filesystem at the specified baseDir and filePathDir.
// If the manifest file already exists, it patches changes from the new default.
// Additionally, it maintains a default version of the manifest in a separate directory for future diff checks.
func WriteObjectsToFilesystem(objects map[string][]byte, baseDir, filePathDir string, fs afero.Afero) error {
	if err := fs.MkdirAll(path.Join(baseDir, filePathDir), 0700); err != nil {
		return err
	}

	for fileName, object := range objects {
		filePath := path.Join(filePathDir, fileName)

		filePathCurrent := path.Join(baseDir, filePath)
		currentYaml, err := fs.ReadFile(filePathCurrent)
		isCurrentNotExistsErr := os.IsNotExist(err)
		if err != nil && !isCurrentNotExistsErr {
			return err
		}

		filePathDefault := path.Join(baseDir, GLKSystemDirName, DefaultDirName, filePath)
		oldDefaultYaml, err := fs.ReadFile(filePathDefault)
		isDefaultNotExistsErr := os.IsNotExist(err)
		if err != nil && !isDefaultNotExistsErr {
			return err
		}

		if !isDefaultNotExistsErr && len(oldDefaultYaml) > 0 && isCurrentNotExistsErr {
			// File has been deleted by the user. Do not recreate until the default file within the .glk directory is deleted.
			continue
		}

		for _, dir := range []string{
			filePathCurrent,
			filePathDefault,
		} {
			if err := fs.MkdirAll(path.Dir(dir), 0700); err != nil {
				return err
			}
		}
		output, err := meta.ThreeWayMergeManifest(oldDefaultYaml, object, currentYaml)
		if err != nil {
			return err
		}
		// write new manifest
		if err := WriteFileToFilesystem(output, filePathCurrent, true, fs); err != nil {
			return err
		}
		// write new default
		if err := WriteFileToFilesystem(object, filePathDefault, true, fs); err != nil {
			return err
		}
	}

	return nil
}

// WriteFileToFilesystem writes the given file to the filesystem at the specified baseDir and filePathDir.
// If overwriteExisting is false and the file already exists, it does nothing.
func WriteFileToFilesystem(contents []byte, filePathDir string, overwriteExisting bool, fs afero.Afero) error {
	exists, err := fs.Exists(filePathDir)
	if err != nil {
		return err
	}
	if !exists || overwriteExisting {
		if err := fs.MkdirAll(path.Dir(filePathDir), 0700); err != nil {
			return err
		}
		return fs.WriteFile(filePathDir, contents, 0600)
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
