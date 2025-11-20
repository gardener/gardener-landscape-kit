// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Component Generation", func() {
	var (
		fs   afero.Afero
		opts Options
	)

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
		opts = NewOptions("/baseDir", "/landscapeDir", fs)
	})

	It("should generate kustomization files within a component directory", func() {
		generateExampleComponentsDirectory(fs, opts)

		Expect(writeLandscapeComponentsKustomizations(opts)).To(Succeed())

		content, err := fs.ReadFile(opts.GetLandscapeDir() + "/components/kustomization.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("- gardener\n"))

		content, err = fs.ReadFile(opts.GetLandscapeDir() + "/components/gardener/kustomization.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("- operator/flux-kustomization.yaml\n"))

		exists, err := fs.Exists("/components/gardener/operator/kustomization.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(exists).To(BeFalse())

		content, err = fs.ReadFile(opts.GetLandscapeDir() + "/components/gardener/operator/resources/kustomization.yaml")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("apiVersion: dummy"))
	})
})

func generateExampleComponentsDirectory(fs afero.Afero, opts Options) {
	operatorDir := opts.GetLandscapeDir() + "/components/gardener/operator"
	ExpectWithOffset(1, fs.MkdirAll(operatorDir, 0700)).To(Succeed())
	ExpectWithOffset(1, fs.WriteFile(operatorDir+"/flux-kustomization.yaml", []byte(`apiVersion: kustomize.config.k8s.io/v1beta1`), 0600)).To(Succeed())

	ExpectWithOffset(1, fs.MkdirAll(operatorDir+"/resources", 0700)).To(Succeed())
	ExpectWithOffset(1, fs.WriteFile(operatorDir+"/resources/kustomization.yaml", []byte(`apiVersion: dummy`), 0600)).To(Succeed())
}
