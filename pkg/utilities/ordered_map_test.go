// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utilities_test

import (
	_ "embed"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities"
)

var _ = Describe("Utilities OrderedMap", func() {
	Describe("#Sections", func() {
		It("should iterate the sections correctly", func() {
			om := utilities.OrderedMap{
				Keys: []string{"this is a comment", "key/of/an/object"},
				Splits: map[string][]byte{
					"key/of/an/object":  []byte("object value"),
					"this is a comment": []byte("this is a comment"),
				},
			}

			Expect(slices.Collect(om.Sections())).To(ConsistOf(
				BeEquivalentTo(&utilities.Section{Key: "key/of/an/object", Content: []byte("object value")}),
				BeEquivalentTo(&utilities.Section{Key: "this is a comment", Content: []byte("this is a comment")}),
			))
		})
	})
})
