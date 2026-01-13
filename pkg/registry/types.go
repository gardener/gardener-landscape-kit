// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"github.com/elliotchance/orderedmap/v3"

	"github.com/gardener/gardener-landscape-kit/pkg/components"
)

// Interface is the interface for a component registry.
type Interface interface {
	// RegisterComponent registers a component in the registry.
	RegisterComponent(name string, component components.Interface)
	// GenerateBase generates the base component.
	GenerateBase(opts components.Options) error
	// GenerateLandscape generates the landscape component.
	GenerateLandscape(opts components.LandscapeOptions) error
}

type registry struct {
	components *orderedmap.OrderedMap[string, components.Interface]
}

// RegisterComponent registers a component in the registry.
func (r *registry) RegisterComponent(name string, component components.Interface) {
	r.components.Set(name, component)
}

// GenerateBase generates the base component.
func (r *registry) GenerateBase(opts components.Options) error {
	for _, component := range r.components.AllFromFront() {
		if err := component.GenerateBase(opts); err != nil {
			return err
		}
	}

	return nil
}

// GenerateLandscape generates the landscape component.
func (r *registry) GenerateLandscape(opts components.LandscapeOptions) error {
	for _, component := range r.components.AllFromFront() {
		if err := component.GenerateLandscape(opts); err != nil {
			return err
		}
	}

	return nil
}

// New creates a new component registry.
func New() Interface {
	return &registry{
		components: orderedmap.NewOrderedMap[string, components.Interface](),
	}
}
