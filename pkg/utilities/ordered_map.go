// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utilities

import "iter"

// OrderedMap is a map that maintains the order of keys.
type OrderedMap struct {
	// Keys are the ordered keys of the map.
	Keys []string
	// Splits are the values of the map.
	Splits map[string][]byte
}

// Entries returns a generator function that yields all entries of the ordered map.
func (om *OrderedMap) Entries() iter.Seq2[string, []byte] {
	return func(yield func(string, []byte) bool) {
		for _, key := range om.Keys {
			if !yield(key, om.Splits[key]) {
				return
			}
		}
	}
}
