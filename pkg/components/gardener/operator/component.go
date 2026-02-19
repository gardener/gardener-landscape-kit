// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operator

import (
	"embed"
	"fmt"
	"path"

	"github.com/gardener/gardener/pkg/utils"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/components/helpers"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
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
	objects, err := files.RenderTemplateFiles(baseTemplates, baseTemplateDir, nil)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem())
}

func getRenderValues(opts components.Options) (map[string]any, error) {
	cv := opts.GetComponentVector().FindComponentVector(componentvector.NameGardenerGardener)
	if cv == nil || len(cv.Resources) == 0 {
		gardenerVersion, exists := opts.GetComponentVector().FindComponentVersion(componentvector.NameGardenerGardener)
		if !exists {
			opts.GetLogger().Info("Component version not found in component vector, falling back to empty version", "component", componentvector.NameGardenerGardener)
		}
		return map[string]any{
			"repository": "europe-docker.pkg.dev/gardener-project/releases/charts/gardener/operator",
			"tag":        gardenerVersion,
		}, nil
	}

	ref, err := getOCIImageReferenceFromComponentVector("operator", cv)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator OCI image reference from component vector: %w", err)
	}
	repository, tag, err := helpers.SplitOCIImageReference(ref)
	if err != nil {
		return nil, err
	}

	values, err := cv.TemplateValues()
	if err != nil {
		return nil, fmt.Errorf("failed to get template values from component vector: %w", err)
	}
	values["repository"] = repository
	values["tag"] = tag
	return values, nil
}

func writeLandscapeTemplateFiles(opts components.LandscapeOptions) error {
	var (
		relativeComponentPath = path.Join(components.DirName, ComponentDirectory)
		relativeRepoRoot      = files.CalculatePathToComponentBase(opts.GetRelativeLandscapePath(), relativeComponentPath)
	)

	values, err := getRenderValues(opts)
	if err != nil {
		return err
	}

	values = utils.MergeMaps(values, map[string]any{
		"relativePathToBaseComponent": path.Join(relativeRepoRoot, opts.GetRelativeBasePath(), relativeComponentPath),
		"landscapeComponentPath":      path.Join(opts.GetRelativeLandscapePath(), relativeComponentPath),
	})
	objects, err := files.RenderTemplateFiles(landscapeTemplates, landscapeTemplateDir, values)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem())
}

func getOCIImageReferenceFromComponentVector(name string, cv *utilscomponentvector.ComponentVector) (string, error) {
	if cv == nil || len(cv.Resources) == 0 {
		return "", fmt.Errorf("component vector or component resources are nil")
	}

	data := cv.Resources[name]
	if data == nil {
		return "", fmt.Errorf("no resources found for component %s", name)
	}
	ref := data.OCIImageReference
	if ref == nil {
		return "", fmt.Errorf("OCI image reference not found for component %s", name)
	}
	return *ref, nil
}
