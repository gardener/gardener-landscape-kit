// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
	"fmt"
	"maps"
	"slices"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"
)

// components is a wrapper type for component vectors that implements Interface.
type components struct {
	nameToComponentVector map[string]*ComponentVector
}

// FindComponentVersion finds the version of the component with the given name.
func (c *components) FindComponentVersion(name string) (string, bool) {
	if component, exists := c.nameToComponentVector[name]; exists {
		return component.Version, exists
	}
	return "", false
}

// FindComponentVector finds the ComponentVector of the component with the given name.
// Returns the ComponentVector if found, otherwise nil.
func (c *components) FindComponentVector(name string) *ComponentVector {
	if component, exists := c.nameToComponentVector[name]; exists {
		h := *component
		return &h
	}
	return nil
}

// ComponentNames returns the sorted list of component names in the component vector.
func (c *components) ComponentNames() []string {
	return slices.Sorted(maps.Keys(c.nameToComponentVector))
}

// New creates a new component vector from the given YAML input.
func New(input []byte) (Interface, error) {
	componentsObj := Components{}

	if err := yaml.Unmarshal(input, &componentsObj); err != nil {
		return nil, err
	}

	// Validate components
	if errList := ValidateComponents(&componentsObj, field.NewPath("")); len(errList) > 0 {
		return nil, errList.ToAggregate()
	}

	components := &components{
		nameToComponentVector: make(map[string]*ComponentVector),
	}

	for _, component := range componentsObj.Components {
		components.nameToComponentVector[component.Name] = component
	}

	return components, nil
}

// TemplateValues returns the template values for the component vector.
func (cv *ComponentVector) TemplateValues() (map[string]any, error) {
	m := map[string]any{
		"resources": cv.Resources,
	}
	if cv.ImageVectorOverwrite != nil {
		// Marshal the ImageVectorOverwrite as it is expected to be as string.
		data, err := yaml.Marshal(cv.ImageVectorOverwrite)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal image vector overwrite: %w", err)
		}
		m["imageVectorOverwrite"] = string(data)
	}
	if cv.ComponentImageVectorOverwrites != nil {
		// Marshal the ComponentImageVectorOverwrites as it is expected to be as string.
		// We need to marshal each ComponentImageVectorOverwrite separately to ensure the correct format.
		type component struct {
			Name                 string `json:"name"`
			ImageVectorOverwrite string `json:"imageVectorOverwrite"`
		}
		type componentImageVectorOverwritesForTemplate struct {
			Components []component `json:"components"`
		}
		var value componentImageVectorOverwritesForTemplate
		for _, c := range cv.ComponentImageVectorOverwrites.Components {
			cdata, err := yaml.Marshal(c.ImageVectorOverwrite)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal image vector overwrite for component %s: %w", c.Name, err)
			}
			value.Components = append(value.Components, component{Name: c.Name, ImageVectorOverwrite: string(cdata)})
		}
		data, err := yaml.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal component image vector overwrites: %w", err)
		}
		m["componentImageVectorOverwrites"] = string(data)
	}
	return m, nil
}

// GetTemplateResourceValue retrieves a nested value from the Resources map using the provided keys.
// It returns nil if any key in the path does not exist.
func (cv *ComponentVector) GetTemplateResourceValue(keys ...string) any {
	var current any = cv.Resources
	for _, key := range keys {
		currentMap, ok := current.(map[string]any)
		if !ok {
			return nil
		}
		current = currentMap[key]
	}
	return current
}
