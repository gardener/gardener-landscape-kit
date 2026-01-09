// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
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
