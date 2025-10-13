// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package config

// Config contains information about root component and repositories.
type Config struct {
	// Repositories is a map from repository name to URL
	Repositories []string `json:"repositories"`
	// RootComponent is the name and version of the root component
	RootComponent NameAndVersion `json:"rootComponent"`
	// OriginalRefs is a flag to output original image references in the image vectors.
	OriginalRefs bool `json:"originalRefs,omitempty"`
}

// NameAndVersion specifies a OCM component.
type NameAndVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (nv NameAndVersion) String() string {
	return nv.Name + ":" + nv.Version
}
