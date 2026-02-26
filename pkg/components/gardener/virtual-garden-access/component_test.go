// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package virtualgardenaccess_test

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	virtualgardenaccess "github.com/gardener/gardener-landscape-kit/pkg/components/gardener/virtual-garden-access"
)

var _ = Describe("Component Generation", func() {
	var (
		fs           afero.Afero
		cmdOpts      *cmd.Options
		generateOpts *generateoptions.Options
	)

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
		cmdOpts = &cmd.Options{Log: logr.Discard()}
		generateOpts = &generateoptions.Options{
			TargetDirPath: "/repo/baseDir",
			Options:       cmdOpts,
		}
	})

	Describe("#GenerateBase", func() {
		var opts components.Options

		BeforeEach(func() {
			var err error
			opts, err = components.NewOptions(generateOpts, fs)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should generate the component base", func() {
			component := virtualgardenaccess.NewComponent()
			Expect(component.GenerateBase(opts)).To(Succeed())

			for _, file := range []string{
				"/repo/baseDir/.glk/defaults/components/gardener/virtual-garden-access/virtual-garden-access-flux.yaml",
				"/repo/baseDir/components/gardener/virtual-garden-access/virtual-garden-access-flux.yaml",
			} {
				content, err := fs.ReadFile(file)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(And(
					ContainSubstring("Secret"),
					ContainSubstring("ManagedResource"),
				))
			}

			content, err := fs.ReadFile("/repo/baseDir/components/gardener/virtual-garden-access/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- virtual-garden-access-flux.yaml"))
		})
	})

	Describe("#GenerateLandscape", func() {
		BeforeEach(func() {
			generateOpts.TargetDirPath = "/repo/landscapeDir"
			generateOpts.Config = &v1alpha1.LandscapeKitConfiguration{
				Git: &v1alpha1.GitRepository{Paths: v1alpha1.PathConfiguration{Landscape: "./landscapeDir", Base: "./baseDir"}},
			}
		})

		It("should generate only the flux kustomization into the landscape dir", func() {
			component := virtualgardenaccess.NewComponent()
			landscapeOpts, err := components.NewLandscapeOptions(generateOpts, fs)
			Expect(err).ToNot(HaveOccurred())
			Expect(component.GenerateLandscape(landscapeOpts)).To(Succeed())

			exists, err := fs.DirExists("/repo/baseDir")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())

			content, err := fs.ReadFile("/repo/landscapeDir/components/gardener/virtual-garden-access/flux-kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("path: landscapeDir/components/gardener/virtual-garden-access"))

			content, err = fs.ReadFile("/repo/landscapeDir/components/gardener/virtual-garden-access/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- ../../../../baseDir/components/gardener/virtual-garden-access"))
		})
	})
})
