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

	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
)

const (
	// ComponentName is the name of the gardener-operator component.
	ComponentName = "gardener-operator"
	// ComponentDirectory is the directory of the gardener-operator component within the base components directory.
	ComponentDirectory = "gardener/operator"

	templateSuffix = ".tpl"
)

var (
	//go:embed version
	fallbackComponentVersion string

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
	objects, err := renderTemplateFiles(baseTemplates, baseTemplateDir, map[string]any{
		"version": fallbackComponentVersion, // TODO(LucaBernstein): Get actual version from the versions handling component once available.
	})
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem())
}

func writeLandscapeTemplateFiles(opts components.LandscapeOptions) error {
	var (
		relativeOverrideDir = path.Join(components.DirName, ComponentDirectory)
		relativeRepoRoot    = files.CalculatePathToComponentBase(opts.GetRelativeLandscapePath(), relativeOverrideDir)
	)

	objects, err := renderTemplateFiles(landscapeTemplates, landscapeTemplateDir, map[string]any{
		"relativePathToBaseComponent": path.Join(relativeRepoRoot, opts.GetRelativeBasePath(), relativeOverrideDir),
		"landscapeComponentPath":      path.Join(opts.GetRelativeLandscapePath(), relativeOverrideDir),
	})
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem())
}

func renderTemplateFiles(templateFS embed.FS, templateDir string, vars map[string]any) (map[string][]byte, error) {
	var objects = make(map[string][]byte)

	dir, err := templateFS.ReadDir(templateDir)
	if err != nil {
		return nil, err
	}
	for _, file := range dir {
		fileName := file.Name()
		fileContents, err := templateFS.ReadFile(path.Join(templateDir, fileName))
		if err != nil {
			return nil, err
		}

		if fileContents, fileName, err = renderTemplate(fileContents, fileName, vars); err != nil {
			return nil, err
		}
		objects[fileName] = fileContents
	}

	return objects, nil
}

func renderTemplate(fileContents []byte, fileName string, vars map[string]any) ([]byte, string, error) {
	fileTemplate, err := template.New(fileName).Funcs(sprig.TxtFuncMap()).Parse(string(fileContents))
	if err != nil {
		return nil, "", fmt.Errorf("error parsing template '%s': %w", fileName, err)
	}

	var templatedResult bytes.Buffer
	if err := fileTemplate.Execute(&templatedResult, vars); err != nil {
		return nil, "", fmt.Errorf("error executing '%s' template: %w", fileName, err)
	}

	fileContents = templatedResult.Bytes()
	fileName = strings.TrimSuffix(fileName, templateSuffix)
	return fileContents, fileName, nil
}
