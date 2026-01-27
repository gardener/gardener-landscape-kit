// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"strings"
	"unicode"
)

// DashToCamelCase converts string from dash-case to camelCase.
func DashToCamelCase(s string) string {
	parts := strings.Split(s, "-")
	if len(parts) <= 1 {
		return s
	}

	var result strings.Builder
	result.WriteString(parts[0]) // First part stays lowercase

	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			// Capitalize first letter of each subsequent part
			runes := []rune(parts[i])
			runes[0] = unicode.ToUpper(runes[0])
			result.WriteString(string(runes))
		}
	}

	return result.String()
}

// DashToCamelCaseForMapKeys converts all keys in the given map from dash-case to camelCase.
func DashToCamelCaseForMapKeys(m map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range m {
		newKey := DashToCamelCase(key)

		// Recursively convert nested maps
		if nestedMap, ok := value.(map[string]any); ok {
			result[newKey] = DashToCamelCaseForMapKeys(nestedMap)
		} else {
			result[newKey] = value
		}
	}
	return result
}
