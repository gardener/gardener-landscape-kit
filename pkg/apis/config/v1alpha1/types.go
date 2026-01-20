// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LandscapeKitConfiguration contains configuration for the Gardener Landscape Kit.
type LandscapeKitConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// OCM is the configuration for the OCM version processing.
	// +optional
	OCM *OCMConfig `json:"ocm,omitempty"`
	// Git is the configuration for the Git repository.
	// +optional
	Git *GitRepository `json:"git,omitempty"`
	// Components is the configuration for the components.
	// +optional
	Components *ComponentsConfiguration `json:"components,omitempty"`
	// VersionConfig is the configuration for versioning.
	// +optional
	VersionConfig *VersionConfiguration `json:"versionConfig,omitempty"`
}

// ComponentsConfiguration contains configuration for components.
type ComponentsConfiguration struct {
	// Exclude is a list of component names to exclude.
	// +optional
	Exclude []string `json:"exclude,omitempty"`
}

// GitRepository contains information the Git repository containing landscape deployments and configurations.
type GitRepository struct {
	// URL specifies the Git repository URL, it can be an HTTP/S or SSH address.
	// +required
	URL string `json:"url"`
	// Reference specifies the Git reference to resolve and monitor for
	// changes, defaults to the 'master' branch.
	// +required
	Ref GitRepositoryRef `json:"ref"`
	// Paths specifies the path configuration within the Git repository.
	// +required
	Paths PathConfiguration `json:"paths"`
}

// PathConfiguration contains path configuration within the Git repository.
type PathConfiguration struct {
	// Base is the relative path to the base directory within the Git repository.
	// +required
	Base string `json:"base"`
	// Landscape is the relative path to the landscape directory within the Git repository.
	// +required
	Landscape string `json:"landscape"`
}

// GitRepositoryRef specifies the Git reference to resolve and checkout.
type GitRepositoryRef struct {
	// Branch to check out, defaults to 'main' if no other field is defined.
	// +optional
	Branch *string `json:"branch,omitempty"`
	// Tag to check out, takes precedence over Branch.
	// +optional
	Tag *string `json:"tag,omitempty"`
	// Commit SHA to check out, takes precedence over all reference fields.
	// +optional
	Commit *string `json:"commit,omitempty"`
}

// OCMConfig contains information about root component.
type OCMConfig struct {
	// Repositories is a map from repository name to URL.
	Repositories []string `json:"repositories"`
	// RootComponent is the configuration of the root component.
	RootComponent OCMComponent `json:"rootComponent"`
	// OriginalRefs is a flag to output original image references in the image vectors.
	OriginalRefs bool `json:"originalRefs"`
}

// OCMComponent specifies a OCM component.
type OCMComponent struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// String returns the string representation of the OCM component.
func (nv *OCMComponent) String() string {
	return nv.Name + ":" + nv.Version
}

const (
	// DefaultVersionsUpdateStrategyReleaseBranch indicates that the versions in the default vector should be updated from the release branch on generate.
	DefaultVersionsUpdateStrategyReleaseBranch = "ReleaseBranch"
	// DefaultVersionsUpdateStrategyDisabled indicates that the versions in the default vector should not be updated on generate.
	DefaultVersionsUpdateStrategyDisabled = "Disabled"
)

// AllowedDefaultVersionsUpdateStrategies lists all allowed strategies for updating versions in the default components vector.
var AllowedDefaultVersionsUpdateStrategies = []string{DefaultVersionsUpdateStrategyReleaseBranch, DefaultVersionsUpdateStrategyDisabled}

// VersionConfiguration contains configuration for versioning.
type VersionConfiguration struct {
	// ComponentsVectorFile is the path to the components vector file. A default vector is applied if not specified.
	// +optional
	ComponentsVectorFile *string `json:"componentsVectorFile,omitempty"`
	// UpdateStrategy determines whether the versions in the default vector should be updated from the release branch on generate.
	// Possible values are "Disabled" (default) and "ReleaseBranch".
	// Only used if no ComponentsVectorFile is specified.
	// +optional
	DefaultVersionsUpdateStrategy *string `json:"defaultVersionsUpdateStrategy,omitempty"`
}
