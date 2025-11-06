// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LandscapeKitConfiguration contains configuration for the Gardener Landscape Kit.
type LandscapeKitConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// OCM is the configuration for the OCM version processing.
	// +optional
	OCM *OCMConfig `json:"ocm,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OCMConfiguration contains information about root component.
type OCMConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	*OCMConfig `json:",inline"`
}

// OCMConfig contains information about root component.
type OCMConfig struct {
	// Repositories is a map from repository name to URL.
	Repositories []string `json:"repositories"`
	// RootComponent is the configuration of the root component.
	RootComponent OCMComponent `json:"rootComponent"`
	// OriginalRefs is a flag to output original image references in the image vectors.
	OriginalRefs bool `json:"originalRefs"`
}

// OCMComponent specifies a OCM component.
type OCMComponent struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// String returns the string representation of the OCM component.
func (nv *OCMComponent) String() string {
	return nv.Name + ":" + nv.Version
}
