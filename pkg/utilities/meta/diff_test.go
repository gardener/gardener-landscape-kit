// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta_test

import (
	"embed"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/meta"
)

var (
	//go:embed testdata
	testdata embed.FS
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

			expectedConfigMapOutputWithNewKey, err := testdata.ReadFile("testdata/expected_configmap_output_newkey.yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(MatchYAML(strings.ReplaceAll(string(expectedConfigMapOutputWithNewKey), "key: value", "key: changedValue")))
		})

		It("should support patching raw yaml manifests with comments", func() {
			manifestDefault, err := testdata.ReadFile("testdata/manifest-1-default.yaml")
			Expect(err).NotTo(HaveOccurred())
			manifestEdited, err := testdata.ReadFile("testdata/manifest-2-edited.yaml")
			Expect(err).NotTo(HaveOccurred())
			manifestDefaultNew, err := testdata.ReadFile("testdata/manifest-3-new-default.yaml")
			Expect(err).NotTo(HaveOccurred())
			manifestGenerated, err := testdata.ReadFile("testdata/manifest-4-expected-generated.yaml")
			Expect(err).NotTo(HaveOccurred())

			mergedManifest, err := meta.ThreeWayMergeManifest(manifestDefault, manifestDefaultNew, manifestEdited)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(mergedManifest)).To(Equal(string(manifestGenerated)))
		})

		It("should handle a non-existent default file gracefully", func() {
			expectedDefaultConfigMapOutput, err := testdata.ReadFile("testdata/expected_configmap_output_default.yaml")
			Expect(err).NotTo(HaveOccurred())
			expectedConfigMapOutputWithNewKey, err := testdata.ReadFile("testdata/expected_configmap_output_newkey.yaml")
			Expect(err).NotTo(HaveOccurred())

			content, err := meta.ThreeWayMergeManifest(nil, expectedConfigMapOutputWithNewKey, []byte(strings.ReplaceAll(string(expectedDefaultConfigMapOutput), "key: value", "key: newDefaultValue")))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(Equal(strings.ReplaceAll(string(expectedConfigMapOutputWithNewKey), "key: value", "key: newDefaultValue") + "\n"))
		})

		It("should handle multiple manifests within a single yaml file correctly", func() {
			multipleManifestsInitial, err := testdata.ReadFile("testdata/multiple-manifests-1-initial.yaml")
			Expect(err).NotTo(HaveOccurred())
			multipleManifestsEdited, err := testdata.ReadFile("testdata/multiple-manifests-2-edited.yaml")
			Expect(err).NotTo(HaveOccurred())
			multipleManifestsNewDefault, err := testdata.ReadFile("testdata/multiple-manifests-3-new-default.yaml")
			Expect(err).NotTo(HaveOccurred())
			multipleManifestsExpectedGenerated, err := testdata.ReadFile("testdata/multiple-manifests-4-expected-generated.yaml")
			Expect(err).NotTo(HaveOccurred())

			content, err := meta.ThreeWayMergeManifest(nil, multipleManifestsInitial, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(multipleManifestsInitial)))

			content, err = meta.ThreeWayMergeManifest(multipleManifestsInitial, multipleManifestsInitial, multipleManifestsInitial)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(multipleManifestsInitial)))

			// Editing the written manifest and updating the manifest with the same default content should not overwrite anything
			content, err = meta.ThreeWayMergeManifest(multipleManifestsInitial, multipleManifestsInitial, multipleManifestsEdited)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(multipleManifestsEdited)))

			// New default manifest changes should be applied, while custom edits should be retained.
			content, err = meta.ThreeWayMergeManifest(multipleManifestsInitial, multipleManifestsNewDefault, multipleManifestsEdited)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(multipleManifestsExpectedGenerated)))
		})

		It("should retain the sequence order in a currently written file", func() {
			oldDefault, err := testdata.ReadFile("testdata/order-1-old-default.yaml")
			Expect(err).NotTo(HaveOccurred())
			newDefault, err := testdata.ReadFile("testdata/order-2-new-default.yaml")
			Expect(err).NotTo(HaveOccurred())
			current, err := testdata.ReadFile("testdata/order-3-current.yaml")
			Expect(err).NotTo(HaveOccurred())
			expected, err := testdata.ReadFile("testdata/order-4-expected.yaml")
			Expect(err).NotTo(HaveOccurred())

			content, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(expected)))
		})
	})
})
