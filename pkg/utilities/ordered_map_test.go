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

var _ = Describe("Utilities", func() {
	Describe("OrderedMap", func() {
		Describe("CRUD operations", Ordered, func() {
			var om *utilities.OrderedMap[string, int]

			BeforeAll(func() {
				om = utilities.NewOrderedMap[string, int]()
			})

			It("#Insert", func() {
				om.Insert("a", 1)
				om.Insert("b", 2)
				om.Insert("c", 3)
				om.Insert("d", 4)

				Expect(collect(om)).To(Equal([]kvPair[string, int]{
					{key: "a", value: 1},
					{key: "b", value: 2},
					{key: "c", value: 3},
					{key: "d", value: 4},
				}))
			})

			It("#Delete", func() {
				om.Delete("x") // Deleting non-existing key should be no-op

				Expect(collect(om)).To(Equal([]kvPair[string, int]{
					{key: "a", value: 1},
					{key: "b", value: 2},
					{key: "c", value: 3},
					{key: "d", value: 4},
				}))

				om.Delete("d") // Delete the last entry

				Expect(collect(om)).To(Equal([]kvPair[string, int]{
					{key: "a", value: 1},
					{key: "b", value: 2},
					{key: "c", value: 3},
				}))

				om.Delete("a") // Delete the first entry

				Expect(collect(om)).To(Equal([]kvPair[string, int]{
					{key: "b", value: 2},
					{key: "c", value: 3},
				}))
			})

			It("#Get", func() {
				value, found := om.Get("b")
				Expect(found).To(BeTrue())
				Expect(value).To(Equal(2))

				value, found = om.Get("d")
				Expect(found).To(BeFalse())
				Expect(value).To(Equal(0)) // zero value for int
			})

			It("should maintain order on re-insert", func() {
				om.Insert("b", 10) // Update existing key
				om.Insert("d", 40)

				Expect(collect(om)).To(Equal([]kvPair[string, int]{
					{key: "b", value: 10},
					{key: "c", value: 3},
					{key: "d", value: 40},
				}))
			})

			It("#Keys", func() {
				om.Insert("a", 1)
				om.Insert("b", 2)
				om.Insert("c", 3)
				om.Insert("d", 4)

				Expect(slices.Collect(om.Keys())).To(ConsistOf("a", "b", "c", "d"))
			})
		})
	})
})

type kvPair[T comparable, V any] struct {
	key   T
	value V
}

func collect[T comparable, V any](om *utilities.OrderedMap[T, V]) []kvPair[T, V] {
	var result []kvPair[T, V]
	entries := om.Entries()
	entries(func(k T, v V) bool {
		result = append(result, kvPair[T, V]{key: k, value: v})
		return true
	})
	return result
}
