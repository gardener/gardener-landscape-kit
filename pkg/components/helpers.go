// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

// DefaultResourcesFactory is a function type that defines a factory for creating default resources based on a given version.
type DefaultResourcesFactory func(version string) map[string]any

// GetRenderValues returns the render values for a component based on the provided options and component name.
// It first checks if the component vector contains the component and its resources.
// If not, it falls back to using the default resources factory with the component version from the component vector.
func GetRenderValues(opts Options, componentName string, factory DefaultResourcesFactory) (map[string]any, error) {
	cv := opts.GetComponentVector().FindComponentVector(componentName)
	if cv == nil || len(cv.Resources) == 0 {
		version, exists := opts.GetComponentVector().FindComponentVersion(componentName)
		if !exists {
			opts.GetLogger().Info("Component version not found in component vector, falling back to empty version", "component", componentName)
		}
		return map[string]any{
			"resources": factory(version),
		}, nil
	}
	return cv.TemplateValues()
}
