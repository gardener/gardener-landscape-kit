// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta_test

import (
	"embed"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/meta"
)

var (
	//go:embed testdata
	testdata embed.FS
)

var _ = Describe("Meta Dir Config Diff", func() {
	format.CharactersAroundMismatchToInclude = 100

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

			newContents, err := meta.ThreeWayMergeManifest(nil, objYaml, nil, configv1alpha1.MergeModeSilent)
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

			content, err = meta.ThreeWayMergeManifest(objYaml, newObjYaml, content, configv1alpha1.MergeModeSilent)
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

			mergedManifest, err := meta.ThreeWayMergeManifest(manifestDefault, manifestDefaultNew, manifestEdited, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(mergedManifest)).To(Equal(string(manifestGenerated)))
		})

		It("should handle a non-existent default file gracefully", func() {
			expectedDefaultConfigMapOutput, err := testdata.ReadFile("testdata/expected_configmap_output_default.yaml")
			Expect(err).NotTo(HaveOccurred())
			expectedConfigMapOutputWithNewKey, err := testdata.ReadFile("testdata/expected_configmap_output_newkey.yaml")
			Expect(err).NotTo(HaveOccurred())

			content, err := meta.ThreeWayMergeManifest(nil, expectedConfigMapOutputWithNewKey, []byte(strings.ReplaceAll(string(expectedDefaultConfigMapOutput), "key: value", "key: newDefaultValue")), configv1alpha1.MergeModeSilent)
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

			content, err := meta.ThreeWayMergeManifest(nil, multipleManifestsInitial, nil, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(multipleManifestsInitial)))

			content, err = meta.ThreeWayMergeManifest(multipleManifestsInitial, multipleManifestsInitial, multipleManifestsInitial, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(multipleManifestsInitial)))

			// Editing the written manifest and updating the manifest with the same default content should not overwrite anything
			content, err = meta.ThreeWayMergeManifest(multipleManifestsInitial, multipleManifestsInitial, multipleManifestsEdited, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(multipleManifestsEdited)))

			// New default manifest changes should be applied, while custom edits should be retained.
			content, err = meta.ThreeWayMergeManifest(multipleManifestsInitial, multipleManifestsNewDefault, multipleManifestsEdited, configv1alpha1.MergeModeSilent)
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

			content, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(expected)))
		})

		It("should error when invalid YAML content is provided", func() {
			var (
				err error

				emptyYaml   = []byte(``)
				validYaml   = []byte(`a: key`)
				invalidYaml = []byte(`keyWith: colonSuffix:`)
			)

			_, err = meta.ThreeWayMergeManifest(emptyYaml, invalidYaml, emptyYaml, configv1alpha1.MergeModeSilent)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("parsing newDefault file for manifest diff failed"))

			_, err = meta.ThreeWayMergeManifest(invalidYaml, validYaml, validYaml, configv1alpha1.MergeModeSilent)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("parsing oldDefault file for manifest diff failed"))

			_, err = meta.ThreeWayMergeManifest(validYaml, validYaml, invalidYaml, configv1alpha1.MergeModeSilent)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("parsing current file for manifest diff failed"))

			_, err = meta.ThreeWayMergeManifest(validYaml, validYaml, validYaml, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())

			_, err = meta.ThreeWayMergeManifest(emptyYaml, emptyYaml, emptyYaml, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("retain a completely replaced manifest content in a glk-managed file", func() {
			It("should keep the data section expanded", func() {
				oldDefault, err := testdata.ReadFile("testdata/replaced-file-1-initial.yaml")
				Expect(err).NotTo(HaveOccurred())
				newDefault, err := testdata.ReadFile("testdata/replaced-file-2-new-default.yaml")
				Expect(err).NotTo(HaveOccurred())
				current, err := testdata.ReadFile("testdata/replaced-file-3-custom.yaml")
				Expect(err).NotTo(HaveOccurred())
				expected, err := testdata.ReadFile("testdata/replaced-file-4-expected-generated.yaml")
				Expect(err).NotTo(HaveOccurred())

				content, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeSilent)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal(string(expected)))
			})

			It("should keep the data section collapsed", func() {
				oldDefault, err := testdata.ReadFile("testdata/replaced-file-3-custom.yaml")
				Expect(err).NotTo(HaveOccurred())
				newDefault, err := testdata.ReadFile("testdata/replaced-file-4-expected-generated.yaml")
				Expect(err).NotTo(HaveOccurred())
				current, err := testdata.ReadFile("testdata/replaced-file-1-initial.yaml")
				Expect(err).NotTo(HaveOccurred())
				expected, err := testdata.ReadFile("testdata/replaced-file-2-new-default.yaml")
				Expect(err).NotTo(HaveOccurred())

				content, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeSilent)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal(string(expected)))
			})
		})

		It("should handle non-kubernetes YAML manifests without apiVersion/kind/name", func() {
			// Non-kubernetes YAML uses a checksum of sorted top-level keys as the manifest key.
			// Two documents with the same structure should be treated as the same manifest across generations.
			nonK8sYaml := []byte("foo: bar\nbaz: qux\n")

			content, err := meta.ThreeWayMergeManifest(nonK8sYaml, nonK8sYaml, nonK8sYaml, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(nonK8sYaml)))

			// A new default changing a value should be reflected while a user modification is retained.
			edited := []byte("foo: user-value\nbaz: qux\n")
			newDefault := []byte("foo: bar\nbaz: updated\n")
			expected := []byte("foo: user-value\nbaz: updated\n")

			content, err = meta.ThreeWayMergeManifest(nonK8sYaml, newDefault, edited, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(MatchYAML(string(expected)))
		})

		It("should merge single manifest files regardless different namespace and name", func() {
			oldDefault, err := testdata.ReadFile("testdata/replaced-file-1-initial.yaml")
			Expect(err).NotTo(HaveOccurred())
			newDefault, err := testdata.ReadFile("testdata/replaced-file-4-expected-generated.yaml")
			Expect(err).NotTo(HaveOccurred())
			current, err := testdata.ReadFile("testdata/replaced-file-5-different-name.yaml")
			Expect(err).NotTo(HaveOccurred())
			expected, err := testdata.ReadFile("testdata/replaced-file-6-different-name-merged.yaml")
			Expect(err).NotTo(HaveOccurred())

			content, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(expected)))
		})

		It("should retain user modifications in slices during a three-way-merge", func() {
			oldDefault, err := testdata.ReadFile("testdata/merge-slice-1-default.yaml")
			Expect(err).NotTo(HaveOccurred())
			newDefault, err := testdata.ReadFile("testdata/merge-slice-3-new-default.yaml")
			Expect(err).NotTo(HaveOccurred())
			current, err := testdata.ReadFile("testdata/merge-slice-2-edited.yaml")
			Expect(err).NotTo(HaveOccurred())
			expected, err := testdata.ReadFile("testdata/merge-slice-4-expected-generated.yaml")
			Expect(err).NotTo(HaveOccurred())

			content, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content)).To(Equal(string(expected)))
		})
	})

	Describe("#ThreeWayMergeManifest - MergeModeHint", func() {
		It("should annotate a scalar value that the operator overrode and GLK updated, but not when there is no conflict", func() {
			oldDefault := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.0
`)
			newDefault := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.1.0
`)
			// Operator pinned to v1.0.5 — conflicts with GLK's new default v1.1.0
			current := []byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.5
`)
			result, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeHint)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(result)).To(ContainSubstring("version: v1.0.5"))
			Expect(string(result)).To(ContainSubstring("# Attention - new default: v1.1.0"))

			// No conflict: operator did not change the value → new default taken silently, no annotation
			result, err = meta.ThreeWayMergeManifest(oldDefault, newDefault, oldDefault, configv1alpha1.MergeModeHint)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(result)).To(ContainSubstring("version: v1.1.0"))
			Expect(string(result)).NotTo(ContainSubstring(meta.GLKDefaultPrefix))
		})
	})
})
