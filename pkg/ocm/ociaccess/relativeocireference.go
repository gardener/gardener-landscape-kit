// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package ociaccess

import (
	ocmruntime "ocm.software/open-component-model/bindings/go/runtime"
)

// RelativeOciReferenceTypeName is the name of the custom type used for relative OCI references in component descriptors.
const RelativeOciReferenceTypeName = "relativeOciReference"

// RelativeOciReference is the 'relativeOciReference' access type used to represent OCI references relative to a repository in component descriptors.
type RelativeOciReference struct {
	Type      ocmruntime.Type `json:"type"`
	Reference string          `json:"reference"`
}

// GetType returns the type of the RelativeOciReference.
func (t *RelativeOciReference) GetType() ocmruntime.Type {
	return t.Type
}

// SetType sets the type of the RelativeOciReference.
func (t *RelativeOciReference) SetType(typ ocmruntime.Type) {
	t.Type = typ
}

// DeepCopyTyped returns a deep copy of the RelativeOciReference as an [ocmruntime.Typed].
func (t *RelativeOciReference) DeepCopyTyped() ocmruntime.Typed {
	return &RelativeOciReference{
		Type:      t.Type,
		Reference: t.Reference,
	}
}
