// SPDX-FileCopyrightText: Contributors to the Gardener project
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
		var (
			fldPath                            *field.Path
			advancedFields                     bool
			componentVector1, componentVector2 *ComponentVector

			commonTests = func(advancedFields func() bool) {
				It("should pass with valid components", func() {
					components := &Components{
						Components: []*ComponentVector{
							componentVector1,
							componentVector2,
						},
					}

					errList := ValidateComponents(components, advancedFields(), fldPath)
					Expect(errList).To(BeEmpty())
				})

				It("should pass with nil components", func() {
					errList := ValidateComponents(nil, advancedFields(), fldPath)
					Expect(errList).To(BeEmpty())
				})

				It("should fail if components list is empty", func() {
					components := &Components{
						Components: []*ComponentVector{},
					}

					errList := ValidateComponents(components, advancedFields(), fldPath)
					Expect(errList).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeRequired),
							"Field": Equal("test.components"),
						})),
					))
				})

				Context("Component Name Validation", func() {
					It("should fail if component name is empty", func() {
						componentVector1.Name = ""

						components := &Components{
							Components: []*ComponentVector{
								componentVector1,
							},
						}

						errList := ValidateComponents(components, advancedFields(), fldPath)
						Expect(errList).To(ConsistOf(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Type":  Equal(field.ErrorTypeRequired),
								"Field": Equal("test.components[0].name"),
							})),
						))
					})

					It("should fail if component name is whitespace only", func() {
						componentVector1.Name = "  "
						components := &Components{
							Components: []*ComponentVector{
								componentVector1,
							},
						}

						errList := ValidateComponents(components, advancedFields(), fldPath)
						Expect(errList).To(ConsistOf(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Type":  Equal(field.ErrorTypeRequired),
								"Field": Equal("test.components[0].name"),
							})),
						))
					})

					It("should fail with duplicate component names", func() {
						componentVector1.Name = "duplicate-component"
						componentVector2.Name = "duplicate-component"

						components := &Components{
							Components: []*ComponentVector{
								componentVector1,
								componentVector2,
							},
						}

						errList := ValidateComponents(components, advancedFields(), fldPath)
						Expect(errList).To(ConsistOf(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Type":     Equal(field.ErrorTypeDuplicate),
								"Field":    Equal("test.components[1].name"),
								"BadValue": Equal("duplicate-component"),
							})),
						))
					})

					It("should fail with multiple duplicate component names", func() {
						componentVector3 := *componentVector1
						componentVector4 := *componentVector1

						components := &Components{
							Components: []*ComponentVector{
								componentVector1,
								componentVector2,
								&componentVector3,
								&componentVector4,
							},
						}

						errList := ValidateComponents(components, advancedFields(), fldPath)
						Expect(errList).To(ConsistOf(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Type":     Equal(field.ErrorTypeDuplicate),
								"Field":    Equal("test.components[2].name"),
								"BadValue": Equal(componentVector3.Name),
							})),
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Type":     Equal(field.ErrorTypeDuplicate),
								"Field":    Equal("test.components[3].name"),
								"BadValue": Equal(componentVector4.Name),
							})),
						))
					})
				})
			}
		)

		BeforeEach(func() {
			fldPath = field.NewPath("test")
		})

		Context("Basic Validation", func() {
			BeforeEach(func() {
				advancedFields = false

				componentVector1 = &ComponentVector{
					Name: "component1",
				}

				componentVector2 = &ComponentVector{
					Name: "component2",
				}
			})

			commonTests(func() bool { return advancedFields })

			It("should pass with only the allowed fields set", func() {
				components := &Components{
					Components: []*ComponentVector{
						{
							Name: "component1",
							Resources: map[string]ResourceData{
								"chart": {OCIImage: &OCIImage{Repository: new("registry.example.com/image")}},
							},
							ImageVectorOverwrite: &ImageVectorOverwrite{},
						},
					},
				}

				errList := ValidateComponents(components, false, fldPath)
				Expect(errList).To(BeEmpty())
			})

			It("should fail when version is set", func() {
				components := &Components{
					Components: []*ComponentVector{
						{Name: "component1", Version: "1.0.0"},
					},
				}

				errList := ValidateComponents(components, false, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeForbidden),
						"Field": Equal("test.components[0].version"),
					})),
				))
			})

			It("should fail when sourceRepository is set", func() {
				components := &Components{
					Components: []*ComponentVector{
						{Name: "component1", SourceRepository: new("https://github.com/org/repo")},
					},
				}

				errList := ValidateComponents(components, false, fldPath)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeForbidden),
						"Field": Equal("test.components[0].sourceRepository"),
					})),
				))
			})
		})

		Context("Advanced Validation", func() {
			BeforeEach(func() {
				advancedFields = true

				componentVector1 = &ComponentVector{
					Name:             "component1",
					SourceRepository: new("https://github.com/org/repo1"),
					Version:          "1.0.0",
				}

				componentVector2 = &ComponentVector{
					Name:             "component2",
					SourceRepository: new("https://github.com/org/repo2"),
					Version:          "2.0.0",
				}
			})

			commonTests(func() bool { return advancedFields })

			Context("Source Repository Validation", func() {
				It("should succeed if source repository is nil", func() {
					components := &Components{
						Components: []*ComponentVector{
							{
								Name:    "component1",
								Version: "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
					Expect(errList).To(BeEmpty())
				})

				It("should fail if source repository is empty", func() {
					components := &Components{
						Components: []*ComponentVector{
							{
								Name:             "component1",
								SourceRepository: new(""),
								Version:          "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("   "),
								Version:          "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("://invalid-url"),
								Version:          "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("github.com/org/repo"),
								Version:          "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("https://"),
								Version:          "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("http://github.com/org/repo"),
								Version:          "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
					Expect(errList).To(BeEmpty())
				})

				It("should pass with valid HTTPS URL", func() {
					components := &Components{
						Components: []*ComponentVector{
							{
								Name:             "component1",
								SourceRepository: new("https://github.com/org/repo"),
								Version:          "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
					Expect(errList).To(BeEmpty())
				})

				It("should pass with SSH URL", func() {
					components := &Components{
						Components: []*ComponentVector{
							{
								Name:             "component1",
								SourceRepository: new("ssh://git@github.com/org/repo"),
								Version:          "1.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
					Expect(errList).To(BeEmpty())
				})
			})

			Context("Version Validation", func() {
				It("should fail if version is empty", func() {
					components := &Components{
						Components: []*ComponentVector{
							{
								Name:             "component1",
								SourceRepository: new("https://github.com/org/repo"),
								Version:          "",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("https://github.com/org/repo"),
								Version:          "   ",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("https://github.com/org/repo1"),
								Version:          "1.0.0",
							},
							nil,
							{
								Name:             "component2",
								SourceRepository: new("https://github.com/org/repo2"),
								Version:          "2.0.0",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new(""),
								Version:          "",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("https://github.com/org/repo1"),
								Version:          "1.0.0",
							},
							{
								Name:             "component2",
								SourceRepository: new("invalid-url"),
								Version:          "2.0.0",
							},
							{
								Name:             "component3",
								SourceRepository: new("https://github.com/org/repo3"),
								Version:          "",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
								SourceRepository: new("https://github.com/org/repo1"),
								Version:          "1.0.0",
							},
							{
								Name:             "duplicate",
								SourceRepository: new(""),
								Version:          "2.0.0",
							},
							{
								Name:             "component3",
								SourceRepository: new("https://github.com/org/repo3"),
								Version:          "",
							},
						},
					}

					errList := ValidateComponents(components, advancedFields, fldPath)
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
})
