// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta_test

import (
	_ "embed"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	Describe("#ThreeWayMergeManifest", func() {
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

			newContents, err := meta.ThreeWayMergeManifest(nil, objYaml, nil)
			Expect(err).NotTo(HaveOccurred())

			// Modify the manifest on disk
			content := []byte(strings.ReplaceAll(string(newContents), "value", "changedValue"))

			// Patch the default object and generate again
			obj = obj.DeepCopy()
			obj.Data = map[string]string{
				"key":    "value",
				"newKey": "anotherValue",
			}

			newObjYaml, err := yaml.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())

			content, err = meta.ThreeWayMergeManifest(objYaml, newObjYaml, content)
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(MatchYAML(strings.ReplaceAll(expectedConfigMapOutputWithNewKey, "key: value", "key: changedValue")))
		})

		It("should support patching raw yaml manifests with comments", func() {
			mergedManifest, err := meta.ThreeWayMergeManifest([]byte(manifestDefault), []byte(manifestDefaultNew), []byte(manifestEdited))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(mergedManifest)).To(Equal(manifestGenerated))
		})

		It("should handle a non-existent default file gracefully", func() {
			content, err := meta.ThreeWayMergeManifest(nil, []byte(expectedConfigMapOutputWithNewKey), []byte(strings.ReplaceAll(expectedDefaultConfigMapOutput, "key: value", "key: newDefaultValue")))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal(strings.ReplaceAll(expectedConfigMapOutputWithNewKey, "key: value", "key: newDefaultValue") + "\n"))
		})
	})
})
