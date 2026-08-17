// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operator

import (
	"embed"
	"fmt"
	"path"
	"strings"

	"github.com/gardener/gardener/pkg/utils"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/components/helpers"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
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

// NewComponent creates a new gardener-operator component.
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

func (c *component) writeBaseTemplateFiles(opts components.Options) error {
	objects, err := files.RenderTemplateFiles(baseTemplates, baseTemplateDir, nil)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, c.Directory), opts.GetFilesystem(), opts.GetMergeMode())
}

func getTemplateValues(opts components.Options) (map[string]any, error) {
	cv := opts.GetComponentVector().FindComponentVector(componentvector.NameGardenerGardener)
	if cv == nil || len(cv.Resources) == 0 {
		gardenerVersion, exists := opts.GetComponentVector().FindComponentVersion(componentvector.NameGardenerGardener)
		if !exists {
			opts.GetLogger().Info("Component version not found in component vector, falling back to empty version", "component", componentvector.NameGardenerGardener)
		}
		return map[string]any{
			"repository": "europe-docker.pkg.dev/gardener-project/releases/charts/gardener/operator",
			"ref":        versionRefTemplateValue(gardenerVersion),
		}, nil
	}

	repository, version, err := getHelmChartRepoTagFromComponentVector("operator", cv)
	if err != nil {
		return nil, fmt.Errorf("failed to get operator Helm chart repository/tag from component vector: %w", err)
	}

	values, err := cv.TemplateValues()
	if err != nil {
		return nil, fmt.Errorf("failed to get template values from component vector: %w", err)
	}
	values["repository"] = repository
	values["ref"] = versionRefTemplateValue(version)
	return values, nil
}

func versionRefTemplateValue(version string) map[string]any {
	refKind := "tag"
	if strings.HasPrefix(version, "sha256:") {
		refKind = "digest"
	}
	return map[string]any{
		"kind":  refKind,
		"value": version,
	}
}

func (c *component) writeLandscapeTemplateFiles(opts components.LandscapeOptions) error {
	relativeComponentPath := path.Join(components.DirName, c.Directory)

	values, err := getTemplateValues(opts)
	if err != nil {
		return err
	}

	values = utils.MergeMaps(values, map[string]any{
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

func getHelmChartRepoTagFromComponentVector(name string, cv *utilscomponentvector.ComponentVector) (string, string, error) {
	if cv == nil || len(cv.Resources) == 0 {
		return "", "", fmt.Errorf("component vector or component resources are nil")
	}

	data, found := cv.Resources[name]
	if !found {
		return "", "", fmt.Errorf("no resources found for component %s", name)
	}
	if data.HelmChart == nil {
		return "", "", fmt.Errorf("HelmChart not found for component %s", name)
	}
	repository, tag, err := helpers.RepoTagFromRefOrParts(data.HelmChart.Repository, data.HelmChart.Tag, data.HelmChart.Ref)
	if err != nil {
		return "", "", fmt.Errorf("HelmChart reference not found for component %s: %w", name, err)
	}
	return repository, tag, nil
}
