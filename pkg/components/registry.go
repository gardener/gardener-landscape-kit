// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

// Registry is the interface for a component registry.
type Registry interface {
	// RegisterComponent registers a component in the registry.
	RegisterComponent(component Interface)
	// Generate generates all registered components.
	Generate(opts Options) error
}

type registry struct {
	components []Interface
}

// RegisterComponent registers a component in the registry.
func (r *registry) RegisterComponent(component Interface) {
	r.components = append(r.components, component)
}

// Generate generates all registered components. Generation happens serially in the order of registration.
func (r *registry) Generate(opts Options) error {
	if opts.GetLandscapeDir() == "" {
		return r.generateBase(opts)
	}
	return r.generateLandscape(opts)
}

func (r *registry) generateBase(opts Options) error {
	for _, component := range r.components {
		if err := component.GenerateBase(opts); err != nil {
			return err
		}
	}
	return nil
}

func (r *registry) generateLandscape(opts Options) error {
	for _, component := range r.components {
		if err := component.GenerateLandscape(opts); err != nil {
			return err
		}
	}
	return nil
}

// NewRegistry creates a new component registry.
func NewRegistry() Registry {
	return &registry{
		components: []Interface{},
	}
}
