// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

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
	SourceRepository string `json:"sourceRepository"`
	// Version is the version of the component.
	Version string `json:"version"`
}

// Interface is a marker interface for component vectors.
type Interface interface {
	// FindComponentVersion finds the version of the component with the given name.
	FindComponentVersion(string) (string, bool)
}
