// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/kustomization"
)

// writeLandscapeComponentsKustomizations traverses through the generated components directory and adds
// Kustomize kustomization.yaml files for each level until each component leaf node containing a Flux Kustomization is reached.
func writeLandscapeComponentsKustomizations(options Options) error {
	fs := options.GetFilesystem()
	targetDir := options.GetTargetPath()
	componentsDir := filepath.Join(targetDir, DirName)

	return fs.Walk(componentsDir, writeKustomizationsToFileTree(fs, targetDir))
}

func writeKustomizationsToFileTree(fs afero.Afero, targetDir string) func(dir string, info os.FileInfo, err error) error {
	var completedPaths []string

	return func(dir string, info os.FileInfo, err error) error {
		if !info.IsDir() || err != nil {
			return err
		}

		for _, p := range completedPaths {
			if isCompleted, err := path.Match(p, dir); err != nil || isCompleted {
				return err
			}
		}

		exists, err := fs.Exists(path.Join(dir, kustomization.FluxKustomizationFileName))
		if err != nil {
			return err
		}
		if exists {
			completedPaths = append(completedPaths, dir+"/*")
			return nil
		}

		subDirs, err := fs.ReadDir(dir)
		if err != nil {
			return err
		}
		var directories []string
		for _, subDir := range subDirs {
			if subDir.IsDir() {
				exists, err := fs.Exists(path.Join(dir, subDir.Name(), kustomization.FluxKustomizationFileName))
				if err != nil {
					return err
				}
				if exists {
					directories = append(directories, path.Join(subDir.Name(), kustomization.FluxKustomizationFileName))
					completedPaths = append(completedPaths, path.Join(dir, subDir.Name(), "*"))
				} else {
					directories = append(directories, subDir.Name())
				}
			}
		}

		relativePath, _ := strings.CutPrefix(dir, targetDir)
		return writeKustomizationFile(fs, targetDir, relativePath, directories)
	}
}

func writeKustomizationFile(fs afero.Afero, landscapeDir, relativePath string, directories []string) error {
	var (
		err     error
		objects = make(map[string][]byte)
	)

	objects[kustomization.KustomizationFileName], err = yaml.Marshal(kustomization.NewKustomization(directories, nil))
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, landscapeDir, relativePath, fs)
}
