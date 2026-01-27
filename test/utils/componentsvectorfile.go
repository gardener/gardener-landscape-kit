// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"errors"
	"fmt"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"
)

type ComponentVectorFactory func() (componentvector.ComponentVector, error)

// CreateComponentsVectorFile creates a components vector YAML file in the given filesystem.
func CreateComponentsVectorFile(fs afero.Afero, cvf ComponentVectorFactory) (string, error) {
	cv, err := cvf()
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

// ComponentVectorFactoryBuilder helps to build ComponentVectorFactory instances.
type ComponentVectorFactoryBuilder struct {
	cv  componentvector.ComponentVector
	err error
}

// ComponentVector creates a new ComponentVectorFactoryBuilder with the given name and version.
func ComponentVector(name, version string) *ComponentVectorFactoryBuilder {
	return &ComponentVectorFactoryBuilder{
		cv: componentvector.ComponentVector{
			Name:    name,
			Version: version,
		},
	}
}

// WithImageVectorOverwrite sets the image vector overwrite of the component vector.
func (b *ComponentVectorFactoryBuilder) WithImageVectorOverwrite(s string) *ComponentVectorFactoryBuilder {
	b.cv.ImageVectorOverwrite = s
	return b
}

// WithComponentImageVectorOverwrites sets the component image vector overwrites of the component vector.
func (b *ComponentVectorFactoryBuilder) WithComponentImageVectorOverwrites(s string) *ComponentVectorFactoryBuilder {
	b.cv.ComponentImageVectorOverwrites = s
	return b
}

// WithResourcesYAML sets the resources of the component vector from the given YAML content.
func (b *ComponentVectorFactoryBuilder) WithResourcesYAML(yaml string) *ComponentVectorFactoryBuilder {
	unstructuredMap, err := UnmarshalToMap(yaml)
	if err != nil {
		b.err = errors.Join(b.err, fmt.Errorf("failed to unmarshal resources YAML: %w", err))
		return b
	}
	b.cv.Resources = unstructuredMap
	return b
}

// Build builds the ComponentVectorFactory.
func (b *ComponentVectorFactoryBuilder) Build() ComponentVectorFactory {
	return func() (componentvector.ComponentVector, error) {
		return b.cv, b.err
	}
}

// UnmarshalToMap unmarshals the given YAML content into a map.
func UnmarshalToMap(content string) (map[string]any, error) {
	result := make(map[string]any)
	if err := yaml.Unmarshal([]byte(content), &result); err != nil {
		return nil, err
	}
	return result, nil
}
