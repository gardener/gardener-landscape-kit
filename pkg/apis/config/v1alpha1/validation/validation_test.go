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

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	. "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1/validation"
)

var _ = Describe("Validation", func() {
	Describe("#ValidateLandscapeKitConfiguration", func() {
		It("should pass if no OCM or Repositories config is provided", func() {
			conf := &v1alpha1.LandscapeKitConfiguration{}

			errList := ValidateLandscapeKitConfiguration(conf)
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
				Repositories: &v1alpha1.RepositoriesConfig{
					Base: &v1alpha1.BaseRepositoryConfig{
						Target: "base",
					},
					Landscape: &v1alpha1.LandscapeRepositoryConfig{
						URL: "https://github.com/gardener/gardener-landscape-kit",
						Ref: v1alpha1.GitRepositoryRef{
							Branch: new("main"),
						},
						BaseLink: "base",
						Target:   "landscape",
					},
				},
			}

			errList := ValidateLandscapeKitConfiguration(conf)
			Expect(errList).To(BeEmpty())
		})

		Context("Repositories Configuration", func() {
			It("should pass with a valid configuration", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Repositories: &v1alpha1.RepositoriesConfig{
						Base: &v1alpha1.BaseRepositoryConfig{
							Target: "base",
						},
						Landscape: &v1alpha1.LandscapeRepositoryConfig{
							URL:      "https://github.com/gardener/gardener-landscape-kit",
							Ref:      v1alpha1.GitRepositoryRef{Branch: new("main")},
							BaseLink: "base",
							Target:   "landscape",
						},
					},
				}

				errList := ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should fail if Repositories config is invalid", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Repositories: &v1alpha1.RepositoriesConfig{
						Base: &v1alpha1.BaseRepositoryConfig{
							Target: "",
						},
						Landscape: &v1alpha1.LandscapeRepositoryConfig{
							URL:      "ftp://github.com/gardener/gardener-landscape-kit",
							Ref:      v1alpha1.GitRepositoryRef{},
							BaseLink: "",
							Target:   "",
						},
					},
				}

				errList := ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeInvalid),
						"Field": Equal("repositories.landscape.url"),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":  Equal(field.ErrorTypeRequired),
						"Field": Equal("repositories.landscape.baseLink"),
					})),
				))
			})

			Context("Landscape URL", func() {
				test := func(url string) field.ErrorList {
					conf := &v1alpha1.LandscapeKitConfiguration{
						Repositories: &v1alpha1.RepositoriesConfig{
							Base: &v1alpha1.BaseRepositoryConfig{Target: "base"},
							Landscape: &v1alpha1.LandscapeRepositoryConfig{
								URL:      url,
								BaseLink: "base",
								Target:   "landscape",
							},
						},
					}

					return ValidateLandscapeKitConfiguration(conf)
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
							"Field": Equal("repositories.landscape.url"),
						}))))
				})

				It("should fail with empty URL", func() {
					Expect(test("")).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeRequired),
							"Field": Equal("repositories.landscape.url"),
						}))))
				})
			})

			Context("Landscape Reference", func() {
				test := func(ref v1alpha1.GitRepositoryRef) field.ErrorList {
					conf := &v1alpha1.LandscapeKitConfiguration{
						Repositories: &v1alpha1.RepositoriesConfig{
							Base: &v1alpha1.BaseRepositoryConfig{Target: "base"},
							Landscape: &v1alpha1.LandscapeRepositoryConfig{
								URL:      "https://github.com/gardener/gardener-landscape-kit",
								Ref:      ref,
								BaseLink: "base",
								Target:   "landscape",
							},
						},
					}

					return ValidateLandscapeKitConfiguration(conf)
				}

				It("should pass with valid refs", func() {
					for _, ref := range []v1alpha1.GitRepositoryRef{
						{Branch: new("main")},
						{Tag: new("v1.0.0")},
						{Commit: new("abc123def456")},
					} {
						Expect(test(ref)).To(BeEmpty(), fmt.Sprintf("Git ref %+v should be valid", ref))
					}
				})

				It("should fail with empty refs", func() {
					for _, refAndField := range []struct {
						ref   v1alpha1.GitRepositoryRef
						field string
					}{
						{v1alpha1.GitRepositoryRef{Branch: new("")}, "repositories.landscape.ref.branch"},
						{v1alpha1.GitRepositoryRef{Tag: new("")}, "repositories.landscape.ref.tag"},
						{v1alpha1.GitRepositoryRef{Commit: new("")}, "repositories.landscape.ref.commit"},
					} {
						Expect(test(refAndField.ref)).To(ConsistOf(
							PointTo(MatchFields(IgnoreExtras, Fields{
								"Type":  Equal(field.ErrorTypeInvalid),
								"Field": Equal(refAndField.field),
							}))))
					}
				})
			})

			Context("Paths", func() {
				test := func(baseTarget, baseLink, landscapeTarget string) field.ErrorList {
					conf := &v1alpha1.LandscapeKitConfiguration{
						Repositories: &v1alpha1.RepositoriesConfig{
							Base: &v1alpha1.BaseRepositoryConfig{Target: baseTarget},
							Landscape: &v1alpha1.LandscapeRepositoryConfig{
								URL:      "https://github.com/gardener/gardener-landscape-kit",
								Ref:      v1alpha1.GitRepositoryRef{},
								BaseLink: baseLink,
								Target:   landscapeTarget,
							},
						},
					}

					return ValidateLandscapeKitConfiguration(conf)
				}

				It("should pass with valid relative paths", func() {
					Expect(test("base", "base", "landscape")).To(BeEmpty())
					Expect(test("base/path", "base/path", "landscape/path")).To(BeEmpty())
					Expect(test("./base", "./base", "./landscape")).To(BeEmpty())
					Expect(test("./", "./", "./")).To(BeEmpty())
					Expect(test(".", ".", ".")).To(BeEmpty())
				})

				It("should fail with empty baseLink while target paths are optional", func() {
					Expect(test("", "", "")).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeRequired),
							"Field": Equal("repositories.landscape.baseLink"),
						})),
					))
				})

				It("should fail with absolute paths", func() {
					Expect(test("/base", "/base", "/landscape")).To(ConsistOf(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("repositories.base.target"),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("repositories.landscape.baseLink"),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Type":  Equal(field.ErrorTypeInvalid),
							"Field": Equal("repositories.landscape.target"),
						})),
					))
				})
			})
		})

		Context("Components Configuration", func() {
			It("should pass with empty include and exclude lists", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{}

				errList := ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should pass with exclude list", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Components: &v1alpha1.ComponentsConfiguration{
						Exclude: []string{"excluded-component-1", "excluded-component-2"},
					},
				}

				errList := ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should fail with duplicate elements in exclude list", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Components: &v1alpha1.ComponentsConfiguration{
						Exclude: []string{"excluded-component-1", "excluded-component-2", "excluded-component-1"},
					},
				}

				errList := ValidateLandscapeKitConfiguration(conf)
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

				errList := ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should fail with duplicate elements in include list", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					Components: &v1alpha1.ComponentsConfiguration{
						Include: []string{"include-component-1", "include-component-2", "include-component-1"},
					},
				}

				errList := ValidateLandscapeKitConfiguration(conf)
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

				errList := ValidateLandscapeKitConfiguration(conf)
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
				return ValidateLandscapeKitConfiguration(conf)
			}, field.NewPath("ocm"))
		})

		Context("VersionConfig Configuration", func() {
			It("should pass with a valid DefaultVersionsUpdateStrategy", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						DefaultVersionsUpdateStrategy: new(v1alpha1.DefaultVersionsUpdateStrategyReleaseBranch),
					},
				}

				errList := ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})
		})

		Context("MergeMode Configuration", func() {
			It("should pass with valid MergeMode values", func() {
				for _, mode := range []v1alpha1.MergeMode{
					v1alpha1.MergeModeHint,
					v1alpha1.MergeModeSilent,
				} {
					conf := &v1alpha1.LandscapeKitConfiguration{
						MergeMode: &mode,
					}
					errList := ValidateLandscapeKitConfiguration(conf)
					Expect(errList).To(BeEmpty(), fmt.Sprintf("MergeMode %q should be valid", mode))
				}
			})

			It("should pass when MergeMode is not set", func() {
				conf := &v1alpha1.LandscapeKitConfiguration{}
				errList := ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(BeEmpty())
			})

			It("should fail with an invalid MergeMode value", func() {
				invalid := v1alpha1.MergeMode("Invalid")
				conf := &v1alpha1.LandscapeKitConfiguration{
					MergeMode: &invalid,
				}

				errList := ValidateLandscapeKitConfiguration(conf)
				Expect(errList).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Type":     Equal(field.ErrorTypeNotSupported),
						"Field":    Equal("mergeMode"),
						"BadValue": Equal(invalid),
					})),
				))
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
