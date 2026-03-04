// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v4"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/meta"
)

var _ = Describe("YAML Preprocessing", func() {
	Describe("#PreProcess and #PostProcess", func() {
		It("should add and remove markers for left-aligned comments", func() {
			input := []byte("# This is a comment\nkey: value\n# Another comment")

			// PreProcess should add markers to comment lines
			preprocessed := meta.PreProcess(input)
			Expect(string(preprocessed)).To(ContainSubstring("###LEVEL=0#"))

			// PostProcess should remove markers and re-align comments
			postprocessed := meta.PostProcess(preprocessed)
			Expect(string(postprocessed)).To(Equal(string(input)))
			Expect(string(postprocessed)).NotTo(ContainSubstring("###LEVEL="))
		})

		It("should handle empty content", func() {
			input := []byte("")

			preprocessed := meta.PreProcess(input)
			Expect(preprocessed).To(BeEmpty())

			postprocessed := meta.PostProcess(preprocessed)
			Expect(postprocessed).To(BeEmpty())
		})

		It("should handle preprocessing-1 testdata as expected", func() {
			input, err := testdata.ReadFile("testdata/preprocessing-1-input.yaml")
			Expect(err).ToNot(HaveOccurred())
			expected, err := testdata.ReadFile("testdata/preprocessing-1-expected.yaml")
			Expect(err).ToNot(HaveOccurred())

			preprocessed := meta.PreProcess(input)
			var node yaml.Node
			Expect(yaml.Unmarshal(preprocessed, &node)).ToNot(HaveOccurred())
			formatted, err := meta.EncodeResult(&node)
			Expect(err).ToNot(HaveOccurred())
			postprocessed := meta.PostProcess(formatted)
			Expect(string(postprocessed)).To(Equal(string(expected)))
		})
	})
})
