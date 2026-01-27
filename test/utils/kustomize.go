// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"path/filepath"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/go-logr/logr"
	"github.com/spf13/afero"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type filesystemAdapter struct {
	afero afero.Afero
}

var _ filesys.FileSystem = &filesystemAdapter{}

func (fsa filesystemAdapter) Create(path string) (filesys.File, error) {
	return fsa.afero.Create(path)
}

func (fsa filesystemAdapter) Mkdir(path string) error {
	return fsa.afero.Mkdir(path, 0755)
}

func (fsa filesystemAdapter) MkdirAll(path string) error {
	return fsa.afero.MkdirAll(path, 0755)
}

func (fsa filesystemAdapter) Open(path string) (filesys.File, error) {
	return fsa.afero.Open(path)
}

func (fsa filesystemAdapter) IsDir(path string) bool {
	b, err := fsa.afero.IsDir(path)
	return err == nil && b
}

func (fsa filesystemAdapter) ReadDir(path string) ([]string, error) {
	fileInfos, err := fsa.afero.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, fi := range fileInfos {
		names = append(names, fi.Name())
	}
	return names, nil
}

func (fsa filesystemAdapter) CleanedAbs(path string) (filesys.ConfirmedDir, string, error) {
	if fsa.IsDir(path) {
		return filesys.ConfirmedDir(path), "", nil
	}
	return filesys.ConfirmedDir(filepath.Dir(path)), filepath.Base(path), nil
}

func (fsa filesystemAdapter) Exists(path string) bool {
	b, err := fsa.afero.Exists(path)
	return err == nil && b
}

func (fsa filesystemAdapter) Glob(pattern string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (fsa filesystemAdapter) RemoveAll(path string) error {
	return fsa.afero.RemoveAll(path)
}

func (fsa filesystemAdapter) ReadFile(path string) ([]byte, error) {
	return fsa.afero.ReadFile(path)
}

func (fsa filesystemAdapter) Walk(path string, walkFn filepath.WalkFunc) error {
	return fsa.afero.Walk(path, walkFn)
}

func (fsa filesystemAdapter) WriteFile(path string, data []byte) error {
	return fsa.afero.WriteFile(path, data, 0644)
}

// KustomizeDir runs kustomize build on the given path using the provided afero filesystem
func KustomizeDir(fs afero.Afero, path string) ([]byte, error) {
	fSys := &filesystemAdapter{afero: fs}
	opts := krusty.MakeDefaultOptions()
	k := krusty.MakeKustomizer(opts)
	resMap, err := k.Run(fSys, path)
	if err != nil {
		return nil, err
	}
	return resMap.AsYaml()
}

// KustomizeComponent generates the base and landscape for the given component and runs kustomize build on the given component path.
func KustomizeComponent(
	fs afero.Afero,
	component components.Interface,
	relativeComponentPath string,
	componentsVectorFile string,
) ([]byte, error) {
	var (
		cmdOpts      = &cmd.Options{Log: logr.Discard()}
		generateOpts = &generateoptions.Options{
			TargetDirPath: "/repo/baseDir",
			Options:       cmdOpts,
			Config: &v1alpha1.LandscapeKitConfiguration{
				Git: &v1alpha1.GitRepository{Paths: v1alpha1.PathConfiguration{Landscape: "./landscapeDir", Base: "./baseDir"}},
			},
		}
	)

	baseOpts, err := components.NewOptions(generateOpts, fs)
	if err != nil {
		return nil, fmt.Errorf("failed to create base options: %w", err)
	}
	if err := component.GenerateBase(baseOpts); err != nil {
		return nil, fmt.Errorf("failed to generate component base: %w", err)
	}

	generateOpts.TargetDirPath = "/repo/landscapeDir"
	generateOpts.Config.VersionConfig = &v1alpha1.VersionConfiguration{
		ComponentsVectorFile: &componentsVectorFile,
	}
	landscapeOpts, err := components.NewLandscapeOptions(generateOpts, fs)
	if err != nil {
		return nil, fmt.Errorf("failed to create landscape options: %w", err)
	}
	if err := component.GenerateLandscape(landscapeOpts); err != nil {
		return nil, fmt.Errorf("failed to generate component landscape: %w", err)
	}

	return KustomizeDir(fs, filepath.Join("/repo/landscapeDir", relativeComponentPath))
}
