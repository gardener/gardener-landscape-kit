// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operator

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/Masterminds/sprig/v3"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/components/flux"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/kustomization"
)

const (
	// ComponentName is the name of the gardener-operator component.
	ComponentName = "gardener-operator"
	// ComponentDirectory is the directory of the gardener-operator component within the base components directory.
	ComponentDirectory = "gardener/operator"

	// HelmReleaseFileName is the name of the helm release manifest file.
	HelmReleaseFileName = "helm-release.yaml"
	// OciRepoFileName is the name of the oci repository manifest file.
	OciRepoFileName = "oci-repository.yaml"

	templateSuffix = ".tpl"
)

var (
	//go:embed version
	fallbackComponentVersion string

	// baseTemplateDir is the directory where the base templates are stored.
	baseTemplateDir = "templates/base"
	//go:embed templates/base
	baseTemplates embed.FS
)

type component struct{}

// NewComponent creates a new gardener-operator component.
func NewComponent() components.Interface {
	return &component{}
}

// Name returns the component name.
func (c *component) Name() string {
	return ComponentName
}

// GenerateBase generates the component base directory.
func (c *component) GenerateBase(options components.Options) error {
	for _, op := range []func(components.Options) error{
		writeTemplateFilesAndKustomization,
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

// GenerateLandscape generates the component landscape directory.
func (c *component) GenerateLandscape(options components.LandscapeOptions) error {
	for _, op := range []func(components.LandscapeOptions) error{
		writeResourcesKustomization,
		writeFluxKustomization,
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

func writeTemplateFilesAndKustomization(opts components.Options) error {
	var objects = make(map[string][]byte)

	dir, err := baseTemplates.ReadDir(baseTemplateDir)
	if err != nil {
		return err
	}
	for _, file := range dir {
		fileName := file.Name()
		fileContents, err := baseTemplates.ReadFile(path.Join(baseTemplateDir, fileName))
		if err != nil {
			return err
		}

		switch fileName {
		case HelmReleaseFileName + templateSuffix:
			if fileContents, fileName, err = renderTemplate(opts, fileContents, fileName); err != nil {
				return err
			}
		case OciRepoFileName + templateSuffix:
			if fileContents, fileName, err = renderTemplate(opts, fileContents, fileName); err != nil {
				return err
			}
		}
		objects[fileName] = fileContents
	}

	return kustomization.WriteKustomizationComponent(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem())
}

func renderTemplate(_ components.Options, fileContents []byte, fileName string) ([]byte, string, error) {
	fileTemplate, err := template.New(fileName).Funcs(sprig.TxtFuncMap()).Parse(string(fileContents))
	if err != nil {
		return nil, "", fmt.Errorf("error parsing template '%s': %w", fileName, err)
	}

	var templatedResult bytes.Buffer
	if err := fileTemplate.Execute(&templatedResult, map[string]any{
		"version": fallbackComponentVersion, // TODO(LucaBernstein): Get actual version from the versions handling component once available.
	}); err != nil {
		return nil, "", fmt.Errorf("error executing '%s' template: %w", fileName, err)
	}

	fileContents = templatedResult.Bytes()
	fileName = strings.TrimSuffix(fileName, templateSuffix)
	return fileContents, fileName, nil
}

func writeResourcesKustomization(options components.LandscapeOptions) error {
	var (
		err error

		objects = make(map[string][]byte)
	)

	relativeOverrideDir := path.Join(components.DirName, ComponentDirectory)
	relativeRoot := files.RelativePathFromDirDepth(path.Join(options.GetRelativeLandscapePath(), relativeOverrideDir))
	objects[kustomization.KustomizationFileName], err = yaml.Marshal(kustomization.NewKustomization([]string{
		path.Join(relativeRoot, options.GetRelativeBasePath(), relativeOverrideDir),
	}, nil))
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, options.GetTargetPath(), relativeOverrideDir, options.GetFilesystem())
}

func writeFluxKustomization(options components.LandscapeOptions) error {
	fluxKustomization, err := yaml.Marshal(&kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kustomizev1.GroupVersion.String(),
			Kind:       kustomizev1.KustomizationKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ComponentName,
			Namespace: flux.GardenNamespaceName,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: flux.SourceRef,
			Path:      path.Join(options.GetRelativeLandscapePath(), components.DirName, ComponentDirectory),
		},
	})
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(
		map[string][]byte{
			kustomization.FluxKustomizationFileName: fluxKustomization,
		},
		options.GetTargetPath(),
		path.Join(components.DirName, ComponentDirectory),
		options.GetFilesystem(),
	)
}
