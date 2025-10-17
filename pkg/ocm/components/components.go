// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	descriptorruntime "ocm.software/open-component-model/bindings/go/descriptor/runtime"
	descriptorv2 "ocm.software/open-component-model/bindings/go/descriptor/v2"
	accessv1 "ocm.software/open-component-model/bindings/go/oci/spec/access/v1"

	ocmimagevector "github.com/gardener/gardener-landscape-kit/pkg/ocm/imagevector"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/ociaccess"
)

// ComponentReference is a reference to a component in the format "name:version".
type ComponentReference string

const (
	rootComponentReference = "<ROOT>"
	resourceTypeOCIImage   = "ociImage"
	resourceTypeHelmChart  = "helmChart/v1"
	// ResourceTypeHelmChartImageMap is a resource type for helm chart image maps.
	ResourceTypeHelmChartImageMap = "helmchart-imagemap"
)

// ComponentReferenceFromNameAndVersion creates a ComponentReference from name and version.
func ComponentReferenceFromNameAndVersion(name, version string) ComponentReference {
	return ComponentReference(fmt.Sprintf("%s:%s", name, version))
}

// Dependency represents a dependency of a component, including its image vector.
type Dependency struct {
	ComponentReference

	ImageVector []ocmimagevector.ExtendedImageSource
}

// Components manages a collection of components, their dependencies, and associated image vectors.
type Components struct {
	lock         sync.Mutex
	dependents   map[ComponentReference]sets.Set[ComponentReference]
	dependencies map[ComponentReference][]Dependency
	mappedImages map[ComponentReference][]*ocmimagevector.ExtendedImageSource
	resources    map[ComponentReference][]Resource

	kubernetesComponent *ComponentReference
}

// Blobs represents a mapping from resource identifiers to their corresponding blob data.
type Blobs map[ociaccess.NameVersionType][]byte

// Resource represents a resource associated with a component.
type Resource struct {
	// Name is the name of the resource.
	Name string `json:"name"`
	// Version is the version of the resource.
	Version string `json:"version"`
	// Type is the type of the resource (e.g., "ociImage", "helmChart/v1").
	Type string `json:"type"`
	// Value is the reference or URL of the resource (for types like "ociImage", "helmChart/v1")
	// or other relevant information depending on the resource type.
	Value string `json:"value"`
}

// ResourcesOutput is the output format for the resources JSON output.
type ResourcesOutput struct {
	Resources []Resource `json:"resources"`
}

// NewComponents creates a new instance of Components.
func NewComponents() *Components {
	return &Components{
		dependents:   make(map[ComponentReference]sets.Set[ComponentReference]),
		dependencies: make(map[ComponentReference][]Dependency),
		mappedImages: make(map[ComponentReference][]*ocmimagevector.ExtendedImageSource),
		resources:    make(map[ComponentReference][]Resource),
	}
}

func (c *Components) addComponent(descriptor *descriptorruntime.Descriptor, blobs Blobs) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	cref := descriptorToComponentReference(descriptor)
	imageVector, err := c.extractImageVectorFromResources(descriptor)
	if err != nil {
		return fmt.Errorf("could not extract image vector from descriptor: %s", err)
	}
	// resources are added as dependency to itself.
	c.dependencies[cref] = append(c.dependencies[cref], Dependency{
		ComponentReference: cref,
		ImageVector:        imageVector,
	})

	resources, err := c.extractResourcesFromDescriptor(descriptor, blobs)
	if err != nil {
		return fmt.Errorf("could not extract resources from descriptor: %s", err)
	}
	c.resources[cref] = resources

	for _, label := range descriptor.Component.Labels {
		switch label.Name {
		case LabelImageVectorImages:
			images, err := rawToImageSources(label.Value)
			if err != nil {
				return fmt.Errorf("could not extract image vector from label value: %s", err)
			}
			if len(images) > 0 {
				c.mappedImages[cref] = images
			}
		case LabelImageVectorApplication:
			labelValue, err := toString(label.Value)
			if err != nil {
				return fmt.Errorf("unexpected value for label %q in component %s: %w", label.Name, cref, err)
			}
			if labelValue == LabelImageVectorApplicationValueKubernetes {
				if c.kubernetesComponent != nil {
					return fmt.Errorf("non-unique kubernetes component: [%s,%s]", *c.kubernetesComponent, cref)
				}
				c.kubernetesComponent = &cref
			}
		}
	}

	if _, exists := c.dependents[cref]; exists {
		return nil
	}

	c.dependents[cref] = sets.New[ComponentReference](rootComponentReference)
	c.dependencies[rootComponentReference] = append(c.dependencies[rootComponentReference], Dependency{ComponentReference: cref})

	return nil
}

func (c *Components) extractImageVectorFromResources(descriptor *descriptorruntime.Descriptor) ([]ocmimagevector.ExtendedImageSource, error) {
	var vector []ocmimagevector.ExtendedImageSource
	for _, res := range descriptor.Component.Resources {
		src, err := resourceToImageSource(res)
		if err != nil {
			return nil, fmt.Errorf("failed to convert resource %s to image source: %w", res.Name, err)
		}
		if src != nil {
			vector = append(vector, *src)
		}
	}
	return vector, nil
}

func (c *Components) extractResourcesFromDescriptor(descriptor *descriptorruntime.Descriptor, blobs Blobs) ([]Resource, error) {
	var resources []Resource
	for _, res := range descriptor.Component.Resources {
		switch res.Type {
		case resourceTypeOCIImage:
			src, err := resourceToImageSource(res)
			if err != nil {
				return nil, fmt.Errorf("failed to convert resource %s to image source: %w", res.Name, err)
			}
			if src != nil {
				var reference string
				if src.Ref != nil {
					reference = *src.Ref
				} else if src.Repository != nil && src.Tag != nil {
					reference = fmt.Sprintf("%s:%s", *src.Repository, *src.Tag)
				}

				if reference == "" {
					return nil, fmt.Errorf("could not determine reference for resource %s", res.Name)
				}
				resources = append(resources, Resource{
					Name:    src.Name,
					Version: res.Version,
					Type:    res.Type,
					Value:   reference,
				})
			}
		case resourceTypeHelmChart:
			var spec accessv1.OCIImage
			if err := ociaccess.DefaultScheme.Convert(res.Access, &spec); err != nil {
				return nil, err
			}
			resources = append(resources, Resource{
				Name:    res.Name,
				Version: res.Version,
				Type:    res.Type,
				Value:   spec.ImageReference,
			})
		case ResourceTypeHelmChartImageMap:
			var localBlob descriptorv2.LocalBlob
			if err := ociaccess.DefaultScheme.Convert(res.Access, &localBlob); err != nil {
				return nil, err
			}
			blob := blobs[ociaccess.NameVersionType{
				Name:    res.Name,
				Version: res.Version,
				Type:    res.Type,
			}]
			if blob == nil {
				return nil, fmt.Errorf("could not find local blob for resource %s:%s of type %s", res.Name, res.Version, res.Type)
			}
			resources = append(resources, Resource{
				Name:    res.Name,
				Version: res.Version,
				Type:    res.Type,
				Value:   string(blob),
			})
		}
	}

	return resources, nil
}

// AddComponentDependency adds a component and its dependency.
// It returns true if the dependency was newly added.
func (c *Components) AddComponentDependency(component ComponentReference, dependency Dependency) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if component == dependency.ComponentReference {
		return false // skip self-references
	}
	set, existing := c.dependents[dependency.ComponentReference]
	if !existing {
		set = sets.New(component)
		c.dependents[dependency.ComponentReference] = set
	} else {
		set.Insert(component)
	}
	c.dependencies[component] = append(c.dependencies[component], dependency)
	return !existing
}

// AddComponentDependencies adds a component and its dependencies from the given descriptor.
// It returns the newly added component references.
func (c *Components) AddComponentDependencies(descriptor *descriptorruntime.Descriptor, blobs Blobs) ([]ComponentReference, error) {
	if err := c.addComponent(descriptor, blobs); err != nil {
		return nil, err
	}

	component := descriptorToComponentReference(descriptor)
	dependencies := make(map[ComponentReference]*Dependency)
	for _, ref := range descriptor.Component.References {
		cref, imageSource, err := referenceToComponentReferenceWithImageSource(ref)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve component reference %q: %w", ref, err)
		}
		dependency := dependencies[cref]
		if dependency == nil {
			dependency = &Dependency{
				ComponentReference: cref,
			}
			dependencies[cref] = dependency
		}
		if imageSource != nil {
			dependency.ImageVector = append(dependency.ImageVector, *imageSource)
		}
	}

	var newComponents []ComponentReference
	for cref, dep := range dependencies {
		if c.AddComponentDependency(component, *dep) {
			newComponents = append(newComponents, cref)
		}
	}

	return newComponents, nil
}

// GetSortedComponents returns all components sorted by their reference.
func (c *Components) GetSortedComponents() []ComponentReference {
	c.lock.Lock()
	defer c.lock.Unlock()

	var sorted []ComponentReference
	for cref := range c.dependents {
		sorted = append(sorted, cref)
	}
	slices.Sort(sorted)
	return sorted
}

// GetResources returns all resources associated with the given component reference.
func (c *Components) GetResources(cref ComponentReference) []Resource {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.resources[cref]
}

// DumpComponentRefListAsYAML dumps all components and their versions as a YAML string.
func (c *Components) DumpComponentRefListAsYAML() (string, error) {
	var (
		names []string
		m     = make(map[string][]string)
	)
	for _, cref := range c.GetSortedComponents() {
		name, version, err := cref.ExtractNameAndVersion()
		if err != nil {
			return "", err
		}
		if _, exists := m[name]; !exists {
			names = append(names, name)
		}
		m[name] = append(m[name], version)
	}

	var sb strings.Builder
	sb.WriteString("components:\n")
	for _, name := range names {
		sb.WriteString("- name: " + name + "\n")
		sb.WriteString("  versions: [" + strings.Join(m[name], ", ") + "]\n")
	}
	return sb.String(), nil
}

// GetDependents returns all components that depend on the given component.
func (c *Components) GetDependents(cref ComponentReference) []ComponentReference {
	c.lock.Lock()
	defer c.lock.Unlock()

	set := c.dependents[cref]
	if set.Len() == 0 {
		return nil
	}
	array := set.UnsortedList()
	slices.Sort(array)
	return array
}

func (c *Components) getImageVector(cref ComponentReference, originalRef bool) ([]imagevector.ImageSource, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var images []imagevector.ImageSource
	for _, dep := range c.dependencies[cref] {
		for _, imgex := range dep.ImageVector {
			if imgex.LookupOnly {
				continue
			}
			img := imgex.ImageSource
			imgOrginalRef := imgex.OriginalRef
			if imgex.ReferencedComponent != nil {
				name := imgex.Name
				if imgex.ResourceID != nil {
					name = imgex.ResourceID.Name
				}
				found := false
			lookup:
				for _, dep := range c.dependencies[ComponentReference(*imgex.ReferencedComponent)] {
					for _, img2 := range dep.ImageVector {
						if img2.EffectiveResourceName() == name {
							img = img2.ImageSource
							imgOrginalRef = img2.OriginalRef
							img.Name = imgex.Name
							img.TargetVersion = imgex.TargetVersion
							found = true
							break lookup
						}
					}
				}
				if !found {
					var names []string
					for _, dep := range c.dependencies[ComponentReference(*imgex.ReferencedComponent)] {
						for _, img2 := range dep.ImageVector {
							names = append(names, img2.EffectiveResourceName())
						}
					}
					return nil, fmt.Errorf("could not find image %q in referenced component %q: [%s]", name, *imgex.ReferencedComponent, strings.Join(names, ", "))
				}
			}
			if img.Ref != nil && img.Repository != nil && img.Tag != nil {
				img.Ref = nil // avoid duplication
			}
			if originalRef && imgOrginalRef != nil {
				parts := strings.SplitN(*imgOrginalRef, ":", 2)
				if len(parts) != 2 {
					return nil, fmt.Errorf("could not split original reference %s", *imgOrginalRef)
				}
				img.Repository = &parts[0]
				img.Tag = &parts[1]
			}
			images = append(images, img)
		}
	}
	if len(images) == 0 {
		return nil, nil
	}

	SortImageSources(images)
	return images, nil
}

// GetImageVector returns the image vector for the given component reference.
// If originalRef is true, the original image references are used if available.
func (c *Components) GetImageVector(cref ComponentReference, originalRef bool) ([]imagevector.ImageSource, error) {
	images, err := c.getImageVector(cref, originalRef)
	if err != nil {
		return nil, err
	}
	mappedImages := c.mappedImages[cref]
	if len(mappedImages) > 0 {
		mappedNames := make(map[string]string)
		for _, img := range mappedImages {
			if img.Repository != nil {
				mappedNames[*img.Repository] = img.Name
			}
		}
		kubernetesCompRef, err := c.getKubernetesComponentRef()
		if err != nil {
			return nil, err
		}
		lssImages, err := c.getImageVector(kubernetesCompRef, originalRef)
		if err != nil {
			return nil, err
		}
		for _, image := range lssImages {
			if name, ok := mappedNames[image.Name]; ok {
				clone := image
				clone.Name = name
				clone.TargetVersion = image.Version
				images = append(images, clone)
			}
		}
	}

	SortImageSources(images)
	return images, nil
}

func (c *Components) getKubernetesComponentRef() (ComponentReference, error) {
	if c.kubernetesComponent != nil {
		return *c.kubernetesComponent, nil
	}
	return "", fmt.Errorf("could not determine kubernetes component reference")
}

// GetRootComponents returns all root components (those without any dependents).
func (c *Components) GetRootComponents() []ComponentReference {
	var roots []ComponentReference
	for _, dep := range c.dependencies[rootComponentReference] {
		roots = append(roots, dep.ComponentReference)
	}
	return roots
}

// ComponentsCount returns the number of components managed.
func (c *Components) ComponentsCount() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return len(c.dependents)
}

func referenceToComponentReferenceWithImageSource(ref descriptorruntime.Reference) (cref ComponentReference, img *ocmimagevector.ExtendedImageSource, err error) {
	cref = ComponentReference(fmt.Sprintf("%s:%s", ref.Component, ref.Version))
	for _, label := range ref.Labels {
		if label.Name == LabelImageVectorImages {
			img, err = rawToImageSource(label.Value)
			if img != nil {
				// needed for image overwrite lookup later
				img.ReferencedComponent = ptr.To(string(cref))
			}
			return
		}
	}
	return
}

func descriptorToComponentReference(descriptor *descriptorruntime.Descriptor) ComponentReference {
	return ComponentReference(fmt.Sprintf("%s:%s", descriptor.Component.Name, descriptor.Component.Version))
}

func rawToImageSource(value json.RawMessage) (*ocmimagevector.ExtendedImageSource, error) {
	images, err := rawToImageSources(value)
	if err != nil {
		return nil, err
	}
	switch len(images) {
	case 1:
		return images[0], nil
	case 0:
		return nil, nil
	default:
		return nil, fmt.Errorf("expected 1 or 0 images, got %d", len(images))
	}
}

func rawToImageSources(value json.RawMessage) ([]*ocmimagevector.ExtendedImageSource, error) {
	data, err := value.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if string(data) == "null" {
		return nil, nil
	}
	obj := &ocmimagevector.ExtendedImageVector{}
	err = json.Unmarshal(data, obj)
	if err != nil {
		return nil, err
	}
	return obj.Images, nil
}

func resourceToImageSource(res descriptorruntime.Resource) (*ocmimagevector.ExtendedImageSource, error) {
	if res.Type != resourceTypeOCIImage {
		return nil, nil
	}

	var (
		src ocmimagevector.ExtendedImageSource
		err error
	)
	for _, value := range res.Labels {
		switch value.Name {
		case labelNameImageVectorName:
			src.Name, err = toString(value.Value)
			src.ResourceName = res.Name
		case labelNameImageVectorRepository:
			src.Repository, err = toStringPtr(value.Value)
		case labelNameImageVectorSourceRepository:
			// ignore
		case labelNameImageVectorTargetVersion:
			src.TargetVersion, err = toStringPtr(value.Value)
		case labelNameCveCategorisation:
			src.Labels = append(src.Labels, value)
		case labelNameOriginalRef:
			src.OriginalRef, err = toStringPtr(value.Value)
		default:
			// ignore
		}
		if err != nil {
			return nil, fmt.Errorf("failed to convert label %s to string: %w", value.Name, err)
		}
	}
	if src.Name == "" {
		src.Name = res.Name
		src.LookupOnly = true
	}
	var spec accessv1.OCIImage
	if err := ociaccess.DefaultScheme.Convert(res.Access, &spec); err != nil {
		return nil, err
	}
	src.Ref = &spec.ImageReference
	repo, tag, err := splitRef(spec.ImageReference)
	if err != nil {
		return nil, fmt.Errorf("failed to split ref %s: %w", spec.ImageReference, err)
	}
	src.Repository = &repo
	src.Tag = &tag
	if res.Version != "" {
		src.Version = &res.Version
	}
	return &src, nil
}

func splitRef(ref string) (string, string, error) {
	parts1 := strings.SplitN(ref, ":", 2)
	parts2 := strings.SplitN(ref, "@", 2)
	var parts []string
	if len(parts1) == 2 {
		parts = parts1
	}
	if len(parts2) == 2 && len(parts) == 2 && len(parts2[1]) > len(parts[1]) {
		parts = parts2
	}
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid reference format")
	}
	return parts[0], parts[1], nil
}

func toString(value json.RawMessage) (string, error) {
	ps, err := toStringPtr(value)
	return ptr.Deref(ps, ""), err
}

func toStringPtr(value json.RawMessage) (*string, error) {
	data, err := value.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if string(data) == "null" {
		return nil, nil
	}
	var s string
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// SortImageSources sorts image sources by name and target version.
func SortImageSources(images []imagevector.ImageSource) {
	slices.SortFunc(images, func(a, b imagevector.ImageSource) int {
		cmp := strings.Compare(a.Name, b.Name)
		if cmp != 0 {
			return cmp
		}
		return strings.Compare(ptr.Deref(a.TargetVersion, ""), ptr.Deref(b.TargetVersion, ""))
	})
}
