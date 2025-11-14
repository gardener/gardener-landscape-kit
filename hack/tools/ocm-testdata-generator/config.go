// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"sigs.k8s.io/yaml"
)

// Config holds the configuration for generating test resources.
type Config struct {
	// RootComponent is the main component to be generated.
	RootComponent Component `json:"rootComponent"`
	// KubernetesVersions contains the list of Kubernetes versions to generate resources for.
	KubernetesVersions []string `json:"kubernetesVersions"`
	// ExternalResources is a list of external resources to be added to the root component.
	ExternalResources []ExternalResource `json:"externalResources"`
	// Components is a list of additional components to be added to the root component.
	Components []Component `json:"components"`
	// TargetRepositoryURL is the URL of the target repository containing the overwrites.
	TargetRepositoryURL string `json:"targetRepositoryURL"`
}

// Component represents a component to be included in the configuration.
type Component struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	ComponentName string `json:"componentName"`
}

// ExternalResource represents an external resource to be included in the configuration.
type ExternalResource struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	RepositoryURL string `json:"repositoryURL"`
}

// LoadConfig loads the configuration from the specified file path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
