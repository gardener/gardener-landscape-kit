// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"embed"
	"path"

	"github.com/gardener/gardener/pkg/utils"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

const (
	// ComponentDirectory is the garden component directory within the base components directory.
	ComponentDirectory = "gardener-extensions/provider-aws"
)

var (
	// baseTemplateDir is the directory where the base templates are stored.
	baseTemplateDir = "templates/base"
	//go:embed templates/base
	baseTemplates embed.FS

	// landscapeTemplateDir is the directory where the landscape templates are stored.
	landscapeTemplateDir = "templates/landscape"
	//go:embed templates/landscape
	landscapeTemplates embed.FS
)

type component struct{}

// NewComponent creates a new garden component.
func NewComponent() components.Interface {
	return &component{}
}

// Name returns the component name.
func (c *component) Name() string {
	return "provider-aws"
}

// GenerateBase generates the component base directory.
func (c *component) GenerateBase(options components.Options) error {
	for _, op := range []func(components.Options) error{
		writeBaseTemplateFiles,
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
		writeLandscapeTemplateFiles,
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

func getRenderValues(opts components.Options) (map[string]any, error) {
	cv := opts.GetComponentVector().FindComponentVector(componentvector.NameGardenerGardenerExtensionProviderAws)
	if cv == nil || len(cv.Resources) == 0 {
		version, exists := opts.GetComponentVector().FindComponentVersion(componentvector.NameGardenerGardenerExtensionProviderAws)
		if !exists {
			opts.GetLogger().Info("Component version not found in component vector, falling back to empty version", "component", componentvector.NameGardenerGardenerExtensionProviderAws)
		}
		return map[string]any{
			"resources": map[string]any{
				"admissionAwsRuntime": map[string]any{
					"helmChart": map[string]any{
						"ref": "europe-docker.pkg.dev/gardener-project/releases/charts/gardener/extensions/admission-aws-runtime:" + version,
					},
				},
				"admissionAwsApplication": map[string]any{
					"helmChart": map[string]any{
						"ref": "europe-docker.pkg.dev/gardener-project/releases/charts/gardener/extensions/admission-aws-application:" + version,
					},
				},
				"providerAws": map[string]any{
					"helmChart": map[string]any{
						"ref": "europe-docker.pkg.dev/gardener-project/releases/charts/gardener/extensions/provider-aws:" + version,
					},
				},
			},
		}, nil
	}
	return cv.TemplateValues()
}

func writeBaseTemplateFiles(opts components.Options) error {
	objects, err := files.RenderTemplateFiles(baseTemplates, baseTemplateDir, nil)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem())
}

func writeLandscapeTemplateFiles(opts components.LandscapeOptions) error {
	var (
		relativeComponentPath = path.Join(components.DirName, ComponentDirectory)
		relativeRepoRoot      = files.CalculatePathToComponentBase(opts.GetRelativeLandscapePath(), relativeComponentPath)
	)

	renderValue, err := getRenderValues(opts)
	if err != nil {
		return err
	}
	values := utils.MergeMaps(renderValue, map[string]any{
		"relativePathToBaseComponent": path.Join(relativeRepoRoot, opts.GetRelativeBasePath(), relativeComponentPath),
		"landscapeComponentPath":      path.Join(opts.GetRelativeLandscapePath(), relativeComponentPath),
	})
	objects, err := files.RenderTemplateFiles(landscapeTemplates, landscapeTemplateDir, values)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem())
}
