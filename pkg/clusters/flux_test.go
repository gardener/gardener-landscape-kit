// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package clusters_test

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/clusters"
)

var _ = Describe("Flux", func() {
	var (
		fs  afero.Afero
		log = logr.Discard()
	)

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
	})

	It("should generate the flux instance and operator resource set manifests", func() {
		Expect(clusters.GenerateFluxSystemCluster(log, "/landscape", fs)).To(Succeed())

		for _, dir := range []string{
			"/landscape/.glk/defaults/clusters/runtime/flux-system/flux-instance.yaml",
			"/landscape/clusters/runtime/flux-system/flux-instance.yaml",

			"/landscape/.glk/defaults/clusters/runtime/flux-system/flux-operator.yaml",
			"/landscape/clusters/runtime/flux-system/flux-operator.yaml",
		} {
			exists, err := fs.Exists(dir)
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
		}
	})
})
