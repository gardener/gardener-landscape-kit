// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helpers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gardener/gardener-landscape-kit/pkg/ocm/components/helpers"
)

var _ = Describe("DashToCamelCase", func() {
	DescribeTable("converting dash-case to camelCase",
		func(input, expected string) {
			Expect(DashToCamelCase(input)).To(Equal(expected))
		},
		Entry("single word", "foo", "foo"),
		Entry("two words", "foo-bar", "fooBar"),
		Entry("multiple words", "admission-calico-runtime", "admissionCalicoRuntime"),
	)
})

var _ = Describe("DashToCamelCaseForMapKeys", func() {
	It("should convert nested map keys recursively", func() {
		input := map[string]any{
			"outer-key": map[string]any{
				"inner-key": "value",
				"deep-nest": map[string]any{
					"very-deep": "nested-value",
				},
				"word": "value3",
			},
		}

		result := DashToCamelCaseForMapKeys(input)

		outerMap := result["outerKey"].(map[string]any)
		Expect(outerMap).To(HaveKeyWithValue("innerKey", "value"))
		Expect(outerMap).To(HaveKeyWithValue("word", "value3"))
		deepMap := outerMap["deepNest"].(map[string]any)
		Expect(deepMap).To(HaveKeyWithValue("veryDeep", "nested-value"))
		Expect(outerMap).To(HaveLen(3))
	})

	It("should handle empty map", func() {
		input := map[string]any{}
		result := DashToCamelCaseForMapKeys(input)
		Expect(result).To(BeEmpty())
	})
})
