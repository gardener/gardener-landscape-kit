// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateComponents validates the given Components.
func ValidateComponents(components *Components, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if components == nil {
		return allErrs
	}

	if len(components.Components) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("components"), "at least one component must be specified"))
		return allErrs
	}

	// Track component names to detect duplicates
	componentNames := sets.New[string]()

	for i, component := range components.Components {
		componentPath := fldPath.Child("components").Index(i)

		// Validate individual component
		allErrs = append(allErrs, validateComponentVector(component, componentPath)...)

		// Check for duplicate names
		if component != nil && component.Name != "" {
			if componentNames.Has(component.Name) {
				allErrs = append(allErrs, field.Duplicate(componentPath.Child("name"), component.Name))
			} else {
				componentNames.Insert(component.Name)
			}
		}
	}

	return allErrs
}

// validateComponentVector validates a single ComponentVector.
func validateComponentVector(component *ComponentVector, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if component == nil {
		allErrs = append(allErrs, field.Required(fldPath, "component must not be empty"))
		return allErrs
	}

	// Validate Name
	if strings.TrimSpace(component.Name) == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("name"), "component name must not be empty"))
	}

	// Validate SourceRepository
	if strings.TrimSpace(component.SourceRepository) == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("sourceRepository"), "source repository must not be empty"))
	} else {
		// Validate URL format
		repoURL, err := url.Parse(component.SourceRepository)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("sourceRepository"), component.SourceRepository, "must be a valid URL"))
		} else if repoURL.Scheme == "" {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("sourceRepository"), component.SourceRepository, "must have a valid URL scheme (e.g., https, http)"))
		} else if repoURL.Host == "" {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("sourceRepository"), component.SourceRepository, "must have a valid host"))
		}
	}

	// Validate Version
	if strings.TrimSpace(component.Version) == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("version"), "component version must not be empty"))
	}

	return allErrs
}
