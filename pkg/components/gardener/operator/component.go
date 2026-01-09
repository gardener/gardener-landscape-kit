// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operator

import (
	"embed"
	"path"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

const (
	// ComponentName is the name of the gardener-operator component.
	ComponentName = "gardener-operator"
	// ComponentDirectory is the directory of the gardener-operator component within the base components directory.
	ComponentDirectory = "gardener/operator"
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

func writeBaseTemplateFiles(opts components.Options) error {
	gardenerVersion, exists := opts.GetComponentVector().FindComponentVersion(componentvector.NameGardenerGardener)
	if !exists {
		opts.GetLogger().Info("Component version not found in component vector, falling back to empty version", "component", componentvector.NameGardenerGardener)
	}

	objects, err := files.RenderTemplateFiles(baseTemplates, baseTemplateDir, map[string]any{
		"version": gardenerVersion,
	})
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

	objects, err := files.RenderTemplateFiles(landscapeTemplates, landscapeTemplateDir, map[string]any{
		"relativePathToBaseComponent": path.Join(relativeRepoRoot, opts.GetRelativeBasePath(), relativeComponentPath),
		"landscapeComponentPath":      path.Join(opts.GetRelativeLandscapePath(), relativeComponentPath),
	})
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem())
}
