// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"

	. "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

var _ = Describe("ComponentVector Validation", func() {
	Describe("#ValidateComponents", func() {
		var fldPath *field.Path

		BeforeEach(func() {
			fldPath = field.NewPath("test")
		})

		It("should pass with valid components", func() {
			components := &Components{
				Components: []*ComponentVector{
					{
						Name:             "component1",
						SourceRepository: "https://github.com/org/repo1",
						Version:          "1.0.0",
					},
					{
						Name:             "component2",
						SourceRepository: "https://github.com/org/repo2",
						Version:          "2.0.0",
					},
				},
			}

			errList := ValidateComponents(components, fldPath)
			Expect(errList).To(BeEmpty())
		})

		It("should pass with nil components", func() {
			errList := ValidateComponents(nil, fldPath)
			Expect(errList).To(BeEmpty())
		})

		It("should fail if components list is empty", func() {
			components := &Components{
				Components: []*ComponentVector{},
			}

			errList := ValidateComponents(components, fldPath)
			Expect(errList).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("test.components"),
				})),
			))
		})

		Context("Component Name Validation", func() {
			It("should fail if component name is empty", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "",
							SourceRepository: "https://github.com/org/repo",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].name"),
					})),
				))
			})

			It("should fail if component name is whitespace only", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "   ",
							SourceRepository: "https://github.com/org/repo",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].name"),
					})),
				))
			})

			It("should fail with duplicate component names", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "duplicate-component",
							SourceRepository: "https://github.com/org/repo1",
							Version:          "1.0.0",
						},
						{
							Name:             "duplicate-component",
							SourceRepository: "https://github.com/org/repo2",
							Version:          "2.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeDuplicate),
						"Field":    Equal("test.components[1].name"),
						"BadValue": Equal("duplicate-component"),
					})),
				))
			})

			It("should fail with multiple duplicate component names", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component-a",
							SourceRepository: "https://github.com/org/repo1",
							Version:          "1.0.0",
						},
						{
							Name:             "component-a",
							SourceRepository: "https://github.com/org/repo2",
							Version:          "2.0.0",
						},
						{
							Name:             "component-b",
							SourceRepository: "https://github.com/org/repo3",
							Version:          "3.0.0",
						},
						{
							Name:             "component-a",
							SourceRepository: "https://github.com/org/repo4",
							Version:          "4.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeDuplicate),
						"Field":    Equal("test.components[1].name"),
						"BadValue": Equal("component-a"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeDuplicate),
						"Field":    Equal("test.components[3].name"),
						"BadValue": Equal("component-a"),
					})),
				))
			})
		})

		Context("Source Repository Validation", func() {
			It("should fail if source repository is empty", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].sourceRepository"),
					})),
				))
			})

			It("should fail if source repository is whitespace only", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "   ",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].sourceRepository"),
					})),
				))
			})

			It("should fail if source repository URL is malformed", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "://invalid-url",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeInvalid),
						"Field":    Equal("test.components[0].sourceRepository"),
						"BadValue": Equal("://invalid-url"),
					})),
				))
			})

			It("should fail if source repository URL has no scheme", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "github.com/org/repo",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeInvalid),
						"Field":    Equal("test.components[0].sourceRepository"),
						"BadValue": Equal("github.com/org/repo"),
					})),
				))
			})

			It("should fail if source repository URL has no host", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "https://",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeInvalid),
						"Field":    Equal("test.components[0].sourceRepository"),
						"BadValue": Equal("https://"),
					})),
				))
			})

			It("should pass with valid HTTP URL", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "http://github.com/org/repo",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(BeEmpty())
			})

			It("should pass with valid HTTPS URL", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "https://github.com/org/repo",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(BeEmpty())
			})

			It("should pass with SSH URL", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "ssh://git@github.com/org/repo",
							Version:          "1.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(BeEmpty())
			})
		})

		Context("Version Validation", func() {
			It("should fail if version is empty", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "https://github.com/org/repo",
							Version:          "",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].version"),
					})),
				))
			})

			It("should fail if version is whitespace only", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "https://github.com/org/repo",
							Version:          "   ",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].version"),
					})),
				))
			})
		})

		Context("Nil Component Entries", func() {
			It("should fail if a component entry is nil", func() {
				components := &Components{
					Components: []*ComponentVector{
						nil,
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0]"),
					})),
				))
			})

			It("should handle mix of nil and valid components", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "component1",
							SourceRepository: "https://github.com/org/repo1",
							Version:          "1.0.0",
						},
						nil,
						{
							Name:             "component2",
							SourceRepository: "https://github.com/org/repo2",
							Version:          "2.0.0",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[1]"),
					})),
				))
			})
		})

		Context("Multiple Validation Errors", func() {
			It("should report all errors for a single component", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "",
							SourceRepository: "",
							Version:          "",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].name"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].sourceRepository"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].version"),
					})),
				))
			})

			It("should report errors across multiple components", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "",
							SourceRepository: "https://github.com/org/repo1",
							Version:          "1.0.0",
						},
						{
							Name:             "component2",
							SourceRepository: "invalid-url",
							Version:          "2.0.0",
						},
						{
							Name:             "component3",
							SourceRepository: "https://github.com/org/repo3",
							Version:          "",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[0].name"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeInvalid),
						"Field":    Equal("test.components[1].sourceRepository"),
						"BadValue": Equal("invalid-url"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[2].version"),
					})),
				))
			})

			It("should report both duplicate names and other validation errors", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name:             "duplicate",
							SourceRepository: "https://github.com/org/repo1",
							Version:          "1.0.0",
						},
						{
							Name:             "duplicate",
							SourceRepository: "",
							Version:          "2.0.0",
						},
						{
							Name:             "component3",
							SourceRepository: "https://github.com/org/repo3",
							Version:          "",
						},
					},
				}

				errList := ValidateComponents(components, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeDuplicate),
						"Field":    Equal("test.components[1].name"),
						"BadValue": Equal("duplicate"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[1].sourceRepository"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("test.components[2].version"),
					})),
				))
			})
		})
	})
})
