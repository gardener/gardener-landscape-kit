// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"slices"
	"strings"

	"github.com/elliotchance/orderedmap/v3"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/components/flux"
	"github.com/gardener/gardener-landscape-kit/pkg/components/gardener/garden"
	"github.com/gardener/gardener-landscape-kit/pkg/components/gardener/operator"
)

// ComponentList contains all available components.
var ComponentList = []func() components.Interface{
	flux.NewComponent,
	operator.NewComponent,
	garden.NewComponent,
}

// RegisterAllComponents registers all available components.
func RegisterAllComponents(registry Interface, config *v1alpha1.LandscapeKitConfiguration) error {
	var excludedComponents []string
	if config != nil && config.Components != nil {
		excludedComponents = slices.Clone(config.Components.Exclude)
	}

	orderedComponents := orderedmap.NewOrderedMap[string, components.Interface]()
	for _, newComponent := range ComponentList {
		component := newComponent()
		orderedComponents.Set(component.Name(), component)
	}

	var invalidComponentNames []string
	for _, excludedComponent := range excludedComponents {
		if _, ok := orderedComponents.Get(excludedComponent); !ok {
			invalidComponentNames = append(invalidComponentNames, excludedComponent)
		} else {
			orderedComponents.Delete(excludedComponent)
		}
	}

	if len(invalidComponentNames) > 0 {
		return fmt.Errorf(
			"configuration contains invalid component excludes: %s - available component names are: %s",
			strings.Join(invalidComponentNames, ", "),
			strings.Join(slices.Collect(orderedComponents.Keys()), ", "),
		)
	}

	for _, component := range orderedComponents.AllFromFront() {
		registry.RegisterComponent(component.Name(), component)
	}

	return nil
}
