// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package utilities

// OrderedMap is a map that maintains the order of keys.
type OrderedMap struct {
	// Keys are the ordered keys of the map.
	Keys []string
	// Splits are the values of the map.
	Splits map[string][]byte
}

// Sections returns a generator function that yields the sections of the ordered map.
func (om *OrderedMap) Sections() func(yield func(*Section) bool) {
	return func(yield func(*Section) bool) {
		for _, key := range om.Keys {
			if !yield(&Section{
				key,
				om.Splits[key],
			}) {
				return
			}
		}
	}
}

// Section represents a section in the ordered map.
type Section struct {
	// Key is the key of the section.
	Key string
	// Content is the content of the section.
	Content []byte
}

// IsComment indicates whether the section is a comment.
func (s *Section) IsComment() bool {
	return len(s.Key) > 0 && s.Key == string(s.Content)
}
