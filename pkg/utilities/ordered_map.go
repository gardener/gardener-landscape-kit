// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utilities

import (
	"iter"
	"slices"
)

// NewOrderedMap creates a new OrderedMap with the given length.
func NewOrderedMap[T comparable, V any]() *OrderedMap[T, V] {
	return &OrderedMap[T, V]{
		keys:   []T{},
		splits: map[T]V{},
	}
}

// OrderedMap is a map that maintains the order of keys.
type OrderedMap[T comparable, V any] struct {
	// keys are the ordered keys of the map.
	keys []T
	// splits are the values of the map.
	splits map[T]V
}

// Insert adds a new key-value pair to the ordered map.
// If a key in already exists in the map, its value will be overwritten.
func (om *OrderedMap[T, V]) Insert(key T, value V) {
	idx := slices.Index(om.keys, key)
	if idx == -1 {
		om.keys = append(om.keys, key)
	}
	om.splits[key] = value
}

// Delete removes a key and its value from the ordered map.
// It returns the removed value and a boolean indicating whether the key was found.
func (om *OrderedMap[T, V]) Delete(key T) (V, bool) {
	var value V
	idx := slices.Index(om.keys, key)
	if idx == -1 {
		return value, false
	}
	value = om.splits[key]
	om.keys = slices.Delete(om.keys, idx, idx+1)
	delete(om.splits, key)
	return value, true
}

// Get retrieves the value for a given key from the ordered map.
// It returns the value and a boolean indicating whether the key was found.
func (om *OrderedMap[T, V]) Get(key T) (V, bool) {
	value, found := om.splits[key]
	return value, found
}

// Entries returns a generator function that yields all entries of the ordered map.
func (om *OrderedMap[T, V]) Entries() iter.Seq2[T, V] {
	return func(yield func(T, V) bool) {
		for _, key := range om.keys {
			if !yield(key, om.splits[key]) {
				return
			}
		}
	}
}
