// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/components/flux"
)

// RegisterAllComponents registers all available components.
func RegisterAllComponents(registry Interface) {
	for _, component := range []components.Interface{
		flux.NewComponent(),
	} {
		registry.RegisterComponent(component.Name(), component)
	}
}
