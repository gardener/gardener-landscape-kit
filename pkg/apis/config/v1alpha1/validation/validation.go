// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"net/url"
	"path"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
)

// ValidateLandscapeKitConfiguration validates the given LandscapeKitConfiguration.
func ValidateLandscapeKitConfiguration(conf *configv1alpha1.LandscapeKitConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}

	if conf.OCM != nil {
		allErrs = append(allErrs, ValidateOCMConfig(conf.OCM, field.NewPath("ocm"))...)
	}

	if conf.Repositories != nil {
		allErrs = append(allErrs, validateRepositories(conf.Repositories, field.NewPath("repositories"))...)
	}

	if conf.Components != nil {
		allErrs = append(allErrs, validateComponentsConfiguration(conf.Components, field.NewPath("components"))...)
	}

	if conf.VersionConfig != nil {
		allErrs = append(allErrs, ValidateVersionConfig(conf.VersionConfig, field.NewPath("versionConfig"))...)
	}

	if conf.MergeMode != nil && !slices.Contains(configv1alpha1.AllowedMergeModes, string(*conf.MergeMode)) {
		allErrs = append(allErrs, field.NotSupported(field.NewPath("mergeMode"), *conf.MergeMode, configv1alpha1.AllowedMergeModes))
	}

	return allErrs
}

func validateComponentsConfiguration(compConf *configv1alpha1.ComponentsConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(compConf.Exclude) > 0 && len(compConf.Include) > 0 {
		allErrs = append(allErrs, field.Forbidden(fldPath, "only one of 'exclude' or 'include' can be specified"))
	}

	foundComponents := sets.New[string]()
	for i, comp := range compConf.Exclude {
		if foundComponents.Has(comp) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("exclude").Index(i), comp))
		}
		foundComponents.Insert(comp)
	}

	foundComponents = sets.New[string]()
	for i, comp := range compConf.Include {
		if foundComponents.Has(comp) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("include").Index(i), comp))
		}
		foundComponents.Insert(comp)
	}

	return allErrs
}

func validateRepositories(repos *configv1alpha1.RepositoriesConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if repos.Base != nil {
		basePath := fldPath.Child("base")
		if path.IsAbs(repos.Base.Target) {
			allErrs = append(allErrs, field.Invalid(basePath.Child("target"), repos.Base.Target, "target must be a relative path within the base repository"))
		}
	}

	if repos.Landscape != nil {
		lsPath := fldPath.Child("landscape")

		if strings.TrimSpace(repos.Landscape.URL) == "" {
			allErrs = append(allErrs, field.Required(lsPath.Child("url"), "url must be specified"))
		} else {
			u, err := url.Parse(repos.Landscape.URL)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(lsPath.Child("url"), repos.Landscape.URL, "must be a valid URL"))
			} else if u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "ssh" {
				allErrs = append(allErrs, field.Invalid(lsPath.Child("url"), repos.Landscape.URL, "must have http(s) or ssh scheme"))
			}
		}

		allErrs = append(allErrs, validateGitRepositoryRef(&repos.Landscape.Ref, lsPath.Child("ref"))...)

		if strings.TrimSpace(repos.Landscape.BaseLink) == "" {
			allErrs = append(allErrs, field.Required(lsPath.Child("baseLink"), "baseLink must be specified"))
		} else if path.IsAbs(repos.Landscape.BaseLink) {
			allErrs = append(allErrs, field.Invalid(lsPath.Child("baseLink"), repos.Landscape.BaseLink, "baseLink must be a relative path within the landscape repository"))
		}

		if strings.TrimSpace(repos.Landscape.Target) != "" && path.IsAbs(repos.Landscape.Target) {
			allErrs = append(allErrs, field.Invalid(lsPath.Child("target"), repos.Landscape.Target, "target must be a relative path within the landscape repository"))
		}
	}

	return allErrs
}

func validateGitRepositoryRef(ref *configv1alpha1.GitRepositoryRef, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if ref.Branch != nil && strings.TrimSpace(*ref.Branch) == "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("branch"), *ref.Branch, "branch must not be empty"))
	}

	if ref.Tag != nil && strings.TrimSpace(*ref.Tag) == "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("tag"), *ref.Tag, "tag must not be empty"))
	}

	if ref.Commit != nil && strings.TrimSpace(*ref.Commit) == "" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("commit"), *ref.Commit, "commit SHA must not be empty"))
	}

	return allErrs
}

// ValidateOCMConfig validates the given OCMConfig.
func ValidateOCMConfig(conf *configv1alpha1.OCMConfig, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateOCMComponent(conf.RootComponent, fldPath.Child("rootComponent"))...)

	if len(conf.Repositories) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("repositories"), "at least one OCI repository must be specified in config file"))
	}

	for i, repo := range conf.Repositories {
		repoURL, err := url.Parse(repo)
		if err != nil || (repoURL != nil && len(repoURL.Host) == 0) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("repositories").Index(i), repo, "must be a valid URL"))
		}
	}

	return allErrs
}

func validateOCMComponent(conf configv1alpha1.OCMComponent, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if strings.TrimSpace(conf.Name) == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "component name is required in config file"))
	} else if len(strings.Split(conf.Name, "/")) == 1 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), conf.Name, "component name must be qualified (format 'example.com/my-org/my-root-component:1.23.4')"))
	}
	if strings.TrimSpace(conf.Version) == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("version"), "component version is required in config file"))
	}

	return allErrs
}

// ValidateVersionConfig validates the given VersionConfiguration.
func ValidateVersionConfig(conf *configv1alpha1.VersionConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if conf.DefaultVersionsUpdateStrategy != nil && !slices.Contains(configv1alpha1.AllowedDefaultVersionsUpdateStrategies, string(*conf.DefaultVersionsUpdateStrategy)) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("defaultVersionsUpdateStrategy"), *conf.DefaultVersionsUpdateStrategy, "allowed values are: "+strings.Join(configv1alpha1.AllowedDefaultVersionsUpdateStrategies, ", ")))
	}

	if conf.CheckMode != nil && !slices.Contains(configv1alpha1.AllowedVersionCheckModes, string(*conf.CheckMode)) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("checkMode"), *conf.CheckMode, "allowed values are: "+strings.Join(configv1alpha1.AllowedVersionCheckModes, ", ")))
	}

	return allErrs
}
