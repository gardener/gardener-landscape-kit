// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta_test

import (
	_ "embed"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/components/meta"
)

var (
	//go:embed testdata/expected_configmap_output_default.yaml
	expectedDefaultConfigMapOutput string
	//go:embed testdata/expected_configmap_output_newkey.yaml
	expectedConfigMapOutputWithNewKey string

	//go:embed testdata/manifest-1-default.yaml
	manifestDefault string
	//go:embed testdata/manifest-2-edited.yaml
	manifestEdited string
	//go:embed testdata/manifest-3-new-default.yaml
	manifestDefaultNew string
	//go:embed testdata/manifest-4-expected-generated.yaml
	manifestGenerated string
)

var _ = Describe("Meta Dir Config Diff", func() {
	var fs afero.Afero

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
	})

	Describe("#CreateOrUpdateManifest", func() {
		It("should overwrite the manifest file if no meta file is present yet", func() {
			obj := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				Data: map[string]string{
					"key": "value",
				},
			}

			objYaml, err := yaml.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())

			Expect(meta.CreateOrUpdateManifest(objYaml, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

			content, err := fs.ReadFile("/landscape/.glk/defaults/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(expectedDefaultConfigMapOutput))

			content, err = fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(expectedDefaultConfigMapOutput))
		})

		It("should patch only changed default values on subsequent generates and retain custom modifications", func() {
			obj := &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				Data: map[string]string{
					"key": "value",
				},
			}

			objYaml, err := yaml.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())

			Expect(meta.CreateOrUpdateManifest(objYaml, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

			// Modify the manifest on disk
			content, err := fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			modifiedContent := []byte(strings.ReplaceAll(string(content), "value", "changedValue"))
			Expect(fs.WriteFile("/landscape/manifest/config.yaml", modifiedContent, 0600)).To(Succeed())

			// Patch the default object and generate again
			obj = obj.DeepCopy()
			obj.Data = map[string]string{
				"key":    "value",
				"newKey": "anotherValue",
			}

			objYaml, err = yaml.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())

			Expect(meta.CreateOrUpdateManifest(objYaml, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

			content, err = fs.ReadFile("/landscape/.glk/defaults/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(expectedConfigMapOutputWithNewKey))

			content, err = fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(strings.ReplaceAll(expectedConfigMapOutputWithNewKey, "key: value", "key: changedValue")))
		})

		It("should support patching raw yaml manifests with comments", func() {
			Expect(meta.CreateOrUpdateManifest([]byte(manifestDefault), "/landscape", "manifest/config.yaml", fs)).To(Succeed())

			content, err := fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal(manifestDefault))

			// Modify the manifest on disk
			Expect(fs.WriteFile("/landscape/manifest/config.yaml", []byte(manifestEdited), 0600)).To(Succeed())

			// Run the generation again with an updated default manifest
			Expect(meta.CreateOrUpdateManifest([]byte(manifestDefaultNew), "/landscape", "manifest/config.yaml", fs)).To(Succeed())

			content, err = fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal(manifestGenerated))
		})

		It("should handle a non-existent default file gracefully", func() {
			// User manifest is there first, no default file yet
			Expect(fs.WriteFile("/landscape/manifest/config.yaml", []byte(strings.ReplaceAll(expectedDefaultConfigMapOutput, "key: value", "key: newDefaultValue")), 0600)).To(Succeed())

			// Run the generation
			Expect(meta.CreateOrUpdateManifest([]byte(expectedConfigMapOutputWithNewKey), "/landscape", "manifest/config.yaml", fs)).To(Succeed())

			// The previous user manifest content should have been retained, only new keys added
			content, err := fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal(strings.ReplaceAll(expectedConfigMapOutputWithNewKey, "key: value", "key: newDefaultValue") + "\n"))
		})
	})
})
