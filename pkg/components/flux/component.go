// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux

import (
	"embed"
	"path"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/flux"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/kustomization"
)

const (
	// ComponentName is the name of the garden component.
	ComponentName = "components"
)

var (
	//go:embed templates/landscape
	landscapeTemplates embed.FS
)

type component struct{}

// NewComponent creates a new gardener-operator component.
func NewComponent() components.Interface {
	return &component{}
}

// GenerateBase generates the component base directory.
func (c *component) GenerateBase(_ components.Options) error {
	return nil
}

// GenerateLandscape generates the component landscape directory.
func (c *component) GenerateLandscape(options components.Options) error {
	var (
		objects                    = make(map[string][]byte)
		kustomizationObjectEntries []string

		relativeLandscapeDir = files.ComputeBasePath(options.GetLandscapeDir(), options.GetBaseDir())

		landscapeTemplateDir  = "templates/landscape"
		gitignoreTemplateFile = "gitignore"
		gitSecretFileName     = "git-sync-secret.yaml" // This template should be created manually and not checked into the landscape Git repo.
	)

	dir, err := landscapeTemplates.ReadDir(landscapeTemplateDir)
	if err != nil {
		return err
	}
	for _, file := range dir {
		fileName := file.Name()
		if file.IsDir() || fileName == gitignoreTemplateFile {
			continue
		}
		if fileName != gitSecretFileName {
			kustomizationObjectEntries = append(kustomizationObjectEntries, fileName)
		}
		fileContents, err := landscapeTemplates.ReadFile(path.Join(landscapeTemplateDir, fileName))
		if err != nil {
			return err
		}
		objects[fileName] = fileContents
	}

	kustomizationManifest := kustomization.NewKustomization(kustomizationObjectEntries, nil)
	objects[kustomization.KustomizationFileName], err = yaml.Marshal(kustomizationManifest)
	if err != nil {
		return err
	}

	gitignore, err := landscapeTemplates.ReadFile(path.Join(landscapeTemplateDir, gitignoreTemplateFile))
	if err != nil {
		return err
	}
	gitignoreDefaultPath := path.Join(options.GetLandscapeDir(), files.GLKSystemDirName, files.DefaultDirName, flux.FluxComponentsDirName, ".gitignore")
	fileDefaultExists, err := options.GetFilesystem().Exists(gitignoreDefaultPath)
	if err == nil && !fileDefaultExists {
		if err := files.WriteFileToFilesystem(gitignore, path.Join(options.GetLandscapeDir(), flux.FluxComponentsDirName, ".gitignore"), false, options.GetFilesystem()); err != nil {
			return err
		}
	}
	// Write the default gitignore file to the .glk defaults system directory.
	if err := files.WriteFileToFilesystem(gitignore, gitignoreDefaultPath, true, options.GetFilesystem()); err != nil {
		return err
	}

	if err := files.WriteObjectsToFilesystem(objects, options.GetLandscapeDir(), flux.FluxComponentsDirName, options.GetFilesystem()); err != nil {
		return err
	}

	k, err := yaml.Marshal(&kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kustomizev1.GroupVersion.String(),
			Kind:       kustomizev1.KustomizationKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ComponentName,
			Namespace: flux.FluxSystemNamespaceName,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: flux.SourceRef,
			Path:      path.Join(relativeLandscapeDir, components.DirName),
		},
	})
	if err != nil {
		return err
	}
	if err := flux.GenerateGardenerNamespaceManifest(options.GetLandscapeDir(), flux.DirName, options.GetFilesystem()); err != nil {
		return err
	}
	return files.WriteObjectsToFilesystem(
		map[string][]byte{ComponentName + ".yaml": k},
		options.GetLandscapeDir(),
		flux.DirName,
		options.GetFilesystem(),
	)
}
