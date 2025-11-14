// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/kustomization"
)

func writeLandscapeComponentsKustomizations(options Options) error {
	baseComponentsDir := filepath.Join(options.GetLandscapeDir(), DirName)
	fs := options.GetFilesystem()

	var completedPaths []string

	return fs.Walk(baseComponentsDir, func(dir string, info os.FileInfo, err error) error {
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
		objects := make(map[string][]byte)
		objects[kustomization.KustomizationFileName], err = yaml.Marshal(kustomization.NewKustomization(directories, nil))
		if err != nil {
			return err
		}
		relativePath, _ := strings.CutPrefix(dir, options.GetLandscapeDir())
		return files.WriteObjectsToFilesystem(objects, options.GetLandscapeDir(), relativePath, options.GetFilesystem())
	})
}
