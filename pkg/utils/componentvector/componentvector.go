// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/version"
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

func newComponents(componentObj *Components) *components {
	result := &components{
		nameToComponentVector: make(map[string]*ComponentVector, len(componentObj.Components)),
	}
	for _, cv := range componentObj.Components {
		result.nameToComponentVector[cv.Name] = cv
	}
	return result
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

	return MergeAndValidateComponents(newComponents(&baseObj), overrides...)
}

// MergeAndValidateComponents merges the base component vector with the overrides and validates the result.
func MergeAndValidateComponents(base Interface, overrides ...[]byte) (Interface, error) {
	result, ok := base.(*components)
	if !ok {
		return nil, fmt.Errorf("base component vector is not of type *components")
	}
	for _, override := range overrides {
		overrideObj := Components{}
		if err := yaml.Unmarshal(override, &overrideObj); err != nil {
			return nil, fmt.Errorf("failed to parse override component vector: %w", err)
		}

		overrideComponents := newComponents(&overrideObj)
		result = mergeComponents(result, overrideComponents)
	}

	// Validate merged entries (name + version required per entry)
	for cv := range maps.Values(result.nameToComponentVector) {
		if errList := validateComponentVector(cv, field.NewPath("")); len(errList) > 0 {
			return nil, fmt.Errorf("invalid merged component vector: %w", errList.ToAggregate())
		}
	}

	return result, nil
}

func mergeComponents(base, override *components) *components {
	merged := make(map[string]*ComponentVector, len(base.nameToComponentVector)+len(override.nameToComponentVector))
	for name, bc := range base.nameToComponentVector {
		oc := override.nameToComponentVector[bc.Name]
		merged[name] = mergeComponentVector(bc, oc)
	}
	// Add components from override that were not present in base.
	for name, ov := range override.nameToComponentVector {
		if _, ok := merged[name]; !ok {
			merged[name] = ov
		}
	}
	return &components{nameToComponentVector: merged}
}

func mergeImageVectorOverwrite(base, override *ImageVectorOverwrite) *ImageVectorOverwrite {
	if override == nil {
		return base
	}

	nameToImage := make(map[string]imagevector.ImageSource)
	if base != nil {
		for _, img := range base.Images {
			nameToImage[img.Name] = img
		}
	}
	for _, img := range override.Images {
		nameToImage[img.Name] = img
	}
	return &ImageVectorOverwrite{Images: slices.SortedFunc(
		maps.Values(nameToImage),
		func(a, b imagevector.ImageSource) int {
			return strings.Compare(a.Name, b.Name)
		},
	)}
}

func mergeComponentVector(bc, ov *ComponentVector) *ComponentVector {
	if ov == nil {
		return bc
	}
	if ov.Version == "" {
		ov.Version = bc.Version
	}
	if ov.SourceRepository == nil {
		ov.SourceRepository = bc.SourceRepository
	}
	ov.Resources = mergeResources(bc.Resources, ov.Resources)
	ov.ImageVectorOverwrite = mergeImageVectorOverwrite(bc.ImageVectorOverwrite, ov.ImageVectorOverwrite)
	ov.ComponentImageVectorOverwrites = mergeComponentImageVectorOverwrites(bc.ComponentImageVectorOverwrites, ov.ComponentImageVectorOverwrites)
	return ov
}

func mergeResources(base, override map[string]ResourceData) map[string]ResourceData {
	if len(override) == 0 {
		return base
	}
	result := make(map[string]ResourceData, len(base))
	maps.Copy(result, base)
	maps.Copy(result, override)
	return result
}

func mergeComponentImageVectorOverwrites(base, override *ComponentImageVectorOverwrites) *ComponentImageVectorOverwrites {
	if override == nil {
		return base
	}
	nameToOverride := make(map[string]ComponentImageVectorOverwrite, len(override.Components))
	for _, oc := range override.Components {
		nameToOverride[oc.Name] = oc
	}
	if base != nil {
		for _, bc := range base.Components {
			if oc, exists := nameToOverride[bc.Name]; exists {
				mergedOverwrite := mergeImageVectorOverwrite(&bc.ImageVectorOverwrite, &oc.ImageVectorOverwrite)
				nameToOverride[bc.Name] = ComponentImageVectorOverwrite{Name: bc.Name, ImageVectorOverwrite: *mergedOverwrite}
			} else {
				nameToOverride[bc.Name] = bc
			}
		}
	}

	return &ComponentImageVectorOverwrites{Components: slices.SortedFunc(
		maps.Values(nameToOverride),
		func(a, b ComponentImageVectorOverwrite) int {
			return strings.Compare(a.Name, b.Name)
		},
	)}
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

		// Also expose images as a navigable map keyed by camelCase name, so templates can reference
		// individual image refs via e.g. {{ .images.sourceController.ref }}.
		imagesMap := make(map[string]any, len(cv.ImageVectorOverwrite.Images))
		for _, img := range cv.ImageVectorOverwrite.Images {
			ref := ""
			if img.Ref != nil {
				ref = *img.Ref
			} else if img.Repository != nil && img.Tag != nil {
				ref = *img.Repository + ":" + *img.Tag
			}
			imagesMap[kebabToCamelCase(img.Name)] = map[string]any{"ref": ref}
		}
		m["images"] = imagesMap
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

// NameVersionBytes marshals cv into a name+version-only Components YAML, stripping all other fields.
// This compact format is used for the .glk/defaults/ snapshot and as the three-way merge baseline in plain.go.
func NameVersionBytes(cv Interface) ([]byte, error) {
	stripped := &Components{}
	for _, name := range cv.ComponentNames() {
		version, _ := cv.FindComponentVersion(name)
		stripped.Components = append(stripped.Components, &ComponentVector{Name: name, Version: version})
	}
	data, err := yaml.Marshal(stripped)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal component versions: %w", err)
	}
	return data, nil
}

// VectorFileName is the name of the component vector metadata file
const VectorFileName = "components.yaml"

// WriteComponentVectorMetadata writes the full component vector to .glk/meta/components.yaml
func WriteComponentVectorMetadata(cv Interface, targetPath string, fs afero.Afero) error {
	components := &Components{}
	for _, name := range cv.ComponentNames() {
		components.Components = append(components.Components, cv.FindComponentVector(name))
	}

	data, err := yaml.Marshal(components)
	if err != nil {
		return fmt.Errorf("failed to marshal component vector: %w", err)
	}

	metaDir := filepath.Join(targetPath, files.GLKSystemDirName, version.MetaDirName)
	if err := fs.MkdirAll(metaDir, 0744); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	componentVectorFilePath := filepath.Join(metaDir, VectorFileName)
	if err := fs.WriteFile(componentVectorFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write component vector metadata: %w", err)
	}

	return nil
}

// ReadComponentVectorMetadata reads the component vector metadata from .glk/meta/components.yaml and returns it as an Interface.
func ReadComponentVectorMetadata(targetPath string, fs afero.Afero) (Interface, error) {
	componentVectorFilePath := filepath.Join(targetPath, files.GLKSystemDirName, version.MetaDirName, VectorFileName)
	exists, err := fs.Exists(componentVectorFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to check existence of component vector metadata file: %w", err)
	}
	// TODO(timuthy): This should be an error in a later version. Instead it should be checked if GLK is generating the first time.
	if !exists {
		return nil, nil
	}

	data, err := fs.ReadFile(componentVectorFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read component vector metadata: %w", err)
	}

	return NewWithOverride(data)
}

// kebabToCamelCase converts a kebab-case string to camelCase, e.g. "source-controller" → "sourceController".
func kebabToCamelCase(s string) string {
	parts := strings.Split(s, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			runes := []rune(parts[i])
			runes[0] = unicode.ToUpper(runes[0])
			parts[i] = string(runes)
		}
	}
	return strings.Join(parts, "")
}
