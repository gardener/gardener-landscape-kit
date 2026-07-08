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
	// Repositories is the configuration for the base and landscape repositories.
	// All paths inside each section are relative to that repository's root.
	// +optional
	Repositories *RepositoriesConfig `json:"repositories,omitempty"`
	// Components is the configuration for the components.
	// +optional
	Components *ComponentsConfiguration `json:"components,omitempty"`
	// VersionConfig is the configuration for versioning.
	// +optional
	VersionConfig *VersionConfiguration `json:"versionConfig,omitempty"`
	// MergeMode determines how merge conflicts are resolved:
	// - "Hint" (default): New default values from GLK are added as comments after any customized values.
	// - "Silent": Operator-customized values are retained, new default values are omitted.
	// +optional
	MergeMode *MergeMode `json:"mergeMode,omitempty"`
}

// ComponentsConfiguration contains configuration for components.
type ComponentsConfiguration struct {
	// Exclude is a list of component names to exclude.
	// +optional
	Exclude []string `json:"exclude,omitempty"`
	// Include is a list of component names to include.
	// +optional
	Include []string `json:"include,omitempty"`
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

// RepositoriesConfig describes the base and landscape repositories.
// All paths inside each section are relative to that repository's root.
type RepositoriesConfig struct {
	// Base configures the base repository.
	// +optional
	Base *BaseRepositoryConfig `json:"base,omitempty"`
	// Landscape configures the landscape repository.
	// +optional
	Landscape *LandscapeRepositoryConfig `json:"landscape,omitempty"`
}

// BaseRepositoryConfig configures the base repository.
type BaseRepositoryConfig struct {
	// Target is the directory of the base content within the base repository.
	// Defaults to "./" if not specified.
	// +optional
	Target string `json:"target,omitempty"`
	// ComponentsYamlOverrides lists additional components.yaml files layered on top of the in-repo base components.yaml.
	// Applied in declared order; later entries win.
	// +optional
	ComponentsYamlOverrides []string `json:"componentsYamlOverrides,omitempty"`
}

// LandscapeRepositoryConfig configures the landscape repository.
type LandscapeRepositoryConfig struct {
	// URL of the landscape Git repository (http/s or ssh).
	// +required
	URL string `json:"url"`
	// Ref to check out (branch / tag / commit).
	// +required
	Ref GitRepositoryRef `json:"ref"`
	// BaseLink is the path inside the landscape repository where the base repository's root is mounted (e.g. via a Git submodule).
	// The base content is located at path.Join(baseLink, repositories.base.target).
	// +required
	BaseLink string `json:"baseLink"`
	// Target is the landscape directory within the landscape repository.
	// Defaults to "./" if not specified.
	// +optional
	Target string `json:"target,omitempty"`
	// ComponentsYamlOverrides lists additional components.yaml files layered on top of the in-repo landscape components.yaml during `generate landscape`.
	// Applied in declared order; later entries win.
	// +optional
	ComponentsYamlOverrides []string `json:"componentsYamlOverrides,omitempty"`
}

// OCMConfig contains information about root component.
type OCMConfig struct {
	// Repositories is a map from repository name to URL.
	Repositories []string `json:"repositories"`
	// RootComponent is the configuration of the root component.
	RootComponent OCMComponent `json:"rootComponent"`
	// OriginalRefs is a flag to output original image references in the image vectors.
	OriginalRefs bool `json:"originalRefs"`
	// IgnoreMissingComponents indicates whether to ignore missing components during resolution.
	// +optional
	IgnoreMissingComponents *bool `json:"ignoreMissingComponents,omitempty"`
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

// DefaultVersionsUpdateStrategy controls whether the versions in the default components vector should be updated from the release branch on generate.
type DefaultVersionsUpdateStrategy string

const (
	// DefaultVersionsUpdateStrategyReleaseBranch indicates that the versions in the default vector should be updated from the release branch on generate.
	DefaultVersionsUpdateStrategyReleaseBranch DefaultVersionsUpdateStrategy = "ReleaseBranch"
	// DefaultVersionsUpdateStrategyDisabled indicates that the versions in the default vector should not be updated on generate.
	DefaultVersionsUpdateStrategyDisabled DefaultVersionsUpdateStrategy = "Disabled"
)

// AllowedDefaultVersionsUpdateStrategies lists all allowed strategies for updating versions in the default components vector.
var AllowedDefaultVersionsUpdateStrategies = []string{
	string(DefaultVersionsUpdateStrategyReleaseBranch),
	string(DefaultVersionsUpdateStrategyDisabled),
}

// VersionCheckMode controls the behavior when the tool version doesn't match the component version.
type VersionCheckMode string

const (
	// VersionCheckModeStrict indicates that version mismatches should cause an error.
	VersionCheckModeStrict VersionCheckMode = "Strict"
	// VersionCheckModeWarning indicates that version mismatches should only log a warning.
	VersionCheckModeWarning VersionCheckMode = "Warning"
)

// AllowedVersionCheckModes lists all allowed version check modes.
var AllowedVersionCheckModes = []string{
	string(VersionCheckModeStrict),
	string(VersionCheckModeWarning),
}

// VersionConfiguration contains configuration for versioning.
type VersionConfiguration struct {
	// UpdateStrategy determines whether the versions in the default vector should be updated from the release branch on resolve.
	// Possible values are "Disabled" (default) and "ReleaseBranch".
	// +optional
	DefaultVersionsUpdateStrategy *DefaultVersionsUpdateStrategy `json:"defaultVersionsUpdateStrategy,omitempty"`
	// CheckMode determines the behavior when the tool version doesn't match the gardener-landscape-kit version in the component vector.
	// Possible values are "Strict" (default) and "Warning".
	// In strict mode, version mismatches cause errors. In warning mode, only warnings are logged.
	// +optional
	CheckMode *VersionCheckMode `json:"checkMode,omitempty"`
}

// MergeMode controls how operator overwrites are handled during three-way merge.
type MergeMode string

const (
	// MergeModeHint annotates operator-overwritten values with a comment showing the current GLK default.
	MergeModeHint MergeMode = "Hint"
	// MergeModeSilent retains operator overwrites without annotation.
	MergeModeSilent MergeMode = "Silent"
)

// AllowedMergeModes lists all allowed merge modes.
var AllowedMergeModes = []string{
	string(MergeModeHint),
	string(MergeModeSilent),
}
