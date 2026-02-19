// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package suse

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
	ComponentDirectory = "gardener-extensions/os-suse-chost"
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
	return "os-suse-chost"
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
	return components.GetRenderValues(opts,
		componentvector.NameGardenerGardenerExtensionOsSuseChost,
		func(version string) map[string]any {
			return map[string]any{
				"osSuseChost": map[string]any{
					"helmChartRef": "europe-docker.pkg.dev/gardener-project/public/charts/gardener/extensions/os-suse-chost:" + version,
				},
			}
		})
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
