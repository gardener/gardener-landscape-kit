// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package github_test

import (
	"bytes"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	dotgithub "github.com/gardener/gardener-landscape-kit/.github"
	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	. "github.com/gardener/gardener-landscape-kit/pkg/components/github"
)

var _ = Describe("Component Generation", func() {
	var (
		memFs        afero.Afero
		generateOpts *generateoptions.Options
	)

	BeforeEach(func() {
		memFs = afero.Afero{Fs: afero.NewMemMapFs()}
		generateOpts = &generateoptions.Options{
			Options: &cmd.Options{Log: logr.Discard()},
			Config:  &v1alpha1.LandscapeKitConfiguration{},
		}
		v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(generateOpts.Config)
	})

	Describe("#Name", func() {
		It("returns 'github'", func() {
			Expect(NewComponent().Name()).To(Equal("github"))
		})
	})

	Describe("#GenerateBase", func() {
		Context("with repositories.base.target = 'base'", func() {
			var opts components.Options

			BeforeEach(func() {
				generateOpts.TargetDirPath = "/repo"
				generateOpts.Config.Repositories = &v1alpha1.RepositoriesConfig{
					Base: &v1alpha1.BaseRepositoryConfig{Target: "base"},
				}
				v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(generateOpts.Config)

				var err error
				opts, err = components.NewOptions(generateOpts, memFs)
				Expect(err).NotTo(HaveOccurred())
			})

			It("writes both .github files at the repository root", func() {
				Expect(NewComponent().GenerateBase(opts)).To(Succeed())

				for _, p := range []string{
					"/repo/.github/actions/glk/action.yaml",
					"/repo/.github/workflows/glk.yaml",
				} {
					exists, err := memFs.Exists(p)
					Expect(err).NotTo(HaveOccurred(), "checking existence of %s", p)
					Expect(exists).To(BeTrue(), "expected %s to exist", p)
				}
			})

			It("writes only the expected subdirectories under .github/", func() {
				Expect(NewComponent().GenerateBase(opts)).To(Succeed())

				dirEntries, err := afero.ReadDir(memFs, "/repo/.github")
				Expect(err).NotTo(HaveOccurred())

				names := make([]string, 0, len(dirEntries))
				for _, e := range dirEntries {
					names = append(names, e.Name())
				}
				Expect(names).To(ConsistOf("actions", "workflows"))
			})

			It("does not create a .glk shadow tree under the target directory", func() {
				Expect(NewComponent().GenerateBase(opts)).To(Succeed())

				exists, err := memFs.DirExists("/repo/base/.glk")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse(), "github component must not write a .glk/defaults shadow tree")
			})

			It("prepends the disclaimer header and preserves the embedded bytes verbatim", func() {
				Expect(NewComponent().GenerateBase(opts)).To(Succeed())

				for _, tc := range []struct {
					writtenPath   string
					embeddedFS    interface{ ReadFile(string) ([]byte, error) }
					embeddedEntry string
				}{
					{
						writtenPath:   "/repo/.github/actions/glk/action.yaml",
						embeddedFS:    dotgithub.DotGitHubActions,
						embeddedEntry: "actions/glk/action.yaml",
					},
					{
						writtenPath:   "/repo/.github/workflows/glk.yaml",
						embeddedFS:    dotgithub.DotGitHubWorkflows,
						embeddedEntry: "workflows/glk.yaml",
					},
				} {
					written, err := memFs.ReadFile(tc.writtenPath)
					Expect(err).NotTo(HaveOccurred(), "reading %s", tc.writtenPath)

					embedded, err := tc.embeddedFS.ReadFile(tc.embeddedEntry)
					Expect(err).NotTo(HaveOccurred())

					Expect(bytes.HasPrefix(written, []byte(DisclaimerHeader))).To(BeTrue(), "expected %s to start with the disclaimer header", tc.writtenPath)
					Expect(bytes.TrimPrefix(written, []byte(DisclaimerHeader))).To(Equal(embedded), "bytes after disclaimer in %s must equal embedded source byte-for-byte", tc.writtenPath)
				}
			})

			It("preserves GitHub Actions ${{ ... }} expressions verbatim", func() {
				Expect(NewComponent().GenerateBase(opts)).To(Succeed())

				for _, tc := range []struct {
					writtenPath   string
					embeddedFS    interface{ ReadFile(string) ([]byte, error) }
					embeddedEntry string
				}{
					{"/repo/.github/actions/glk/action.yaml", dotgithub.DotGitHubActions, "actions/glk/action.yaml"},
					{"/repo/.github/workflows/glk.yaml", dotgithub.DotGitHubWorkflows, "workflows/glk.yaml"},
				} {
					written, err := memFs.ReadFile(tc.writtenPath)
					Expect(err).NotTo(HaveOccurred())

					embedded, err := tc.embeddedFS.ReadFile(tc.embeddedEntry)
					Expect(err).NotTo(HaveOccurred())

					for _, marker := range extractGitHubExpressions(string(embedded)) {
						Expect(string(written)).To(ContainSubstring(marker), "expected %q to be preserved in %s", marker, tc.writtenPath)
					}
				}
			})

			It("overwrites pre-existing user content", func() {
				const userContent = "# user-edited content that should be wiped\n"
				Expect(memFs.MkdirAll("/repo/.github/actions/glk", 0700)).To(Succeed())
				Expect(memFs.WriteFile("/repo/.github/actions/glk/action.yaml", []byte(userContent), 0600)).To(Succeed())

				Expect(NewComponent().GenerateBase(opts)).To(Succeed())

				written, err := memFs.ReadFile("/repo/.github/actions/glk/action.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(written)).NotTo(ContainSubstring("user-edited"))
				Expect(bytes.HasPrefix(written, []byte(DisclaimerHeader))).To(BeTrue())
			})
		})

		Context("with repositories.base.target = './base' (leading dot-slash)", func() {
			It("writes files at the same location as 'base' without the leading dot-slash", func() {
				generateOpts.TargetDirPath = "/repo"
				generateOpts.Config.Repositories = &v1alpha1.RepositoriesConfig{
					Base: &v1alpha1.BaseRepositoryConfig{Target: "./base"},
				}
				v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(generateOpts.Config)

				opts, err := components.NewOptions(generateOpts, memFs)
				Expect(err).NotTo(HaveOccurred())

				Expect(NewComponent().GenerateBase(opts)).To(Succeed())

				for _, p := range []string{
					"/repo/.github/actions/glk/action.yaml",
					"/repo/.github/workflows/glk.yaml",
				} {
					exists, err := memFs.Exists(p)
					Expect(err).NotTo(HaveOccurred(), "checking existence of %s", p)
					Expect(exists).To(BeTrue(), "expected %s to exist", p)
				}
			})
		})
	})

	Describe("#GenerateLandscape", func() {
		Context("with repositories.landscape.target = './landscapes/test' (two-level deep)", func() {
			var opts components.LandscapeOptions

			BeforeEach(func() {
				generateOpts.TargetDirPath = "/repo"
				generateOpts.Config.Repositories = &v1alpha1.RepositoriesConfig{
					Landscape: &v1alpha1.LandscapeRepositoryConfig{
						URL:    "https://github.com/gardener/gardener-ref-landscape",
						Target: "./landscapes/test",
					},
				}
				v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(generateOpts.Config)

				var err error
				opts, err = components.NewLandscapeOptions(generateOpts, memFs)
				Expect(err).NotTo(HaveOccurred())
			})

			It("writes both .github files two levels up at the repository root", func() {
				Expect(NewComponent().GenerateLandscape(opts)).To(Succeed())

				for _, p := range []string{
					"/repo/.github/actions/glk/action.yaml",
					"/repo/.github/workflows/glk.yaml",
				} {
					exists, err := memFs.Exists(p)
					Expect(err).NotTo(HaveOccurred(), "checking existence of %s", p)
					Expect(exists).To(BeTrue(), "expected %s to exist", p)
				}
			})

			It("writes only the expected subdirectories under .github/", func() {
				Expect(NewComponent().GenerateLandscape(opts)).To(Succeed())

				dirEntries, err := afero.ReadDir(memFs, "/repo/.github")
				Expect(err).NotTo(HaveOccurred())

				names := make([]string, 0, len(dirEntries))
				for _, e := range dirEntries {
					names = append(names, e.Name())
				}
				Expect(names).To(ConsistOf("actions", "workflows"))
			})

			It("does not create a .glk shadow tree under the landscape directory", func() {
				Expect(NewComponent().GenerateLandscape(opts)).To(Succeed())

				exists, err := memFs.DirExists("/repo/landscapes/test/.glk")
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse(), "github component must not write a .glk/defaults shadow tree")
			})

			It("prepends the disclaimer header and preserves the embedded bytes verbatim", func() {
				Expect(NewComponent().GenerateLandscape(opts)).To(Succeed())

				for _, tc := range []struct {
					writtenPath   string
					embeddedFS    interface{ ReadFile(string) ([]byte, error) }
					embeddedEntry string
				}{
					{"/repo/.github/actions/glk/action.yaml", dotgithub.DotGitHubActions, "actions/glk/action.yaml"},
					{"/repo/.github/workflows/glk.yaml", dotgithub.DotGitHubWorkflows, "workflows/glk.yaml"},
				} {
					written, err := memFs.ReadFile(tc.writtenPath)
					Expect(err).NotTo(HaveOccurred())

					embedded, err := tc.embeddedFS.ReadFile(tc.embeddedEntry)
					Expect(err).NotTo(HaveOccurred())

					Expect(bytes.HasPrefix(written, []byte(DisclaimerHeader))).To(BeTrue(), "expected %s to start with the disclaimer header", tc.writtenPath)
					Expect(bytes.TrimPrefix(written, []byte(DisclaimerHeader))).To(Equal(embedded), "bytes after disclaimer in %s must equal embedded source byte-for-byte", tc.writtenPath)
				}
			})
		})
	})
})

// extractGitHubExpressions returns all unique "${{ ... }}" substrings found in s.
// These are GitHub Actions expression markers that must survive verbatim through
// the write pipeline (no Go template rendering must occur).
func extractGitHubExpressions(s string) []string {
	var result []string
	seen := make(map[string]bool)
	for {
		start := strings.Index(s, "${{")
		if start < 0 {
			break
		}
		end := strings.Index(s[start:], "}}")
		if end < 0 {
			break
		}
		marker := s[start : start+end+2]
		if !seen[marker] {
			seen[marker] = true
			result = append(result, marker)
		}
		s = s[start+end+2:]
	}
	return result
}
