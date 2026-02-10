// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"slices"
	"strings"

	"github.com/elliotchance/orderedmap/v3"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/components/flux"
	calico "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/networking-calico"
	cilium "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/networking-cilium"
	gardenlinux "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/os-gardenlinux"
	suse "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/os-suse-chost"
	aws "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-aws"
	azure "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-azure"
	"github.com/gardener/gardener-landscape-kit/pkg/components/gardener/garden"
	"github.com/gardener/gardener-landscape-kit/pkg/components/gardener/operator"
)

// ComponentList contains all available components.
var ComponentList = []func() components.Interface{
	flux.NewComponent,
	operator.NewComponent,
	garden.NewComponent,
	calico.NewComponent,
	cilium.NewComponent,
	aws.NewComponent,
	azure.NewComponent,
	gardenlinux.NewComponent,
	suse.NewComponent,
}

// RegisterAllComponents registers all available components.
func RegisterAllComponents(registry Interface, config *v1alpha1.LandscapeKitConfiguration) error {
	orderedComponents := orderedmap.NewOrderedMap[string, components.Interface]()
	for _, newComponent := range ComponentList {
		component := newComponent()
		orderedComponents.Set(component.Name(), component)
	}

	if err := excludeComponents(config, orderedComponents); err != nil {
		return err
	}

	if err := includeComponents(config, orderedComponents); err != nil {
		return err
	}

	for _, component := range orderedComponents.AllFromFront() {
		registry.RegisterComponent(component.Name(), component)
	}

	return nil
}

func excludeComponents(config *v1alpha1.LandscapeKitConfiguration, orderedComponents *orderedmap.OrderedMap[string, components.Interface]) error {
	excludedComponents := sets.New[string]()
	if config != nil && config.Components != nil {
		excludedComponents = excludedComponents.Insert(config.Components.Exclude...)
	}
	if excludedComponents.Len() == 0 {
		return nil
	}

	availableComponents := slices.Collect(orderedComponents.Keys())
	invalidComponentNames := excludedComponents.Difference(sets.New(availableComponents...))
	if len(invalidComponentNames) > 0 {
		return fmt.Errorf(
			"configuration contains invalid component excludes: %s - available component names are: %s",
			strings.Join(invalidComponentNames.UnsortedList(), ", "),
			strings.Join(availableComponents, ", "),
		)
	}

	for _, excludedComponent := range excludedComponents.UnsortedList() {
		orderedComponents.Delete(excludedComponent)
	}

	return nil
}

func includeComponents(config *v1alpha1.LandscapeKitConfiguration, orderedComponents *orderedmap.OrderedMap[string, components.Interface]) error {
	includedComponents := sets.New[string]()
	if config != nil && config.Components != nil {
		includedComponents = includedComponents.Insert(config.Components.Include...)
	}
	if includedComponents.Len() == 0 {
		return nil
	}

	availableComponents := slices.Collect(orderedComponents.Keys())
	invalidComponentNames := includedComponents.Difference(sets.New(availableComponents...))
	if len(invalidComponentNames) > 0 {
		return fmt.Errorf(
			"configuration contains invalid component includes: %s - available component names are: %s",
			strings.Join(invalidComponentNames.UnsortedList(), ", "),
			strings.Join(availableComponents, ", "),
		)
	}

	for _, componentName := range availableComponents {
		if !includedComponents.Has(componentName) {
			orderedComponents.Delete(componentName)
		}
	}

	return nil
}
