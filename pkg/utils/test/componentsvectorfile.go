// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"errors"
	"fmt"

	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

// BuildComponentVectorFn is a function type that builds a component vector.
type BuildComponentVectorFn func() (componentvector.ComponentVector, error)

// CreateComponentsVectorFile creates a components vector YAML file in the given filesystem.
func CreateComponentsVectorFile(fs afero.Afero, build BuildComponentVectorFn) (string, error) {
	cv, err := build()
	if err != nil {
		return "", err
	}
	if cv.SourceRepository == "" {
		cv.SourceRepository = "https://github.com/dummy/repo"
	}
	components := &componentvector.Components{
		Components: []*componentvector.ComponentVector{&cv},
	}
	content, err := yaml.Marshal(components)
	if err != nil {
		return "", err
	}
	filePath := "/repo/components-vector.yaml"
	if err := fs.WriteFile(filePath, content, 0o644); err != nil {
		return "", err
	}
	return filePath, nil
}

// ComponentVectorFactoryBuilder helps to build BuildComponentVectorFn instances.
type ComponentVectorFactoryBuilder struct {
	cv  componentvector.ComponentVector
	err error
}

// NewComponentVectorFactoryBuilder creates a new ComponentVectorFactoryBuilder with the given name and version.
func NewComponentVectorFactoryBuilder(name, version string) *ComponentVectorFactoryBuilder {
	return &ComponentVectorFactoryBuilder{
		cv: componentvector.ComponentVector{
			Name:    name,
			Version: version,
		},
	}
}

// WithImageVectorOverwrite sets the image vector overwrite of the component vector.
func (b *ComponentVectorFactoryBuilder) WithImageVectorOverwrite(v componentvector.ImageVectorOverwrite) *ComponentVectorFactoryBuilder {
	b.cv.ImageVectorOverwrite = &v
	return b
}

// WithComponentImageVectorOverwrites sets the component image vector overwrites of the component vector.
func (b *ComponentVectorFactoryBuilder) WithComponentImageVectorOverwrites(v componentvector.ComponentImageVectorOverwrites) *ComponentVectorFactoryBuilder {
	b.cv.ComponentImageVectorOverwrites = &v
	return b
}

// WithResourcesYAML sets the resources of the component vector from the given YAML content.
func (b *ComponentVectorFactoryBuilder) WithResourcesYAML(yaml string) *ComponentVectorFactoryBuilder {
	unstructuredMap, err := UnmarshalToResources(yaml)
	if err != nil {
		b.err = errors.Join(b.err, fmt.Errorf("failed to unmarshal resources YAML: %w", err))
		return b
	}
	b.cv.Resources = unstructuredMap
	return b
}

// Build builds the BuildComponentVectorFn.
func (b *ComponentVectorFactoryBuilder) Build() BuildComponentVectorFn {
	return func() (componentvector.ComponentVector, error) {
		return b.cv, b.err
	}
}

// UnmarshalToResources unmarshals the given YAML content into a resources map.
func UnmarshalToResources(content string) (map[string]*componentvector.ResourceData, error) {
	result := make(map[string]*componentvector.ResourceData)
	if err := yaml.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	return result, nil
}
