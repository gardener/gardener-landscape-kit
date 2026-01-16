// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"net/url"
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
)

// ValidateLandscapeKitConfiguration validates the given LandscapeKitConfiguration.
func ValidateLandscapeKitConfiguration(conf *configv1alpha1.LandscapeKitConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}

	if conf.OCM != nil {
		allErrs = append(allErrs, ValidateOCMConfig(conf.OCM, field.NewPath("ocm"))...)
	}

	if conf.Git != nil {
		allErrs = append(allErrs, validateGitRepository(conf.Git, field.NewPath("git"))...)
	}

	if conf.VersionConfig != nil {
		allErrs = append(allErrs, ValidateVersionConfig(conf.VersionConfig, field.NewPath("versionConfig"))...)
	}

	return allErrs
}

func validateGitRepository(repo *configv1alpha1.GitRepository, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	u, err := url.Parse(repo.URL)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("url"), repo.URL, "must be a valid URL"))
	}
	if u.Scheme != "https" && u.Scheme != "http" && u.Scheme != "ssh" {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("url"), repo.URL, "must have http(s) or ssh scheme"))
	}

	allErrs = append(allErrs, validateGitRepositoryRef(&repo.Ref, fldPath.Child("ref"))...)
	allErrs = append(allErrs, validatePathConfiguration(&repo.Paths, fldPath.Child("paths"))...)

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

func validatePathConfiguration(paths *configv1alpha1.PathConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if strings.TrimSpace(paths.Base) == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("base"), "base path must be specified"))
	}

	if path.IsAbs(paths.Base) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("base"), paths.Base, "base path must be a relative path to the base dir within the git repository"))
	}

	if strings.TrimSpace(paths.Landscape) == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("landscape"), "landscape path must be specified"))
	}

	if path.IsAbs(paths.Landscape) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("landscape"), paths.Landscape, "landscape path must be a relative path to the landscape dir within the git repository"))
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

	if conf.ComponentsVectorFile != nil && strings.TrimSpace(*conf.ComponentsVectorFile) == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("componentsVectorFile"), "components vector file path must be specified"))
	}

	return allErrs
}
