// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
	"sort"

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
	names := make([]string, 0, len(c.nameToComponentVector))
	for name := range c.nameToComponentVector {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
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
