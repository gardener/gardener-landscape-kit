// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	_ "embed"

	"sigs.k8s.io/yaml"
)

const (
	// DirName is the directory name where components are stored.
	DirName = "components"
)

// Interface is the components interface that each component must implement.
type Interface interface {
	MetadataInterface
	// GenerateBase generates the component base dir.
	GenerateBase(Context, Options) error
	// GenerateLandscape generates the component landscape dir.
	GenerateLandscape(Context, LandscapeOptions) error
}

// MetadataInterface is the interface that each component must implement to provide metadata information.
type MetadataInterface interface {
	// GetComponentMetadata returns the component metadata.
	GetComponentMetadata() *Metadata
}

// Metadata contains metadata information for a component.
type Metadata struct {
	// Name is the component name.
	Name string `json:"name"`
	// Directory is the directory, where the components store their generated files. It is relative to the base or landscape target directory.
	Directory string `json:"directory"`
	// ComponentRef is the component reference to a component in the component vector.
	ComponentRef *string `json:"componentRef,omitempty"`
}

// GetComponentMetadata returns the component metadata.
func (m *Metadata) GetComponentMetadata() *Metadata { return m }

// NewMetadata creates a new Metadata instance from the given YAML bytes.
func NewMetadata(yamlBytes []byte) (*Metadata, error) {
	metadata := &Metadata{}
	if err := yaml.Unmarshal(yamlBytes, metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}
