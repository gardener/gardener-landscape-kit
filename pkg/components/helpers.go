// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package components

import "fmt"

// GetComponentVectorTemplateValues returns the template values from the component vector, such as OCI image references, Helm chart references, etc.
// See `docs/usage/component.md` for details.
func GetComponentVectorTemplateValues(opts Options, relativeComponentDir string, componentName string) (map[string]any, error) {
	componentsVector, err := opts.GetComponentsVector(relativeComponentDir)
	if err != nil {
		return nil, fmt.Errorf("error getting components vector: %w", err)
	}
	cv := componentsVector.FindComponentVector(componentName)
	if cv == nil {
		err := fmt.Errorf("component vector not found for component %s", componentName)
		opts.GetLogger().Error(err, "GetComponentVectorTemplateValues failed", "component", componentName)
		return nil, err
	}
	return cv.TemplateValues()
}
