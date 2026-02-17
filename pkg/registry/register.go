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
	alicloud "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-alicloud"
	aws "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-aws"
	azure "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-azure"
	gcp "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-gcp"
	openstack "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-openstack"
	gvisor "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/runtime-gvisor"
	certservice "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/shoot-cert-service"
	dnsservice "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/shoot-dns-service"
	networkingproblemdetector "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/shoot-networking-problemdetector"
	oidcservice "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/shoot-oidc-service"
	"github.com/gardener/gardener-landscape-kit/pkg/components/gardener/garden"
	"github.com/gardener/gardener-landscape-kit/pkg/components/gardener/operator"
	virtualgardenaccess "github.com/gardener/gardener-landscape-kit/pkg/components/gardener/virtual-garden-access"
	gardenproject "github.com/gardener/gardener-landscape-kit/pkg/components/virtual-garden/garden-project"
)

// ComponentList contains all available components.
var ComponentList = []func() components.Interface{
	flux.NewComponent,
	operator.NewComponent,
	garden.NewComponent,
	calico.NewComponent,
	cilium.NewComponent,
	alicloud.NewComponent,
	aws.NewComponent,
	azure.NewComponent,
	gcp.NewComponent,
	openstack.NewComponent,
	gardenlinux.NewComponent,
	suse.NewComponent,
	certservice.NewComponent,
	dnsservice.NewComponent,
	oidcservice.NewComponent,
	networkingproblemdetector.NewComponent,
	gvisor.NewComponent,
	virtualgardenaccess.NewComponent,
	gardenproject.NewComponent,
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
