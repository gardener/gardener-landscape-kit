// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1/validation"
)

var _ = Describe("Validation", func() {
	Describe("#ValidateLandscapeKitConfiguration", func() {
		It("should pass if no OCM config is provided", func() {
			conf := &v1alpha1.LandscapeKitConfiguration{}

			errList := validation.ValidateLandscapeKitConfiguration(conf)
			Expect(errList).To(BeEmpty())
		})

		It("should pass with a valid configuration", func() {
			conf := &v1alpha1.LandscapeKitConfiguration{
				OCM: &v1alpha1.OCMConfig{
					Repositories: []string{"https://example.com/repo"},
					RootComponent: v1alpha1.OCMComponent{
						Name:    "example.com/org/component",
						Version: "1.0.0",
					},
				},
			}

			errList := validation.ValidateLandscapeKitConfiguration(conf)
			Expect(errList).To(BeEmpty())
		})

		It("should fail if OCM config is invalid", func() {
			conf := &v1alpha1.LandscapeKitConfiguration{
				OCM: &v1alpha1.OCMConfig{
					Repositories: []string{}, // empty repositories
					RootComponent: v1alpha1.OCMComponent{
						Name:    "", // missing name
						Version: "", // missing version
					},
				},
			}

			errList := validation.ValidateLandscapeKitConfiguration(conf)
			Expect(errList).To(HaveLen(3))
		})
	})

	Describe("#ValidateOCMConfiguration", func() {
		It("should pass with a valid configuration", func() {
			conf := &v1alpha1.OCMConfiguration{
				OCMConfig: &v1alpha1.OCMConfig{
					Repositories: []string{"https://example.com/repo"},
					RootComponent: v1alpha1.OCMComponent{
						Name:    "example.com/org/component",
						Version: "1.0.0",
					},
				},
			}

			errList := validation.ValidateOCMConfiguration(conf)
			Expect(errList).To(BeEmpty())
		})

		It("should fail if config is invalid", func() {
			conf := &v1alpha1.OCMConfiguration{
				OCMConfig: &v1alpha1.OCMConfig{
					Repositories: []string{}, // empty repositories
					RootComponent: v1alpha1.OCMComponent{
						Name:    "", // missing name
						Version: "", // missing version
					},
				},
			}

			errList := validation.ValidateOCMConfiguration(conf)
			Expect(errList).To(HaveLen(3))
		})
	})

	Describe("#ValidateOCMConfig", func() {
		It("should pass with a valid configuration", func() {
			conf := &v1alpha1.OCMConfig{
				Repositories: []string{"https://example.com/repo"},
				RootComponent: v1alpha1.OCMComponent{
					Name:    "example.com/org/component",
					Version: "1.0.0",
				},
			}

			errList := validation.ValidateOCMConfig(conf, field.NewPath("ocm"))
			Expect(errList).To(BeEmpty())
		})

		It("should fail if root component name is missing", func() {
			conf := &v1alpha1.OCMConfig{
				Repositories: []string{"https://example.com/repo"},
				RootComponent: v1alpha1.OCMComponent{
					Name:    "", // missing name
					Version: "1.0.0",
				},
			}

			errList := validation.ValidateOCMConfig(conf, field.NewPath("ocm"))
			Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("ocm.rootComponent.name"),
			}))))
		})

		It("should fail if root component name is unqualified", func() {
			conf := &v1alpha1.OCMConfig{
				Repositories: []string{"https://example.com/repo"},
				RootComponent: v1alpha1.OCMComponent{
					Name:    "component", // unqualified name
					Version: "1.0.0",
				},
			}

			errList := validation.ValidateOCMConfig(conf, field.NewPath("ocm"))
			Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":     Equal(field.ErrorTypeInvalid),
				"Field":    Equal("ocm.rootComponent.name"),
				"BadValue": Equal("component"),
			}))))
		})

		It("should fail if root component version is missing", func() {
			conf := &v1alpha1.OCMConfig{
				Repositories: []string{"https://example.com/repo"},
				RootComponent: v1alpha1.OCMComponent{
					Name:    "example.com/org/component",
					Version: "", // missing version
				},
			}

			errList := validation.ValidateOCMConfig(conf, field.NewPath("ocm"))
			Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("ocm.rootComponent.version"),
			}))))
		})

		It("should fail if no repositories are provided", func() {
			conf := &v1alpha1.OCMConfig{
				Repositories: []string{}, // empty repositories
				RootComponent: v1alpha1.OCMComponent{
					Name:    "example.com/org/component",
					Version: "1.0.0",
				},
			}

			errList := validation.ValidateOCMConfig(conf, field.NewPath("ocm"))
			Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":  Equal(field.ErrorTypeRequired),
				"Field": Equal("ocm.repositories"),
			}))))
		})

		It("should fail if a repository URL is invalid", func() {
			conf := &v1alpha1.OCMConfig{
				Repositories: []string{"invalid-url"},
				RootComponent: v1alpha1.OCMComponent{
					Name:    "example.com/org/component",
					Version: "1.0.0",
				},
			}

			errList := validation.ValidateOCMConfig(conf, field.NewPath("ocm"))
			Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":     Equal(field.ErrorTypeInvalid),
				"Field":    Equal("ocm.repositories[0]"),
				"BadValue": Equal("invalid-url"),
			}))))
		})

		It("should fail with multiple invalid repositories", func() {
			conf := &v1alpha1.OCMConfig{
				Repositories: []string{"invalid-url", "another-invalid"},
				RootComponent: v1alpha1.OCMComponent{
					Name:    "example.com/org/component",
					Version: "1.0.0",
				},
			}

			errList := validation.ValidateOCMConfig(conf, field.NewPath("ocm"))
			Expect(errList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":     Equal(field.ErrorTypeInvalid),
					"Field":    Equal("ocm.repositories[0]"),
					"BadValue": Equal("invalid-url"),
				})),
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":     Equal(field.ErrorTypeInvalid),
					"Field":    Equal("ocm.repositories[1]"),
					"BadValue": Equal("another-invalid"),
				})),
			))
		})
	})
})
