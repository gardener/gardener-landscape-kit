// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"bytes"
	"fmt"

	"github.com/elliotchance/orderedmap/v3"
	"go.yaml.in/yaml/v4"
)

// section represents a single section in a manifest file (either a manifest or a comment)
type section struct {
	key     string
	content []byte
}

func newSection(key string, content []byte) *section {
	return &section{
		key:     key,
		content: content,
	}
}

func (s *section) isComment() bool {
	return len(s.key) > 0 && s.key == string(s.content)
}

// manifestDiff holds the three versions of manifest maps used in three-way merge
type manifestDiff struct {
	oldDefault, newDefault, current *orderedmap.OrderedMap[string, []byte]
}

// ThreeWayMergeManifest creates or updates a manifest based on a given YAML object.
// It performs a three-way merge between the old default template, the new default template, and the current user-modified version.
// It preserves user modifications while applying updates from the new default template.
// Contents from the current manifest are prioritized and sorted first.
func ThreeWayMergeManifest(oldDefaultYaml, newDefaultYaml, currentYaml []byte) ([]byte, error) {
	var (
		output []byte

		diff, err = newManifestDiff(preProcess(oldDefaultYaml), preProcess(newDefaultYaml), preProcess(currentYaml))
	)
	if err != nil {
		return nil, err
	}

	for key, value := range diff.current.AllFromFront() {
		sect := newSection(key, value)
		if sect.isComment() {
			output = addWithSeparator(output, sect.content)
			continue
		}

		current := sect.content
		newDefault, _ := diff.newDefault.Get(sect.key)
		oldDefault, _ := diff.oldDefault.Get(sect.key)
		merged, err := threeWayMergeSection(oldDefault, newDefault, current)
		if err != nil {
			return nil, err
		}
		output = addWithSeparator(output, merged)
	}

	appendix := collectAppendix(diff)
	for _, sect := range appendix {
		if sect.isComment() {
			output = addWithSeparator(output, sect.content)
			continue
		}
		// Applying threeWayMergeSection with only the new section content to ensure proper formatting (idempotency).
		merged, err := threeWayMergeSection(nil, sect.content, nil)
		if err != nil {
			return nil, err
		}
		output = addWithSeparator(output, merged)
	}

	// Ensure output ends with a newline for readability
	if len(output) > 0 && output[len(output)-1] != '\n' {
		output = append(output, '\n')
	}
	return postProcess(output), nil
}

func newManifestDiff(oldDefaultYaml, newDefaultYaml, currentYaml []byte) (*manifestDiff, error) {
	md := &manifestDiff{}
	var err error
	if md.oldDefault, err = splitManifestFile(oldDefaultYaml); err != nil {
		return nil, fmt.Errorf("parsing oldDefault file for manifest diff failed: %w", err)
	}
	if md.newDefault, err = splitManifestFile(newDefaultYaml); err != nil {
		return nil, fmt.Errorf("parsing newDefault file for manifest diff failed: %w", err)
	}
	if md.current, err = splitManifestFile(currentYaml); err != nil {
		return nil, fmt.Errorf("parsing current file for manifest diff failed: %w", err)
	}

	// Handle single manifest files with different names/namespaces
	md.normalizeSingleManifestKeys()

	return md, nil
}

// splitManifestFile splits a multi-document YAML file into separate manifests
func splitManifestFile(combinedYaml []byte) (*orderedmap.OrderedMap[string, []byte], error) {
	var values [][]byte
	if len(combinedYaml) > 0 { // Only split if there is content
		values = bytes.Split(combinedYaml, []byte("\n---\n"))
	}
	om := orderedmap.NewOrderedMap[string, []byte]()
	for _, v := range values {
		var t map[string]any
		if err := yaml.Unmarshal(v, &t); err != nil {
			return nil, err
		}
		key := buildKey(t)
		if key == "" {
			key = string(v)
		}
		om.Set(key, v)
	}
	return om, nil
}

// buildKey builds a unique key for a manifest using apiVersion/kind/namespace/name
func buildKey(t map[string]any) string {
	typeKey := buildTypeKey(t)
	metadata, _ := t["metadata"].(map[string]any)
	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)
	if typeKey == "" && namespace == "" && name == "" {
		return ""
	}

	return typeKey + "/" + namespace + "/" + name
}

// buildTypeKey builds a type key from apiVersion and kind
func buildTypeKey(t map[string]any) string {
	apiVersion, _ := t["apiVersion"].(string)
	kind, _ := t["kind"].(string)
	if apiVersion == "" || kind == "" {
		return ""
	}
	return apiVersion + "/" + kind
}

// extractTypeKey extracts "apiVersion/kind" from a manifest YAML
func extractTypeKey(yamlContent []byte) string {
	var t map[string]any
	if err := yaml.Unmarshal(yamlContent, &t); err != nil {
		return ""
	}
	return buildTypeKey(t)
}

// normalizeSingleManifestKeys handles the case where all three files contain a single manifest
// of the same type (apiVersion/kind). In this case, even if name/namespace differ,
// we normalize the keys so they match for merging. This allows users to rename resources
// while still receiving GLK updates.
func (md *manifestDiff) normalizeSingleManifestKeys() {
	// Only apply this logic if all three maps have exactly one non-comment entry
	if countNonCommentEntries(md.oldDefault) != 1 || countNonCommentEntries(md.newDefault) != 1 || countNonCommentEntries(md.current) != 1 {
		return
	}

	// Get the single entries
	oldKey, oldValue := getSingleNonCommentEntry(md.oldDefault)
	newKey, newValue := getSingleNonCommentEntry(md.newDefault)
	currentKey, currentValue := getSingleNonCommentEntry(md.current)

	// Extract apiVersion/kind from each
	oldType := extractTypeKey(oldValue)
	newType := extractTypeKey(newValue)
	currentType := extractTypeKey(currentValue)

	// If all three have the same type, normalize keys to match
	if oldType != "" && oldType == newType && oldType == currentType {
		normalizedKey := oldType + "/*/normalized"

		// Rebuild the ordered maps with normalized keys
		md.oldDefault = rebuildWithKey(md.oldDefault, oldKey, normalizedKey)
		md.newDefault = rebuildWithKey(md.newDefault, newKey, normalizedKey)
		md.current = rebuildWithKey(md.current, currentKey, normalizedKey)
	}
}

// countNonCommentEntries counts entries that are not comments (not isComment())
func countNonCommentEntries(om *orderedmap.OrderedMap[string, []byte]) int {
	count := 0
	for key, value := range om.AllFromFront() {
		sect := newSection(key, value)
		if !sect.isComment() {
			count++
		}
	}
	return count
}

// getSingleNonCommentEntry returns the single non-comment entry (assumes count == 1)
func getSingleNonCommentEntry(om *orderedmap.OrderedMap[string, []byte]) (string, []byte) {
	for key, value := range om.AllFromFront() {
		sect := newSection(key, value)
		if !sect.isComment() {
			return key, value
		}
	}
	return "", nil
}

// rebuildWithKey rebuilds an ordered map, replacing oldKey with newKey
func rebuildWithKey(om *orderedmap.OrderedMap[string, []byte], oldKey, newKey string) *orderedmap.OrderedMap[string, []byte] {
	newOM := orderedmap.NewOrderedMap[string, []byte]()
	for key, value := range om.AllFromFront() {
		if key == oldKey {
			newOM.Set(newKey, value)
		} else {
			newOM.Set(key, value)
		}
	}
	return newOM
}

func addWithSeparator(output, content []byte) []byte {
	if len(output) > 0 {
		if output[len(output)-1] != '\n' {
			output = append(output, '\n')
		}
		output = append(output, []byte("---\n")...)
	}
	return append(output, content...)
}

// collectAppendix gathers custom file content and keys not covered by current.
func collectAppendix(diff *manifestDiff) []*section {
	var appendix []*section
	for key, value := range diff.newDefault.AllFromFront() {
		sect := newSection(key, value)
		_, isIncludedInCurrent := diff.current.Get(sect.key)
		_, isIncludedInOldDefault := diff.oldDefault.Get(sect.key)
		if !isIncludedInCurrent && !isIncludedInOldDefault {
			appendix = append(appendix, sect)
		}
	}
	return appendix
}
