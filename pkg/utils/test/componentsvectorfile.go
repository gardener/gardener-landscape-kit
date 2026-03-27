// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

// BuildComponentVectorFn is a function type that builds a component vector.
type BuildComponentVectorFn func() (utilscomponentvector.ComponentVector, error)

// CreateComponentsVectorFile creates a components vector YAML file in the given filesystem.
func CreateComponentsVectorFile(fs afero.Afero, build BuildComponentVectorFn) error {
	cv, err := build()
	if err != nil {
		return err
	}
	components := &utilscomponentvector.Components{
		Components: []*utilscomponentvector.ComponentVector{&cv},
	}
	content, err := yaml.Marshal(components)
	if err != nil {
		return err
	}
	filePath := "/repo/baseDir/components.yaml"
	if err := fs.WriteFile(filePath, content, 0o644); err != nil {
		return err
	}
	return nil
}

// ComponentVectorFactoryBuilder helps to build BuildComponentVectorFn instances.
type ComponentVectorFactoryBuilder struct {
	cv  utilscomponentvector.ComponentVector
	err error
}

// NewComponentVectorFactoryBuilder creates a new ComponentVectorFactoryBuilder with the given name and version.
func NewComponentVectorFactoryBuilder(name, version string) *ComponentVectorFactoryBuilder {
	return &ComponentVectorFactoryBuilder{
		cv: utilscomponentvector.ComponentVector{
			Name:    name,
			Version: version,
		},
	}
}

// WithImageVectorOverwrite sets the image vector overwrite of the component vector.
func (b *ComponentVectorFactoryBuilder) WithImageVectorOverwrite(v utilscomponentvector.ImageVectorOverwrite) *ComponentVectorFactoryBuilder {
	b.cv.ImageVectorOverwrite = &v
	return b
}

// WithComponentImageVectorOverwrites sets the component image vector overwrites of the component vector.
func (b *ComponentVectorFactoryBuilder) WithComponentImageVectorOverwrites(v utilscomponentvector.ComponentImageVectorOverwrites) *ComponentVectorFactoryBuilder {
	b.cv.ComponentImageVectorOverwrites = &v
	return b
}

// WithResourcesYAML sets the resources of the component vector from the given YAML content.
func (b *ComponentVectorFactoryBuilder) WithResourcesYAML(yaml string) *ComponentVectorFactoryBuilder {
	var err error
	b.cv.Resources, err = UnmarshalToResources(yaml)
	if err != nil {
		b.err = errors.Join(b.err, fmt.Errorf("failed to unmarshal resources YAML: %w", err))
	}
	return b
}

// WithDefaultResources sets the resources of the component vector from the default components YAML.
func (b *ComponentVectorFactoryBuilder) WithDefaultResources() *ComponentVectorFactoryBuilder {
	componentVector, err := utilscomponentvector.NewWithOverride(componentvector.DefaultComponentsYAML, nil)
	if err != nil {
		b.err = errors.Join(b.err, fmt.Errorf("failed to unmarshal default components YAML: %w", err))
		return b
	}
	cv := componentVector.FindComponentVector(b.cv.Name)
	if cv == nil {
		b.err = errors.Join(b.err, fmt.Errorf("failed to find component vector with name %s in default components: %w", b.cv.Name, err))
		return b
	}
	b.cv.Resources = cv.Resources
	return b
}

// Build builds the BuildComponentVectorFn.
func (b *ComponentVectorFactoryBuilder) Build() BuildComponentVectorFn {
	return func() (utilscomponentvector.ComponentVector, error) {
		return b.cv, b.err
	}
}

// UnmarshalToResources unmarshals the given YAML content into a resources map.
func UnmarshalToResources(content string) (map[string]utilscomponentvector.ResourceData, error) {
	result := make(map[string]utilscomponentvector.ResourceData)
	if err := yaml.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	return result, nil
}
