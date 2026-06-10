// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// SetDefaults_LandscapeKitConfiguration sets default values for LandscapeKitConfiguration fields.
func SetDefaults_LandscapeKitConfiguration(obj *LandscapeKitConfiguration) {
	if obj.VersionConfig == nil {
		obj.VersionConfig = &VersionConfiguration{}
	}

	if obj.VersionConfig.DefaultVersionsUpdateStrategy == nil {
		obj.VersionConfig.DefaultVersionsUpdateStrategy = new(DefaultVersionsUpdateStrategyDisabled)
	}

	if obj.VersionConfig.CheckMode == nil {
		obj.VersionConfig.CheckMode = new(VersionCheckModeStrict)
	}

	if obj.MergeMode == nil {
		obj.MergeMode = new(MergeModeHint)
	}
}

// SetDefaults_OCMConfig sets defaults for OCMConfig.
func SetDefaults_OCMConfig(obj *OCMConfig) {
	if obj.IgnoreMissingComponents == nil {
		obj.IgnoreMissingComponents = new(false)
	}
}

// SetDefaults_BaseRepositoryConfig sets defaults for BaseRepositoryConfig.
func SetDefaults_BaseRepositoryConfig(obj *BaseRepositoryConfig) {
	if obj.Target == "" {
		obj.Target = "./"
	}
}

// SetDefaults_LandscapeRepositoryConfig sets defaults for LandscapeRepositoryConfig.
func SetDefaults_LandscapeRepositoryConfig(obj *LandscapeRepositoryConfig) {
	if obj.Target == "" {
		obj.Target = "./"
	}
}
