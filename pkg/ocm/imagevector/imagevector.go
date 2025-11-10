// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package imagevector

import (
	"github.com/gardener/gardener/pkg/utils/imagevector"
	descriptorruntime "ocm.software/open-component-model/bindings/go/descriptor/runtime"
)

// ExtendedImageSource is an extended image source that includes additional metadata for internal processing.
type ExtendedImageSource struct {
	imagevector.ImageSource

	Labels       []descriptorruntime.Label
	ResourceName string
	ResourceID   *ResourceID
	// ReferencedComponent is the referenced component (if any) from which this image originates.
	ReferencedComponent *string
	LookupOnly          bool
	OriginalRef         *string
}

// EffectiveResourceName returns the ResourceName if set, otherwise the Name.
func (img *ExtendedImageSource) EffectiveResourceName() string {
	if img.ResourceName != "" {
		return img.ResourceName
	}
	return img.Name
}

// ResourceID represents a resource identifier with a name.
type ResourceID struct {
	Name string `json:"name"`
}

// ExtendedImageVector is an extended image vector format that includes additional metadata for each image source.
type ExtendedImageVector struct {
	Images []*ExtendedImageSource `json:"images"`
}

// ImageVectorOutput is the output format for the image vector JSON output.
type ImageVectorOutput struct {
	Images []imagevector.ImageSource `json:"images"`
}
