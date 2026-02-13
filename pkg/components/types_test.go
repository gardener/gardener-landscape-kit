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
	"github.com/gardener/gardener-landscape-kit/pkg/components"
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
			}
		})

		Describe("#GetTargetPath", func() {
			It("should return the target path", func() {
				opts.TargetDirPath = "/path/to/target"

				componentOpts, err := components.NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetTargetPath()).To(Equal("/path/to/target"))
			})

			It("should return empty path when not set", func() {
				opts.TargetDirPath = ""

				componentOpts, err := components.NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetTargetPath()).To(BeEmpty())
			})
		})

		Describe("#GetFilesystem", func() {
			It("should return the filesystem", func() {
				componentOpts, err := components.NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetFilesystem()).To(Equal(fs))
			})
		})

		Describe("#GetLogger", func() {
			It("should return the logger", func() {
				opts.Options = &cmd.Options{
					Log: logger,
				}

				componentOpts, err := components.NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetLogger()).To(Equal(logger))
			})
		})

		Describe("#GetComponentVector", func() {
			It("should return an empty component vector when no component vector file is provided", func() {
				componentOpts, err := components.NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetComponentVector()).NotTo(BeNil())

				_, exists := componentOpts.GetComponentVector().FindComponentVersion("test-component")
				Expect(exists).To(BeFalse())
			})

			It("should return an empty component vector when config is nil", func() {
				opts.Config = nil

				componentOpts, err := components.NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetComponentVector()).NotTo(BeNil())

				_, exists := componentOpts.GetComponentVector().FindComponentVersion("test-component")
				Expect(exists).To(BeFalse())
			})

			It("should return an empty component vector when VersionConfig is nil", func() {
				opts.Config = &v1alpha1.LandscapeKitConfiguration{}

				componentOpts, err := components.NewOptions(opts, fs)

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
				componentVectorFile := "/tmp/component-vector.yaml"
				err := fs.WriteFile(componentVectorFile, []byte(componentVectorYAML), 0644)
				Expect(err).NotTo(HaveOccurred())

				opts.Config = &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						ComponentsVectorFile: new(componentVectorFile),
					},
				}

				componentOpts, err := components.NewOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(componentOpts.GetComponentVector()).NotTo(BeNil())

				version, exists := componentOpts.GetComponentVector().FindComponentVersion("github.com/gardener/gardener")
				Expect(exists).To(BeTrue())
				Expect(version).To(Equal("v1.134.0"))

				version, exists = componentOpts.GetComponentVector().FindComponentVersion("github.com/gardener/gardener-extension-networking-cilium")
				Expect(exists).To(BeTrue())
				Expect(version).To(Equal("v1.45.0"))
			})

			It("should return an error when component vector file does not exist", func() {
				opts.Config = &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						ComponentsVectorFile: new("/non/existent/file.yaml"),
					},
				}

				_, err := components.NewOptions(opts, fs)

				Expect(err).To(MatchError("failed to read component vector file: open /non/existent/file.yaml: file does not exist"))
			})

			It("should return an error when component vector file contains invalid YAML", func() {
				componentVectorFile := "/tmp/invalid-component-vector.yaml"
				err := fs.WriteFile(componentVectorFile, []byte("invalid: yaml: content: [[["), 0644)
				Expect(err).NotTo(HaveOccurred())

				opts.Config = &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						ComponentsVectorFile: new(componentVectorFile),
					},
				}

				_, err = components.NewOptions(opts, fs)

				Expect(err).To(MatchError("failed to create component vector: error converting YAML to JSON: yaml: mapping values are not allowed in this context"))
			})

			It("should return an error when component vector file is empty but file exists", func() {
				componentVectorFile := "/tmp/empty-component-vector.yaml"
				err := fs.WriteFile(componentVectorFile, []byte(""), 0644)
				Expect(err).NotTo(HaveOccurred())

				opts.Config = &v1alpha1.LandscapeKitConfiguration{
					VersionConfig: &v1alpha1.VersionConfiguration{
						ComponentsVectorFile: new(componentVectorFile),
					},
				}

				componentOpts, err := components.NewOptions(opts, fs)

				Expect(err).To(MatchError("failed to create component vector: [].components: Required value: at least one component must be specified"))
				Expect(componentOpts).To(BeNil())
			})
		})

		Describe("#NewOptions", func() {
			It("should create options with all fields", func() {
				opts := &options.Options{
					Options: &cmd.Options{
						Log: logger,
					},
					TargetDirPath: "/path/to/target",
				}

				result, err := components.NewOptions(opts, fs)

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
					Git: &v1alpha1.GitRepository{
						URL: "https://github.com/example/repo.git",
						Ref: v1alpha1.GitRepositoryRef{
							Branch: new("main"),
						},
						Paths: v1alpha1.PathConfiguration{
							Base:      "base",
							Landscape: "landscape",
						},
					},
				},
				TargetDirPath: "",
			}
		})

		Describe("#GetGitRepository", func() {
			It("should return the git repository", func() {
				landscapeOpts, err := components.NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetGitRepository()).To(Equal(opts.Config.Git))
			})
		})

		Describe("#GetRelativeBasePath", func() {
			It("should return the base path", func() {
				opts.Config.Git.Paths.Base = "./base"

				landscapeOpts, err := components.NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetRelativeBasePath()).To(Equal("./base"))
			})
		})

		Describe("#GetRelativeLandscapePath", func() {
			It("should return the landscape path", func() {
				opts.Config.Git.Paths.Landscape = "./landscape"

				landscapeOpts, err := components.NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(landscapeOpts.GetRelativeLandscapePath()).To(Equal("./landscape"))
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
						Git: &v1alpha1.GitRepository{
							URL: "https://github.com/example/repo.git",
							Ref: v1alpha1.GitRepositoryRef{
								Branch: new("main"),
							},
							Paths: v1alpha1.PathConfiguration{
								Base:      "base",
								Landscape: "landscape",
							},
						},
					},
				}

				result, err := components.NewLandscapeOptions(opts, fs)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.GetTargetPath()).To(Equal("/path/to/target"))
				Expect(result.GetFilesystem()).To(Equal(fs))
				Expect(result.GetGitRepository()).To(Equal(opts.Config.Git))
				Expect(result.GetRelativeBasePath()).To(Equal("base"))
				Expect(result.GetRelativeLandscapePath()).To(Equal("landscape"))
			})
		})
	})
})
