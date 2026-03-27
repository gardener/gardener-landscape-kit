// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
	"fmt"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
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

// NewWithOverride creates a component vector by merging overrides entries on top of the base YAML.
// The overrides files use the same Components schema but may list only a subset of components.
// Components present in the override replace their counterparts in base; new names are appended.
// Overrides are applied in order: later entries take precedence over earlier ones.
func NewWithOverride(base []byte, overrides ...[]byte) (Interface, error) {
	baseObj := Components{}
	if err := yaml.Unmarshal(base, &baseObj); err != nil {
		return nil, fmt.Errorf("failed to parse base component vector: %w", err)
	}
	if errList := ValidateComponents(&baseObj, field.NewPath("")); len(errList) > 0 {
		return nil, fmt.Errorf("invalid base component vector: %w", errList.ToAggregate())
	}

	merged := &baseObj
	for _, override := range overrides {
		overrideObj := Components{}
		if err := yaml.Unmarshal(override, &overrideObj); err != nil {
			return nil, fmt.Errorf("failed to parse override component vector: %w", err)
		}
		merged = mergeComponents(merged, &overrideObj)
	}

	// Validate merged entries (name + version required per entry)
	for i, cv := range merged.Components {
		if errList := validateComponentVector(cv, field.NewPath("").Child("components").Index(i)); len(errList) > 0 {
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
		oc := nameToOverride[bc.Name]
		merged = append(merged, mergeComponentVector(oc, bc))
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

const (
	defaultVersionCommentMarker = "# <-- gardener-landscape-kit version default"
)

// stripDefaultVersionComments removes GLK-managed default-version comment lines from a components.yaml file.
// A line is considered GLK-managed when it contains the unique GLK marker suffix.
// Stripping them before the three-way merge ensures the canonical comment is always (re-)written on the next run, even when the user has edited the comment text.
func stripDefaultVersionComments(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, defaultVersionCommentMarker) {
			out = append(out, line)
		}
	}
	return []byte(strings.Join(out, "\n"))
}

// WriteComponentVectorFile writes the component vector file effectively used to the target directory if applicable.
func WriteComponentVectorFile(fs afero.Afero, targetDirPath string, componentVector Interface) error {
	var (
		comp                                 = &Components{}
		postGenerateDefaultVersionCommentFns []func(string) string
	)
	cvDefault, err := NewWithOverride(componentvector.DefaultComponentsYAML)
	if err != nil {
		return fmt.Errorf("failed to build default component vector: %w", err)
	}
	for _, componentName := range componentVector.ComponentNames() {
		componentVersion, _ := componentVector.FindComponentVersion(componentName)
		comp.Components = append(comp.Components, &ComponentVector{
			Name:    componentName,
			Version: componentVersion,
		})
		defaultVersion, found := cvDefault.FindComponentVersion(componentName)
		if found && componentVersion != defaultVersion {
			defaultVersionComment := "# version: " + defaultVersion + " " + defaultVersionCommentMarker
			postGenerateDefaultVersionCommentFns = append(postGenerateDefaultVersionCommentFns, func(data string) string {
				return strings.ReplaceAll(data, componentName+"\n", componentName+"\n"+defaultVersionComment+"\n")
			})
		}
	}
	data, err := yaml.Marshal(comp)
	if err != nil {
		return fmt.Errorf("failed to marshal component vector: %w", err)
	}

	header := []byte(strings.Join([]string{
		"# This file is updated by the gardener-landscape-kit.",
		"# If this file is specified in the gardener-landscape-kit configuration file, the component versions will be used as overrides.",
		"# If custom component versions should be used, it is recommended to modify the specified versions here and run the `generate` command afterwards.",
	}, "\n") + "\n")

	// Before writing, strip any GLK-managed default-version comment lines from the on-disk file.
	// This resets GLK-owned annotations so the canonical comment is always (re-)applied below, even when the user has edited or removed the comment line.
	filePath := filepath.Join(targetDirPath, ComponentVectorFilename)
	if existing, readErr := fs.ReadFile(filePath); readErr == nil {
		if stripped := stripDefaultVersionComments(existing); string(stripped) != string(existing) {
			if writeErr := fs.WriteFile(filePath, stripped, 0600); writeErr != nil {
				return writeErr
			}
		}
	}

	// Pass 1: write without default-version comments so the three-way merge operates on
	// comment-free content. This establishes a clean baseline in the .glk/defaults/ snapshot.
	dataWithoutComments := append(header, data...)
	if err := files.WriteObjectsToFilesystem(map[string][]byte{ComponentVectorFilename: dataWithoutComments}, targetDirPath, "", fs); err != nil {
		return err
	}

	// Pass 2: inject default-version comments and write again. Because the .glk/defaults/ snapshot from Pass 1 has no comments,
	// the comments are always treated as "new" by the three-way merge and are therefore reliably written into the output file.
	for _, fn := range postGenerateDefaultVersionCommentFns {
		data = []byte(fn(string(data)))
	}
	dataWithComments := append(header, data...)
	return files.WriteObjectsToFilesystem(map[string][]byte{ComponentVectorFilename: dataWithComments}, targetDirPath, "", fs)
}
