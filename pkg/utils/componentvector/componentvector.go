// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
	"fmt"
	"maps"
	"slices"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

const (
	resourcesKey                      = "resources"
	imageVectorOverwriteKey           = "imageVectorOverwrite"
	componentImageVectorOverwritesKey = "componentImageVectorOverwrites"
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
		return new(*component)
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
// It converts the resources to an unstructured map and marshals the image vector overwrites as strings if they are present.
// The resources are patched by replacing the string `${version}` with the version of the component vector before being returned as template values.
func (cv *ComponentVector) TemplateValues() (map[string]any, error) {
	if err := ensureReferences(cv.Resources, cv.Version); err != nil {
		return nil, fmt.Errorf("failed to ensure references for Helm charts and OCI images: %w", err)
	}
	resources, err := resourcesToUnstructuredMap(cv.Resources)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resources to unstructured map: %w", err)
	}
	m := map[string]any{
		"version":    cv.Version,
		resourcesKey: resources,
	}
	if cv.ImageVectorOverwrite != nil {
		// Marshal the ImageVectorOverwrite as it is expected to be as string.
		data, err := yaml.Marshal(cv.ImageVectorOverwrite)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal image vector overwrite: %w", err)
		}
		m[imageVectorOverwriteKey] = string(data)
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
		m[componentImageVectorOverwritesKey] = string(data)
	}
	return m, nil
}

func ensureReferences(resources map[string]ResourceData, version string) error {
	for resourceName, resourceData := range resources {
		if resourceData.HelmChart != nil && resourceData.HelmChart.Ref == nil {
			if resourceData.HelmChart.Repository == nil {
				return fmt.Errorf("missing reference or repository for Helm chart in resource %s", resourceName)
			}
			resourceData.HelmChart.Ref = new(*resourceData.HelmChart.Repository + ":" + ptr.Deref(resourceData.HelmChart.Tag, version))
		}
		if resourceData.OCIImage != nil && resourceData.OCIImage.Ref == nil {
			if resourceData.OCIImage.Repository == nil {
				return fmt.Errorf("missing reference or repository for OCI image in resource %s", resourceName)
			}
			resourceData.OCIImage.Ref = new(*resourceData.OCIImage.Repository + ":" + ptr.Deref(resourceData.OCIImage.Tag, version))
		}
		resources[resourceName] = resourceData
	}
	return nil
}

func resourcesToUnstructuredMap(resources map[string]ResourceData) (map[string]any, error) {
	unstructuredMap := make(map[string]any)
	if len(resources) > 0 {
		data, err := yaml.Marshal(resources)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal resources: %w", err)
		}
		if err := yaml.Unmarshal(data, &unstructuredMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resources: %w", err)
		}
	}
	return unstructuredMap, nil
}
