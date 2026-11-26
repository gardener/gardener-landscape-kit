// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package meta_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/meta"
)

var _ = Describe("Three-way merge regressions", func() {
	const header = "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: test\n"

	DescribeTable("preserves mapping order and is stable on subsequent writes",
		func(oldDefault, newDefault, current, expected string) {
			result, err := meta.ThreeWayMergeManifest([]byte(oldDefault), []byte(newDefault), []byte(current), configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(result)).To(Equal(expected))

			repeated, err := meta.ThreeWayMergeManifest([]byte(newDefault), []byte(newDefault), result, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(repeated)).To(Equal(expected))
		},
		Entry("a user-only key reorder",
			"z: 1\na: 2\nm: 3\n", "z: 1\na: 2\nm: 3\n", "m: 3\nz: 1\na: 2\n", "m: 3\nz: 1\na: 2\n"),
		Entry("a default-only key reorder does not move existing keys",
			"z: 1\na: 2\nm: 3\n", "a: 2\nm: 3\nz: 1\n", "z: 1\na: 2\nm: 3\n", "z: 1\na: 2\nm: 3\n"),
		Entry("top-level Kubernetes keys with a default value update",
			header+"data:\n  value: old\n", header+"data:\n  value: new\n",
			"metadata:\n  name: test\ndata:\n  value: old\nkind: ConfigMap\napiVersion: v1\n",
			"metadata:\n  name: test\ndata:\n  value: new\nkind: ConfigMap\napiVersion: v1\n"),
		Entry("nested keys, interleaved user additions, deletions, and new defaults",
			header+"data:\n  z: 1\n  a: 2\n  removed: old\n  deleted: old\n",
			header+"data:\n  a: 3\n  z: 1\n  deleted: new\n  b: 4\n  c: 5\n",
			header+"data:\n  # User note\n  custom: 9\n  a: 2\n  removed: user\n  z: 1\n",
			header+"data:\n  # User note\n  custom: 9\n  a: 3\n  z: 1\n  b: 4\n  c: 5\n"),
		Entry("existing keys without an old default",
			"", header+"data:\n  a: default\n  z: default\n  b: new\n", header+"data:\n  z: user\n  custom: user\n  a: user\n", header+"data:\n  z: user\n  custom: user\n  a: user\n  b: new\n"),
		Entry("first generation follows the default order",
			"", header+"data:\n  z: 1\n  a: 2\n", "", header+"data:\n  z: 1\n  a: 2\n"),
		Entry("mappings inside keyed sequences",
			header+"items:\n- name: item\n  z: 1\n  a: 2\n",
			header+"items:\n- name: item\n  a: 3\n  z: 1\n  b: 4\n",
			header+"items:\n- z: 1\n  custom: user\n  name: item\n  a: 2\n",
			header+"items:\n- z: 1\n  custom: user\n  name: item\n  a: 3\n  b: 4\n"),
		Entry("mappings inside positional sequences",
			header+"items:\n- z: 1\n  a: 2\n", header+"items:\n- a: 3\n  z: 1\n", header+"items:\n- a: 2\n  z: 1\n", header+"items:\n- a: 3\n  z: 1\n"),
		Entry("unchanged keyed lists follow default item order but preserve mapping key order",
			header+"items:\n- name: first\n  z: 1\n  a: 2\n- name: second\n  z: 3\n  a: 4\n",
			header+"items:\n- name: second\n  a: 5\n  z: 3\n- name: added\n- name: first\n  a: 2\n  z: 1\n",
			header+"items:\n- name: first\n  z: 1\n  a: 2\n- name: second\n  z: 3\n  a: 4\n",
			header+"items:\n- name: second\n  z: 3\n  a: 5\n- name: added\n- name: first\n  z: 1\n  a: 2\n"),
		Entry("unchanged positional lists preserve mapping key order",
			header+"items:\n- z: 1\n  a: 2\n", header+"items:\n- a: 3\n  z: 1\n", header+"items:\n- z: 1\n  a: 2\n", header+"items:\n- z: 1\n  a: 3\n"),
	)

	DescribeTable("handles type changes without losing user edits",
		func(oldDefault, newDefault, current, expected string) {
			result, err := meta.ThreeWayMergeManifest([]byte(header+oldDefault), []byte(header+newDefault), []byte(header+current), configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(result)).To(Equal(header + expected))
		},
		Entry("an integer default becomes a string",
			"data:\n  value: 1\n  other: old\n", "data:\n  value: \"1\"\n  other: old\n", "data:\n  value: 1\n  other: user\n", "data:\n  value: \"1\"\n  other: user\n"),
		Entry("a user changes only the scalar type",
			"data:\n  value: 1\n", "data:\n  value: 2\n", "data:\n  value: \"1\"\n", "data:\n  value: \"1\"\n"),
		Entry("sequence to mapping, both changed",
			"data:\n- old\n", "data:\n  value: new\n", "data:\n  value: user\n", "data:\n  value: user\n"),
		Entry("mapping to sequence, both changed",
			"data:\n  value: old\n", "data:\n- new\n", "data:\n- user\n", "data:\n- user\n"),
		Entry("sequence to mapping, only default changed",
			"data:\n- old\n", "data:\n  value: new\n", "data:\n- old\n", "data:\n  value: new\n"),
		Entry("mapping to sequence, only default changed",
			"data:\n  value: old\n", "data:\n- new\n", "data:\n  value: old\n", "data:\n- new\n"),
		Entry("scalar types in a modified set-based list",
			"items:\n- 1\n", "items:\n- \"1\"\n", "items:\n- 1\n- user\n", "items:\n- user\n- \"1\"\n"),
		Entry("a collection tag changes in the default",
			"data:\n  value: old\n", "data: !custom\n  value: new\n", "data:\n  value: old\n", "data: !custom\n  value: new\n"),
	)

	It("does not duplicate converged comments", func() {
		oldDefault := []byte(header + "data:\n  # old\n  value: old\n")
		newDefault := []byte(header + "data:\n  # new\n  value: old\n")
		current := []byte(header + "data:\n  # new\n  value: user\n")
		result, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeSilent)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(Equal(string(current)))
	})

	DescribeTable("merges line and foot comments independently of values",
		func(oldDefault, newDefault, current, expected string) {
			result, err := meta.ThreeWayMergeManifest([]byte(header+oldDefault), []byte(header+newDefault), []byte(header+current), configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(result)).To(Equal(header + expected))

			repeated, err := meta.ThreeWayMergeManifest([]byte(header+newDefault), []byte(header+newDefault), result, configv1alpha1.MergeModeSilent)
			Expect(err).NotTo(HaveOccurred())
			Expect(repeated).To(Equal(result))
		},
		Entry("converged line comments",
			"data:\n  value: old # old\n", "data:\n  value: old # new\n", "data:\n  value: user # new\n", "data:\n  value: user # new\n"),
		Entry("converged foot comments",
			"data:\n  value: old\n\n# old\n", "data:\n  value: old\n\n# new\n", "data:\n  value: user\n\n# new\n", "data:\n  value: user\n\n# new\n"),
		Entry("a default value update preserves the user's comment",
			"data:\n  value: old # default\n", "data:\n  value: new # default\n", "data:\n  value: old # user\n", "data:\n  value: new # user\n"),
	)

	It("keeps conflict hints stable and removes them when values converge", func() {
		oldDefault := []byte(header + "data:\n  value: old\n")
		newDefault := []byte(header + "data:\n  value: new\n")
		current := []byte(header + "data:\n  custom: user\n  value: user # User note\n")
		result, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeHint)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(Equal(header + "data:\n  custom: user\n  value: user # User note  # Attention - new default: new\n"))

		repeated, err := meta.ThreeWayMergeManifest(newDefault, newDefault, result, configv1alpha1.MergeModeHint)
		Expect(err).NotTo(HaveOccurred())
		Expect(repeated).To(Equal(result))

		converged := []byte(header + "data:\n  custom: user\n  value: new # User note  # Attention - new default: new\n")
		result, err = meta.ThreeWayMergeManifest(newDefault, newDefault, converged, configv1alpha1.MergeModeHint)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(Equal(header + "data:\n  custom: user\n  value: new # User note\n"))
	})

	It("annotates conflicting collection type changes", func() {
		oldDefault := []byte(header + "data:\n- old\n")
		newDefault := []byte(header + "data:\n  value: new\n")
		current := []byte(header + "data:\n  value: user\n")
		result, err := meta.ThreeWayMergeManifest(oldDefault, newDefault, current, configv1alpha1.MergeModeHint)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(result)).To(Equal(header + "# Attention - new default: (complex node changed)\ndata:\n  value: user\n"))
	})
})
