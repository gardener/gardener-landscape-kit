// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
	"fmt"
	"maps"
	"reflect"
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

// NewWithOverride creates a component vector by merging override entries on top of the base YAML.
// The override file uses the same Components schema but may list only a subset of components.
// Components present in the override replace their counterparts in base; new names are appended.
func NewWithOverride(base []byte, override []byte) (Interface, error) {
	baseObj := Components{}
	if err := yaml.Unmarshal(base, &baseObj); err != nil {
		return nil, fmt.Errorf("failed to parse base component vector: %w", err)
	}
	if errList := ValidateComponents(&baseObj, field.NewPath("")); len(errList) > 0 {
		return nil, fmt.Errorf("invalid base component vector: %w", errList.ToAggregate())
	}

	overrideObj := Components{}
	if err := yaml.Unmarshal(override, &overrideObj); err != nil {
		return nil, fmt.Errorf("failed to parse override component vector: %w", err)
	}
	merged := mergeComponents(&baseObj, &overrideObj)

	// Validate merged entries (name + version required per entry)
	for i, ov := range merged.Components {
		if errList := validateComponentVector(ov, field.NewPath("").Child("components").Index(i)); len(errList) > 0 {
			return nil, fmt.Errorf("invalid merged component vector: %w", errList.ToAggregate())
		}
	}

	result := &components{
		nameToComponentVector: make(map[string]*ComponentVector, len(merged.Components)),
	}
	for _, cv := range merged.Components {
		result.nameToComponentVector[cv.Name] = cv
	}
	return result, nil
}

// mergeComponents merges override components on top of base components.
// For each component in override: if its name exists in base, the base entry is merged with the override (override values take precedence);
// otherwise it is appended. Returns a new Components struct with the merged result.
func mergeComponents(base, override *Components) *Components {
	// Build an ordered list starting from base, replacing entries found in override.
	nameToOverride := make(map[string]*ComponentVector, len(override.Components))
	for _, ov := range override.Components {
		nameToOverride[ov.Name] = ov
	}

	merged := make([]*ComponentVector, 0, len(base.Components)+len(override.Components))
	seen := make(map[string]struct{}, len(base.Components))
	for _, bc := range base.Components {
		ov := nameToOverride[bc.Name]
		merged = append(merged, mergeComponentVector(ov, bc))
		seen[bc.Name] = struct{}{}
	}
	// Append components from override that were not present in base.
	for _, ov := range override.Components {
		if _, exists := seen[ov.Name]; !exists {
			merged = append(merged, ov)
		}
	}
	return &Components{Components: merged}
}

// mergeComponentVector fills zero-value fields in ov from the corresponding fields in bc.
// The Name field is skipped as it is the identity key.
func mergeComponentVector(ov, bc *ComponentVector) *ComponentVector {
	if ov == nil {
		return bc
	}
	ovVal := reflect.ValueOf(ov).Elem()
	bcVal := reflect.ValueOf(bc).Elem()
	for i := range ovVal.NumField() {
		if ovVal.Type().Field(i).Name == "Name" {
			continue
		}
		if ovVal.Field(i).IsZero() {
			ovVal.Field(i).Set(bcVal.Field(i))
		}
	}
	return ov
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
