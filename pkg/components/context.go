// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"fmt"

	"github.com/go-logr/logr"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

// UpgradePath contains information about the upgrade path of a component.
type UpgradePath struct {
	// CurrentVersion is the currently installed version of the component.
	CurrentVersion string
	// NextVersion is the next version of the component to upgrade to.
	NextVersion string
}

// ComponentContext is the interface that providers context information for components.
type ComponentContext interface {
	// GetUpgradePath returns the upgrade path of the component.
	GetUpgradePath() UpgradePath
}

type componentContext struct {
	UpgradePath UpgradePath
}

// GetUpgradePath returns the upgrade path of the component.
func (c *componentContext) GetUpgradePath() UpgradePath {
	return c.UpgradePath
}

// newComponentContext creates a new componentContext instance for the given component.
func newComponentContext(currentVersion, nextVersion string) ComponentContext {
	return &componentContext{
		UpgradePath: UpgradePath{
			CurrentVersion: currentVersion,
			NextVersion:    nextVersion,
		},
	}
}

// Context is the interface that provides context information for components.
type Context interface {
	// Own is the context for the own component.
	Own(metadataInterface MetadataInterface) (ComponentContext, error)
	// FindComponentContext finds the context for the given component name.
	FindComponentContext(string) (ComponentContext, bool)
}

// ComponentsContext is the implementation of the Context interface that provides context information for components.
type ComponentsContext struct {
	componentToContext map[string]ComponentContext
}

// Own is the context for the own component.
func (c *ComponentsContext) Own(metadataInterface MetadataInterface) (ComponentContext, error) {
	componentName := metadataInterface.GetComponentMetadata().Name
	componentContext, ok := c.componentToContext[componentName]
	if !ok {
		return nil, fmt.Errorf("no component context found for '%s'", componentName)
	}
	return componentContext, nil
}

// FindComponentContext finds the context for the given component name.
func (c *ComponentsContext) FindComponentContext(name string) (ComponentContext, bool) {
	componentContext, ok := c.componentToContext[name]
	return componentContext, ok
}

// AddComponentContext adds a component context for the given component name and current and next component vectors.
func (c *ComponentsContext) AddComponentContext(log logr.Logger, component MetadataInterface, currentVector, nextVector componentvector.Interface) error {
	componentName := component.GetComponentMetadata().Name
	if _, ok := c.componentToContext[componentName]; ok {
		return fmt.Errorf("component context for component %s already exists", componentName)
	}

	var (
		currentVersion string
		nextVersion    string
	)

	if componentRef := component.GetComponentMetadata().ComponentRef; componentRef == nil {
		log.V(1).Info("component has no componentRef", "component", componentName)
	} else {
		if currentVector != nil {
			currentVersion, _ = currentVector.FindComponentVersion(*componentRef)
		}
		nextVersion, _ = nextVector.FindComponentVersion(*componentRef)
	}

	c.componentToContext[componentName] = newComponentContext(currentVersion, nextVersion)
	return nil
}

// NewContext creates a new context implementation for components.
func NewContext() *ComponentsContext {
	return &ComponentsContext{
		componentToContext: make(map[string]ComponentContext),
	}
}
