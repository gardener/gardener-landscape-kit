// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components_test

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	. "github.com/gardener/gardener-landscape-kit/pkg/components"
)

var _ = Describe("Types", func() {
	var (
		fs     afero.Afero
		logger logr.Logger
	)

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
		logger = logr.Discard()
	})

	Describe("Options", func() {
		var opts *options.Options

		BeforeEach(func() {
			opts = &options.Options{
				Options:       &cmd.Options{},
				TargetDirPath: "",
				Config:        &v1alpha1.LandscapeKitConfiguration{},
			}
			v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(opts.Config)
		})

		Describe("#GetTargetPath", func() {
			It("should return the target path", func() {
				opts.TargetDirPath = "/path/to/target"

				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetTargetPath()).To(Equal("/path/to/target"))
			})

			It("should return empty path when not set", func() {
				opts.TargetDirPath = ""

				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetTargetPath()).To(Equal("."))
			})

			It("should return a cleaned path", func() {
				opts.TargetDirPath = "/path/to/target/../landscape"

				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetTargetPath()).To(Equal("/path/to/landscape"))
			})
		})

		Describe("#GetFilesystem", func() {
			It("should return the filesystem", func() {
				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetFilesystem()).To(Equal(fs))
			})
		})

		Describe("#GetLogger", func() {
			It("should return the logger", func() {
				opts.Options = &cmd.Options{
					Log: logger,
				}

				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetLogger()).To(Equal(logger))
			})
		})

		Describe("#GetComponentVector", func() {
			var componentVectorFile string
			BeforeEach(func() {
				opts.TargetDirPath = "/path/to/target"
				componentVectorFile = opts.TargetDirPath + "/components.yaml"
			})

			It("should return an empty component vector when no component vector file is provided", func() {
				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetComponentVector()).NotTo(BeNil())

				_, exists := componentOpts.GetComponentVector().FindComponentVersion("test-component")
				Expect(exists).To(BeFalse())
			})

			It("should return an empty component vector when config is empty", func() {
				opts.Config = &v1alpha1.LandscapeKitConfiguration{}
				v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(opts.Config)

				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetComponentVector()).NotTo(BeNil())

				_, exists := componentOpts.GetComponentVector().FindComponentVersion("test-component")
				Expect(exists).To(BeFalse())
			})

			It("should return a valid component vector when a valid component vector file is provided", func() {
				componentVectorYAML := `components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
  version: v1.134.0
- name: github.com/gardener/gardener-extension-networking-cilium
  sourceRepository: https://github.com/gardener/gardener-extension-networking-cilium
  version: v1.45.0
`
				err := fs.WriteFile(componentVectorFile, []byte(componentVectorYAML), 0644)
				Expect(err).NotTo(HaveOccurred())

				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetComponentVector()).NotTo(BeNil())

				version, exists := componentOpts.GetComponentVector().FindComponentVersion("github.com/gardener/gardener")
				Expect(exists).To(BeTrue())
				Expect(version).To(Equal("v1.134.0"))

				version, exists = componentOpts.GetComponentVector().FindComponentVersion("github.com/gardener/gardener-extension-networking-cilium")
				Expect(exists).To(BeTrue())
				Expect(version).To(Equal("v1.45.0"))
			})

			It("should return an error when component vector file contains invalid YAML", func() {
				err := fs.WriteFile(componentVectorFile, []byte("invalid: yaml: content: [[["), 0644)
				Expect(err).NotTo(HaveOccurred())

				_, err = NewOptions(opts, fs)

				Expect(err).To(MatchError("failed to create component vector: failed to parse override component vector: error converting YAML to JSON: yaml: mapping values are not allowed in this context"))
			})

			It("should return a component vector with the specified version overriding the default when a partial file is provided", func() {
				partialVectorYAML := `components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
  version: v9.9.9
`
				err := fs.WriteFile(componentVectorFile, []byte(partialVectorYAML), 0644)
				Expect(err).NotTo(HaveOccurred())

				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				version, exists := componentOpts.GetComponentVector().FindComponentVersion("github.com/gardener/gardener")
				Expect(exists).To(BeTrue())
				Expect(version).To(Equal("v9.9.9"))
			})

			It("should return the default component vector when an empty file is provided", func() {
				err := fs.WriteFile(componentVectorFile, []byte(""), 0644)
				Expect(err).NotTo(HaveOccurred())

				componentOpts, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetComponentVector()).NotTo(BeNil())
			})
		})

		Describe("#NewOptions", func() {
			It("should create options with all fields", func() {
				opts := &options.Options{
					Options: &cmd.Options{
						Log: logger,
					},
					TargetDirPath: "/path/to/target",
					Config:        &v1alpha1.LandscapeKitConfiguration{},
				}
				v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(opts.Config)

				result, err := NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.GetTargetPath()).To(Equal("/path/to/target"))
				Expect(result.GetFilesystem()).To(Equal(fs))
				Expect(result.GetLogger()).To(Equal(logger))
			})
		})
	})

	Describe("LandscapeOptions", func() {
		var opts *options.Options

		BeforeEach(func() {
			opts = &options.Options{
				Options: &cmd.Options{},
				Config: &v1alpha1.LandscapeKitConfiguration{
					Repositories: &v1alpha1.RepositoriesConfig{
						Base: &v1alpha1.BaseRepositoryConfig{Target: "content"},
						Landscape: &v1alpha1.LandscapeRepositoryConfig{
							URL: "https://github.com/example/repo.git",
							Ref: v1alpha1.GitRepositoryRef{
								Branch: new("main"),
							},
							BaseLink: "base",
							Target:   "landscape",
						},
					},
				},
				TargetDirPath: "",
			}
			v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(opts.Config)
		})

		Describe("#GetLandscapeURL", func() {
			It("should return the landscape repository URL", func() {
				landscapeOpts, err := NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetLandscapeURL()).To(Equal("https://github.com/example/repo.git"))
			})
		})

		Describe("#GetLandscapeRef", func() {
			It("should return the landscape repository ref", func() {
				landscapeOpts, err := NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetLandscapeRef()).To(Equal(opts.Config.Repositories.Landscape.Ref))
			})
		})

		Describe("#GetRelativeBasePath", func() {
			It("should return baseLink", func() {
				opts.Config.Repositories.Landscape.BaseLink = "./base"

				landscapeOpts, err := NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetRelativeBasePath()).To(Equal("./base"))
			})
		})

		Describe("#GetRelativeLandscapePath", func() {
			It("should return the landscape path", func() {
				opts.Config.Repositories.Landscape.Target = "./landscape"

				landscapeOpts, err := NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetRelativeLandscapePath()).To(Equal("./landscape"))
			})
		})

		Describe("#GetRelativeBaseComponentPath", func() {
			It("should return the relative path from a landscape component dir to the corresponding base component dir", func() {
				opts.Config.Repositories.Landscape.Target = "landscapes/showroom"
				opts.Config.Repositories.Landscape.BaseLink = "base/content"

				landscapeOpts, err := NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetRelativeBaseComponentPath("gardener/garden")).To(Equal("../../../../../base/content/components/gardener/garden"))
			})

			It("should handle a single-segment landscape target", func() {
				opts.Config.Repositories.Landscape.Target = "landscape"
				opts.Config.Repositories.Landscape.BaseLink = "base/content"

				landscapeOpts, err := NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetRelativeBaseComponentPath("gardener/garden")).To(Equal("../../../../base/content/components/gardener/garden"))
			})
		})

		Describe("NewLandscapeOptions", func() {
			It("should create landscape options with all fields", func() {
				opts := &options.Options{
					Options: &cmd.Options{
						Log: logger,
					},
					TargetDirPath: "/path/to/target",
					Config: &v1alpha1.LandscapeKitConfiguration{
						Repositories: &v1alpha1.RepositoriesConfig{
							Base: &v1alpha1.BaseRepositoryConfig{Target: "base"},
							Landscape: &v1alpha1.LandscapeRepositoryConfig{
								URL: "https://github.com/example/repo.git",
								Ref: v1alpha1.GitRepositoryRef{
									Branch: new("main"),
								},
								BaseLink: "base",
								Target:   "landscape",
							},
						},
					},
				}
				v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(opts.Config)

				result, err := NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.GetTargetPath()).To(Equal("/path/to/target/landscape"))
				Expect(result.GetRepoRoot()).To(Equal("/path/to/target"))
				Expect(result.GetFilesystem()).To(Equal(fs))
				Expect(result.GetLandscapeURL()).To(Equal("https://github.com/example/repo.git"))
				Expect(result.GetLandscapeRef()).To(Equal(opts.Config.Repositories.Landscape.Ref))
				Expect(result.GetRelativeBasePath()).To(Equal("base"))
				Expect(result.GetRelativeLandscapePath()).To(Equal("landscape"))
			})
		})
	})
})
