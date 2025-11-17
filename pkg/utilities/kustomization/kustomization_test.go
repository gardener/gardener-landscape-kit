// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package kustomization_test

import (
	_ "embed"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/kustomization"
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

			Expect(kustomization.WriteKustomizationComponent(objects, landscapeDir, componentDir, fs)).To(Succeed())

			contents, err := fs.ReadFile(filepath.Join(landscapeDir, componentDir, "configmap.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(MatchYAML(objYaml))

			contents, err = fs.ReadFile(filepath.Join(landscapeDir, componentDir, "kustomization.yaml"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(ContainSubstring("- configmap.yaml"))
		})
	})
})
