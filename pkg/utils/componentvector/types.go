// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import "github.com/gardener/gardener/pkg/utils/imagevector"

// ComponentVectorFilename is the default filename for component vectors to be placed at the root of either a base or landscape directory.
const ComponentVectorFilename = "components.yaml"

// Components is a list of components.
type Components struct {
	// Components is the list of component vectors.
	Components []*ComponentVector `json:"components,omitempty"`
}

// ComponentVector contains component versions and other related metadata of a component.
type ComponentVector struct {
	// Name is the name of the component.
	Name string `json:"name"`
	// SourceRepository is the source repository of the component.
	SourceRepository *string `json:"sourceRepository,omitempty"`
	// Version is the version of the component.
	Version string `json:"version"`
	// Resources contains additional data for component resources like OCI image references and Helm chart references.
	Resources map[string]ResourceData `json:"resources,omitempty"`
	// ImageVectorOverwrite is an optional image vector overwrite for the component.
	ImageVectorOverwrite *ImageVectorOverwrite `json:"imageVectorOverwrite,omitempty"`
	// ComponentImageVectorOverwrites are optional component image vector overwrites for components deployed with this component.
	ComponentImageVectorOverwrites *ComponentImageVectorOverwrites `json:"componentImageVectorOverwrites,omitempty"`
}

// ImageVectorOverwrite is the list of image sources that overwrite the default image vector for a component.
type ImageVectorOverwrite struct {
	Images []imagevector.ImageSource `json:"images"`
}

// ComponentImageVectorOverwrites is list of ComponentImageVectorOverwrite.
type ComponentImageVectorOverwrites struct {
	Components []ComponentImageVectorOverwrite `json:"components"`
}

// ComponentImageVectorOverwrite is the named ImageVectorOverwrite for a subcomponent.
type ComponentImageVectorOverwrite struct {
	Name                 string               `json:"name"`
	ImageVectorOverwrite ImageVectorOverwrite `json:"imageVectorOverwrite"`
}

// ResourceData contains additional data for component resources like OCI image references and Helm chart references.
type ResourceData struct {
	// OCIImage is the OCI image reference of the component resource.
	OCIImage *OCIImage `json:"ociImage,omitempty"`
	// HelmChart is the Helm chart reference and image map of the component resource.
	HelmChart *HelmChart `json:"helmChart,omitempty"`
}

// OCIImage contains the OCI image reference of a component resource.
type OCIImage struct {
	// Ref is the full artifact Ref and takes precedence over all other fields.
	Ref *string `json:"ref,omitempty"`
	// Repository is a reference to an OCI artifact repository.
	Repository *string `json:"repository,omitempty"`
	// Tag is the image tag to pull.
	Tag *string `json:"tag,omitempty"`
}

// HelmChart contains the Helm chart reference and image map of a component resource.
type HelmChart struct {
	// Ref is the full artifact Ref and takes precedence over all other fields.
	Ref *string `json:"ref,omitempty"`
	// Repository is a reference to an OCI artifact repository.
	Repository *string `json:"repository,omitempty"`
	// Tag is the artifact tag to pull.
	Tag *string `json:"tag,omitempty"`
	// ImageMap is the map of image names to helm chart values for overwriting OCI image repository and tag.
	ImageMap map[string]any `json:"imageMap,omitempty"`
}

// Interface is a marker interface for component vectors.
type Interface interface {
	// FindComponentVersion finds the version of the component with the given name.
	FindComponentVersion(string) (string, bool)
	// FindComponentVector finds the ComponentVector of the component with the given name.
	FindComponentVector(string) *ComponentVector
	// ComponentNames returns the sorted list of component names in the component vector.
	ComponentNames() []string
}
