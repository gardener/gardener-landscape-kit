// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"fmt"
	"path/filepath"
)

// Registry is the interface for a component registry.
type Registry interface {
	// RegisterComponent registers a component in the registry.
	RegisterComponent(component Interface)
	// GenerateBase generates the base component.
	GenerateBase(opts Options) error
	// GenerateLandscape generates the landscape component.
	GenerateLandscape(opts LandscapeOptions) error
}

type registry struct {
	components []Interface
}

// RegisterComponent registers a component in the registry.
func (r *registry) RegisterComponent(component Interface) {
	r.components = append(r.components, component)
}

// GenerateBase generates the base component.
func (r *registry) GenerateBase(opts Options) error {
	if err := opts.GetFilesystem().MkdirAll(filepath.Join(opts.GetTargetPath(), DirName), 0700); err != nil {
		return fmt.Errorf("cannot create components directory: %w", err)
	}

	for _, component := range r.components {
		if err := component.GenerateBase(opts); err != nil {
			return err
		}
	}

	return nil
}

// GenerateLandscape generates the landscape component.
func (r *registry) GenerateLandscape(opts LandscapeOptions) error {
	if err := opts.GetFilesystem().MkdirAll(filepath.Join(opts.GetTargetPath(), DirName), 0700); err != nil {
		return fmt.Errorf("cannot create components directory: %w", err)
	}

	for _, component := range r.components {
		if err := component.GenerateLandscape(opts); err != nil {
			return err
		}
	}

	return writeLandscapeComponentsKustomizations(opts)
}

// NewRegistry creates a new component registry.
func NewRegistry() Registry {
	return &registry{
		components: []Interface{},
	}
}
