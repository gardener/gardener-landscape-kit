// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import "fmt"

// GetTemplateValues returns the template values for a component based on the provided options and component name.
func GetTemplateValues(opts Options, componentName string) (map[string]any, error) {
	cv := opts.GetComponentVector().FindComponentVector(componentName)
	if cv == nil {
		err := fmt.Errorf("component vector not found for component %s", componentName)
		opts.GetLogger().Error(err, "GetTemplateValues failed", "component", componentName)
		return nil, err
	}
	return cv.TemplateValues()
}
