// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import "fmt"

// GetComponentVectorTemplateValues returns the template values from the component vector, such as OCI image references, Helm chart references, etc.
// See `docs/usage/component.md` for details.
func GetComponentVectorTemplateValues(opts Options, componentName string) (map[string]any, error) {
	cv := opts.GetComponentVector().FindComponentVector(componentName)
	if cv == nil {
		err := fmt.Errorf("component vector not found for component %s", componentName)
		opts.GetLogger().Error(err, "GetComponentVectorTemplateValues failed", "component", componentName)
		return nil, err
	}
	return cv.TemplateValues()
}
