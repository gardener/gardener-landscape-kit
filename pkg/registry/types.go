// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

// Interface is the interface for a component registry.
type Interface interface {
	// RegisterComponent registers a component in the registry.
	RegisterComponent(name string, component components.Interface)
	// GenerateBase generates the base component.
	GenerateBase(opts components.Options) error
	// GenerateLandscape generates the landscape component.
	GenerateLandscape(opts components.LandscapeOptions) error
	// WriteComponentVectorFile writes the component vector file effectively used to the target directory if applicable.
	WriteComponentVectorFile(opts components.Options, outputDir string) error
}

type registry struct {
	components *orderedmap.OrderedMap[string, components.Interface]
}

// RegisterComponent registers a component in the registry.
func (r *registry) RegisterComponent(name string, component components.Interface) {
	r.components.Set(name, component)
}

// GenerateBase generates the base component.
func (r *registry) GenerateBase(opts components.Options) error {
	for _, component := range r.components.AllFromFront() {
		if err := component.GenerateBase(opts); err != nil {
			return err
		}
	}

	return r.findAndRenderCustomComponents(opts)
}

// GenerateLandscape generates the landscape component.
func (r *registry) GenerateLandscape(opts components.LandscapeOptions) error {
	for _, component := range r.components.AllFromFront() {
		if err := component.GenerateLandscape(opts); err != nil {
			return err
		}
	}

	return r.findAndRenderCustomComponents(opts)
}

func (r *registry) findAndRenderCustomComponents(opts components.Options) error {
	return opts.GetFilesystem().Walk(opts.GetTargetPath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || info.Name() != ocm.CustomOCMComponentNameFilename {
			return nil
		}

		content, err := opts.GetFilesystem().ReadFile(path)
		if err != nil {
			return err
		}
		name := strings.TrimSpace(string(content))
		opts.GetLogger().Info("Found custom component", "name", name, "file", path)

		return r.renderCustomComponents(name, filepath.Dir(path), opts)
	})
}

func (r *registry) renderCustomComponents(ocmComponentName, componentDir string, opts components.Options) error {
	cv := opts.GetComponentVector().FindComponentVector(ocmComponentName)
	if cv == nil {
		return fmt.Errorf("no component vector found for custom component %s", ocmComponentName)
	}

	return opts.GetFilesystem().Walk(componentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".template") {
			return nil
		}
		content, err := opts.GetFilesystem().ReadFile(path)
		if err != nil {
			return err
		}
		values, err := cv.TemplateValues()
		if err != nil {
			return fmt.Errorf("error getting template values for custom component %s: %w", ocmComponentName, err)
		}
		renderedContent, _, err := files.RenderTemplate(content, info.Name(), values)
		if err != nil {
			return fmt.Errorf("error rendering template file %s for custom component %s: %w", path, ocmComponentName, err)
		}
		targetFile := strings.TrimSuffix(path, ".template")
		if err := opts.GetFilesystem().WriteFile(targetFile, renderedContent, 0600); err != nil {
			return fmt.Errorf("error writing rendered template file %s for custom component %s: %w", targetFile, ocmComponentName, err)
		}
		opts.GetLogger().Info("Rendered custom component template file", "component", ocmComponentName, "templateFile", path, "outputFile", targetFile)
		return nil
	})
}

const (
	componentVectorFilename     = "components.yaml"
	defaultVersionCommentMarker = "# <-- gardener-landscape-kit version default"
)

// stripDefaultVersionComments removes GLK-managed default-version comment lines from a components.yaml file.
// A line is considered GLK-managed when it contains the unique GLK marker suffix.
// Stripping them before the three-way merge ensures the canonical comment is always (re-)written on the next run, even when the user has edited the comment text.
func stripDefaultVersionComments(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, defaultVersionCommentMarker) {
			out = append(out, line)
		}
	}
	return []byte(strings.Join(out, "\n"))
}

// WriteComponentVectorFile writes the component vector file effectively used to the target directory if applicable.
func (r *registry) WriteComponentVectorFile(opts components.Options, outputDir string) error {
	var (
		comp                                 = &utilscomponentvector.Components{}
		postGenerateDefaultVersionCommentFns []func(string) string
	)
	cvDefault, err := utilscomponentvector.NewWithOverride(componentvector.DefaultComponentsYAML, nil)
	if err != nil {
		return fmt.Errorf("failed to build default component vector: %w", err)
	}
	for _, componentName := range opts.GetComponentVector().ComponentNames() {
		componentVersion, _ := opts.GetComponentVector().FindComponentVersion(componentName)
		comp.Components = append(comp.Components, &utilscomponentvector.ComponentVector{
			Name:    componentName,
			Version: componentVersion,
		})
		defaultVersion, found := cvDefault.FindComponentVersion(componentName)
		if found && componentVersion != defaultVersion {
			defaultVersionComment := "# version: " + defaultVersion + " " + defaultVersionCommentMarker
			postGenerateDefaultVersionCommentFns = append(postGenerateDefaultVersionCommentFns, func(data string) string {
				return strings.ReplaceAll(data, componentName+"\n", componentName+"\n"+defaultVersionComment+"\n")
			})
		}
	}
	data, err := yaml.Marshal(comp)
	if err != nil {
		return fmt.Errorf("failed to marshal image vector: %w", err)
	}

	header := []byte(strings.Join([]string{
		"# This file is updated by the gardener-landscape-kit.",
		"# If this file is specified in the gardener-landscape-kit configuration file, the component versions will be used as overrides.",
		"# If custom component versions should be used, it is recommended to modify the specified versions here and run the `generate` command afterwards.",
	}, "\n") + "\n")

	// Before writing, strip any GLK-managed default-version comment lines from the on-disk file.
	// This resets GLK-owned annotations so the canonical comment is always (re-)applied below, even when the user has edited or removed the comment line.
	filePath := filepath.Join(outputDir, componentVectorFilename)
	if existing, readErr := opts.GetFilesystem().ReadFile(filePath); readErr == nil {
		if stripped := stripDefaultVersionComments(existing); string(stripped) != string(existing) {
			if writeErr := opts.GetFilesystem().WriteFile(filePath, stripped, 0600); writeErr != nil {
				return writeErr
			}
		}
	}

	// Pass 1: write without default-version comments so the three-way merge operates on
	// comment-free content. This establishes a clean baseline in the .glk/defaults/ snapshot.
	dataWithoutComments := append(header, data...)
	if err := files.WriteObjectsToFilesystem(map[string][]byte{componentVectorFilename: dataWithoutComments}, outputDir, "", opts.GetFilesystem()); err != nil {
		return err
	}

	// Pass 2: inject default-version comments and write again. Because the .glk/defaults/ snapshot from Pass 1 has no comments,
	// the comments are always treated as "new" by the three-way merge and are therefore reliably written into the output file.
	for _, fn := range postGenerateDefaultVersionCommentFns {
		data = []byte(fn(string(data)))
	}
	dataWithComments := append(header, data...)
	return files.WriteObjectsToFilesystem(map[string][]byte{componentVectorFilename: dataWithComments}, outputDir, "", opts.GetFilesystem())
}

// New creates a new component registry.
func New() Interface {
	return &registry{
		components: orderedmap.NewOrderedMap[string, components.Interface](),
	}
}
