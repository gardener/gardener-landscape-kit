// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux_test

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/components/flux"
)

var _ = Describe("Flux Component Generation", func() {
	var (
		fs   afero.Afero
		opts components.LandscapeOptions
	)

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
		opts = components.NewLandscapeOptions("/landscapeDir", "/baseDir", "/landscapeDir", fs, logr.Discard())
	})

	Describe("#GenerateLandscape", func() {
		It("should correctly generate the flux landscape directory", func() {
			component := flux.NewComponent()
			Expect(component.GenerateLandscape(opts)).To(Succeed())
		})

		It("should not recreate a deleted gitignore file", func() {
			component := flux.NewComponent()
			Expect(component.GenerateLandscape(opts)).To(Succeed())
			Expect(fs.Exists("/landscapeDir/flux/flux-system/.gitignore")).To(BeTrue())

			Expect(fs.Remove("/landscapeDir/flux/flux-system/.gitignore")).To(Succeed())

			Expect(component.GenerateLandscape(opts)).To(Succeed())

			Expect(fs.Exists("/landscapeDir/flux/flux-system/.gitignore")).To(BeFalse())
		})

		It("should not reformat previously generated manifests (idempotency)", func() {
			component := flux.NewComponent()
			Expect(component.GenerateLandscape(opts)).To(Succeed())

			initialContents, err := fs.ReadFile("/landscapeDir/flux/flux-system/gotk-sync.yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(component.GenerateLandscape(opts)).To(Succeed())

			newContents, err := fs.ReadFile("/landscapeDir/flux/flux-system/gotk-sync.yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(initialContents)).To(Equal(string(newContents)))
		})
	})
})
