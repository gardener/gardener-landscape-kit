// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components_test

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"k8s.io/utils/ptr"

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

				componentOpts := components.NewOptions(opts, fs)

				Expect(componentOpts.GetTargetPath()).To(Equal("/path/to/target"))
			})

			It("should return empty path when not set", func() {
				opts.TargetDirPath = ""

				componentOpts := components.NewOptions(opts, fs)

				Expect(componentOpts.GetTargetPath()).To(BeEmpty())
			})
		})

		Describe("#GetFilesystem", func() {
			It("should return the filesystem", func() {
				componentOpts := components.NewOptions(opts, fs)

				Expect(componentOpts.GetFilesystem()).To(Equal(fs))
			})
		})

		Describe("#GetLogger", func() {
			It("should return the logger", func() {
				opts.Options = &cmd.Options{
					Log: logger,
				}

				componentOpts := components.NewOptions(opts, fs)

				Expect(componentOpts.GetLogger()).To(Equal(logger))
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

				result := components.NewOptions(opts, fs)

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
							Branch: ptr.To("main"),
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
				landscapeOpts := components.NewLandscapeOptions(opts, fs)

				Expect(landscapeOpts.GetGitRepository()).To(Equal(opts.Config.Git))
			})
		})

		Describe("#GetRelativeBasePath", func() {
			It("should return the base path", func() {
				opts.Config.Git.Paths.Base = "./base"

				landscapeOpts := components.NewLandscapeOptions(opts, fs)

				Expect(landscapeOpts.GetRelativeBasePath()).To(Equal("./base"))
			})
		})

		Describe("#GetRelativeLandscapePath", func() {
			It("should return the landscape path", func() {
				opts.Config.Git.Paths.Landscape = "./landscape"

				landscapeOpts := components.NewLandscapeOptions(opts, fs)

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
								Branch: ptr.To("main"),
							},
							Paths: v1alpha1.PathConfiguration{
								Base:      "base",
								Landscape: "landscape",
							},
						},
					},
				}

				result := components.NewLandscapeOptions(opts, fs)

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
