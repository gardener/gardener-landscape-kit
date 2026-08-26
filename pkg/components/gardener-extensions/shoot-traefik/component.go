// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package traefik

import (
	"embed"
	"path"

	"github.com/gardener/gardener/pkg/utils"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
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

	//go:embed meta.yaml
	metadataYAML []byte
)

type component struct {
	*components.Metadata
}

// NewComponent creates a new garden component.
func NewComponent() (components.Interface, error) {
	metadata, err := components.NewMetadata(metadataYAML)
	if err != nil {
		return nil, err
	}
	return &component{Metadata: metadata}, nil
}

// GenerateBase generates the component base directory.
func (c *component) GenerateBase(_ components.Context, options components.Options) error {
	for _, op := range []func(components.Options) error{
		c.writeBaseTemplateFiles,
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

// GenerateLandscape generates the component landscape directory.
func (c *component) GenerateLandscape(_ components.Context, options components.LandscapeOptions) error {
	for _, op := range []func(components.LandscapeOptions) error{
		c.writeLandscapeTemplateFiles,
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

func (c *component) getTemplateValues(opts components.Options) (map[string]any, error) {
	return components.GetComponentVectorTemplateValues(opts, componentvector.NameGardenerGardenerExtensionShootTraefik)
}

func (c *component) writeBaseTemplateFiles(opts components.Options) error {
	objects, err := files.RenderTemplateFiles(baseTemplates, baseTemplateDir, nil)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, c.Directory), opts.GetFilesystem(), opts.GetMergeMode())
}

func (c *component) writeLandscapeTemplateFiles(opts components.LandscapeOptions) error {
	relativeComponentPath := path.Join(components.DirName, c.Directory)

	renderValue, err := c.getTemplateValues(opts)
	if err != nil {
		return err
	}
	values := utils.MergeMaps(renderValue, map[string]any{
		"sourceKind":                  opts.GetSourceKind(),
		"relativePathToBaseComponent": opts.GetRelativeBaseComponentPath(c.Directory),
		"landscapeComponentPath":      path.Join(opts.GetRelativeLandscapePath(), relativeComponentPath),
	})
	objects, err := files.RenderTemplateFiles(landscapeTemplates, landscapeTemplateDir, values)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, c.Directory), opts.GetFilesystem(), opts.GetMergeMode())
}
