// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package kustomization_test

import (
	_ "embed"
	"path/filepath"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	. "github.com/gardener/gardener-landscape-kit/pkg/utils/kustomization"
)

var _ = Describe("Kustomization", func() {
	Describe("#WriteKustomizationComponent", func() {
		var (
			fs afero.Afero

			obj     *corev1.ConfigMap
			objYaml []byte
		)

		BeforeEach(func() {
			fs = afero.Afero{Fs: afero.NewMemMapFs()}

			obj = &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				Data: map[string]string{
					"key": "value",
				},
			}

			var err error
			objYaml, err = yaml.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should write a kustomization component", func() {
			var (
				landscapeDir = "/landscape"
				componentDir = "component/dir"

				objects = map[string][]byte{
					"configmap.yaml": objYaml,
				}
			)

			Expect(WriteKustomizationComponent(objects, landscapeDir, componentDir, fs)).To(Succeed())

			contents, err := fs.ReadFile(filepath.Join(landscapeDir, componentDir, "configmap.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(MatchYAML(objYaml))

			contents, err = fs.ReadFile(filepath.Join(landscapeDir, componentDir, "kustomization.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring("- configmap.yaml"))
		})
	})

	Describe("#writeLandscapeComponentsKustomizations", func() {
		var (
			fs   afero.Afero
			opts components.Options
		)

		BeforeEach(func() {
			fs = afero.Afero{Fs: afero.NewMemMapFs()}
			opts = components.NewOptions(&generateoptions.Options{
				Options: &cmd.Options{
					Log: logr.Discard(),
				},
				TargetDirPath: "/landscapeDir",
			}, fs)
		})

		It("should generate kustomization files within a component directory", func() {
			generateExampleComponentsDirectory(fs, opts)

			Expect(WriteLandscapeComponentsKustomizations(opts)).To(Succeed())

			content, err := fs.ReadFile(opts.GetTargetPath() + "/components/kustomization.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- gardener\n"))

			content, err = fs.ReadFile(opts.GetTargetPath() + "/components/gardener/kustomization.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- operator/flux-kustomization.yaml\n"))

			exists, err := fs.Exists("/components/gardener/operator/kustomization.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())

			content, err = fs.ReadFile(opts.GetTargetPath() + "/components/gardener/operator/resources/kustomization.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal("apiVersion: dummy"))
		})
	})
})

func generateExampleComponentsDirectory(fs afero.Afero, opts components.Options) {
	operatorDir := opts.GetTargetPath() + "/components/gardener/operator"
	ExpectWithOffset(1, fs.MkdirAll(operatorDir, 0700)).To(Succeed())
	ExpectWithOffset(1, fs.WriteFile(operatorDir+"/flux-kustomization.yaml", []byte(`apiVersion: kustomize.config.k8s.io/v1beta1`), 0600)).To(Succeed())

	ExpectWithOffset(1, fs.MkdirAll(operatorDir+"/resources", 0700)).To(Succeed())
	ExpectWithOffset(1, fs.WriteFile(operatorDir+"/resources/kustomization.yaml", []byte(`apiVersion: dummy`), 0600)).To(Succeed())
}
