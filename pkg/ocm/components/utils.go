// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"fmt"
	"strings"
)

// ToFilename translates a ComponentReference into a json file name.
func (cr ComponentReference) ToFilename(dir string) string {
	return fmt.Sprintf("%s/%s.json", dir, strings.ReplaceAll(strings.ReplaceAll(string(cr), "/", "_"), ":", "-"))
}

// ExtractNameAndVersion extracts name and version from a ComponentReference.
func (cr ComponentReference) ExtractNameAndVersion() (string, string, error) {
	parts := strings.SplitN(string(cr), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid component reference format: %s", cr)
	}
	return parts[0], parts[1], nil
}

// HasName checks if the ComponentReference has the given name (ignoring version).
func (cr ComponentReference) HasName(name string) bool {
	return strings.HasPrefix(string(cr), name+":")
}

// SplitOCIImageReference splits an OCI image reference into repository and tag.
func SplitOCIImageReference(ref string) (string, string, error) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected reference '%s'", ref)
	}
	return parts[0], parts[1], nil
}
