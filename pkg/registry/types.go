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

	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm"
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
	if cv == nil || len(cv.Resources) == 0 {
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
		renderedContent, _, err := files.RenderTemplate(content, info.Name(), cv.TemplateValues())
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

// New creates a new component registry.
func New() Interface {
	return &registry{
		components: orderedmap.NewOrderedMap[string, components.Interface](),
	}
}
