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

	if obj.MergeMode == nil {
		obj.MergeMode = new(MergeModeHint)
	}
}
