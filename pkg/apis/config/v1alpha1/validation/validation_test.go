// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1/validation"
)

var _ = Describe("Validation", func() {
	Describe("#ValidateLandscapeKitConfiguration", func() {
		It("should pass if no OCM or Git config is provided", func() {
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
				Git: &v1alpha1.GitRepository{
					URL: "https://github.com/gardener/gardener-landscape-kit",
					Ref: v1alpha1.GitRepositoryRef{
						Branch: ptr.To("main"),
					},
					Paths: v1alpha1.PathConfiguration{
						Base:      "base",
						Landscape: "landscape",
					},
				},
				VersionConfig: &v1alpha1.VersionConfiguration{
					ComponentsVectorFile: ptr.To("components.yaml"),
				},
			}

			errList := validation.ValidateLandscapeKitConfiguration(conf)
			Expect(errList).To(BeEmpty())
		})

		Context("Git Configuration", func() {
			It("should fail if Git config is invalid", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Git: &v1alpha1.GitRepository{
						URL: "invalid-url",
						Ref: v1alpha1.GitRepositoryRef{},
						Paths: v1alpha1.PathConfiguration{
							Base:      "",
							Landscape: "",
						},
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("git.url"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("git.paths.base"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("git.paths.landscape"),
					})),
				))
			})

			Context("URL", func() {
				test := func(url string) field.ErrorList {
					conf := &v1alpha1.LandscapeKitConfiguration{
						Git: &v1alpha1.GitRepository{
							URL: url,
							Paths: v1alpha1.PathConfiguration{
								Base:      "base",
								Landscape: "landscape",
							},
						},
					}

					return validation.ValidateLandscapeKitConfiguration(conf)
				}

				It("should pass with valid URL", func() {
					for _, urlScheme := range []string{
						"http://github.com/gardener/gardener-landscape-kit",
						"https://github.com/gardener/gardener-landscape-kit",
						"ssh://github.com/gardener/gardener-landscape-kit",
					} {
						Expect(test(urlScheme)).To(BeEmpty(), fmt.Sprintf("URL scheme %s should be valid", urlScheme))
					}
				})

				It("should fail with invalid URL scheme", func() {
					Expect(test("ftp://github.com/gardener/gardener-landscape-kit")).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("git.url"),
						}))))
				})
			})

			Context("Reference", func() {
				test := func(ref v1alpha1.GitRepositoryRef) field.ErrorList {
					conf := &v1alpha1.LandscapeKitConfiguration{
						Git: &v1alpha1.GitRepository{
							URL: "https://github.com/gardener/gardener-landscape-kit",
							Ref: ref,
							Paths: v1alpha1.PathConfiguration{
								Base:      "base",
								Landscape: "landscape",
							},
						},
					}

					return validation.ValidateLandscapeKitConfiguration(conf)
				}

				It("should pass with valid refs", func() {
					for _, ref := range []v1alpha1.GitRepositoryRef{
						{Branch: ptr.To("main")},
						{Tag: ptr.To("v1.0.0")},
						{Commit: ptr.To("abc123def456")},
					} {
						Expect(test(ref)).To(BeEmpty(), fmt.Sprintf("Git ref %+v should be valid", ref))
					}
				})

				It("should fail with empty refs", func() {
					for _, ref := range []v1alpha1.GitRepositoryRef{
						{Branch: ptr.To("")},
						{Tag: ptr.To("")},
						{Commit: ptr.To("")},
					} {
						Expect(test(ref)).To(ConsistOf(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Type": Equal(field.ErrorTypeInvalid),
							}))))
					}
				})
			})

			Context("Paths", func() {
				test := func(basePath, landscapePath string) field.ErrorList {
					conf := &v1alpha1.LandscapeKitConfiguration{
						Git: &v1alpha1.GitRepository{
							URL: "https://github.com/gardener/gardener-landscape-kit",
							Ref: v1alpha1.GitRepositoryRef{},
							Paths: v1alpha1.PathConfiguration{
								Base:      basePath,
								Landscape: landscapePath,
							},
						},
					}

					return validation.ValidateLandscapeKitConfiguration(conf)
				}

				It("should pass with valid relative paths", func() {
					Expect(test("base", "landscape")).To(BeEmpty())
					Expect(test("base/path", "landscape/path")).To(BeEmpty())
					Expect(test("./base", "./landscape")).To(BeEmpty())
					Expect(test("./", "./")).To(BeEmpty())
					Expect(test(".", ".")).To(BeEmpty())
				})

				It("should fail with absolute paths", func() {
					Expect(test("/base", "/landscape")).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("git.paths.base"),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("git.paths.landscape"),
						})),
					))
				})
			})
		})

		Context("Components Configuration", func() {
			It("should pass with empty include and exclude lists", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should pass with exclude list", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Components: &v1alpha1.ComponentsConfiguration{
						Exclude: []string{"excluded-component-1", "excluded-component-2"},
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should fail with duplicate elements in exclude list", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Components: &v1alpha1.ComponentsConfiguration{
						Exclude: []string{"excluded-component-1", "excluded-component-2", "excluded-component-1"},
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeDuplicate),
						"Field": Equal("components.exclude[2]"),
					})),
				))
			})

			It("should pass with include list", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Components: &v1alpha1.ComponentsConfiguration{
						Include: []string{"include-component-1", "include-component-2"},
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should fail with duplicate elements in include list", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Components: &v1alpha1.ComponentsConfiguration{
						Include: []string{"include-component-1", "include-component-2", "include-component-1"},
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeDuplicate),
						"Field": Equal("components.include[2]"),
					})),
				))
			})

			It("should fail if both include and exclude lists are provided", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Components: &v1alpha1.ComponentsConfiguration{
						Exclude: []string{"exclude-component-1", "exclude-component-2"},
						Include: []string{"include-component-1", "include-component-2"},
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeForbidden),
						"Field": Equal("components"),
					})),
				))
			})
		})

		Context("OCM Configuration", func() {
			setupOCMConfigTests(func(ocmConf *v1alpha1.OCMConfig) field.ErrorList {
				conf := &v1alpha1.LandscapeKitConfiguration{
					OCM: ocmConf,
				}
				return validation.ValidateLandscapeKitConfiguration(conf)
			}, field.NewPath("ocm"))
		})

		Context("VersionConfig Configuration", func() {
			It("should fail if ComponentsVectorFile is empty", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						ComponentsVectorFile: ptr.To(""),
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("versionConfig.componentsVectorFile"),
					})),
				))
			})

			It("should fail if ComponentsVectorFile is whitespace only", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						ComponentsVectorFile: ptr.To("   "),
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("versionConfig.componentsVectorFile"),
					})),
				))
			})

			It("should pass with a valid ComponentsVectorFile", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						ComponentsVectorFile: ptr.To("path/to/components.yaml"),
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should pass with a valid DefaultVersionsUpdateStrategy", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						DefaultVersionsUpdateStrategy: ptr.To("ReleaseBranch"),
					},
				}

				errList := validation.ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})
		})
	})
})

func setupOCMConfigTests(test func(conf *v1alpha1.OCMConfig) field.ErrorList, baseFldPath *field.Path) {
	It("should fail if OCM config is invalid", func() {
		conf := &v1alpha1.OCMConfig{
			Repositories: []string{}, // empty repositories
			RootComponent: v1alpha1.OCMComponent{
				Name:    "", // missing name
				Version: "", // missing version
			},
		}

		errList := test(conf)
		Expect(errList).To(HaveLen(3))
	})

	It("should pass with a valid configuration", func() {
		conf := &v1alpha1.OCMConfig{
			Repositories: []string{"https://example.com/repo"},
			RootComponent: v1alpha1.OCMComponent{
				Name:    "example.com/org/component",
				Version: "1.0.0",
			},
		}

		errList := test(conf)
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

		errList := test(conf)
		Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":  Equal(field.ErrorTypeRequired),
			"Field": Equal(baseFldPath.Child("rootComponent.name").String()),
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

		errList := test(conf)
		Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal(baseFldPath.Child("rootComponent.name").String()),
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

		errList := test(conf)
		Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":  Equal(field.ErrorTypeRequired),
			"Field": Equal(baseFldPath.Child("rootComponent.version").String()),
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

		errList := test(conf)
		Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":  Equal(field.ErrorTypeRequired),
			"Field": Equal(baseFldPath.Child("repositories").String()),
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

		errList := test(conf)
		Expect(errList).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
			"Type":     Equal(field.ErrorTypeInvalid),
			"Field":    Equal(baseFldPath.Child("repositories[0]").String()),
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

		errList := test(conf)
		Expect(errList).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":     Equal(field.ErrorTypeInvalid),
				"Field":    Equal(baseFldPath.Child("repositories[0]").String()),
				"BadValue": Equal("invalid-url"),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Type":     Equal(field.ErrorTypeInvalid),
				"Field":    Equal(baseFldPath.Child("repositories[1]").String()),
				"BadValue": Equal("another-invalid"),
			})),
		))
	})
}
