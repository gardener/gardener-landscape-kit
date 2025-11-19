// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"bytes"
	"slices"

	"go.yaml.in/yaml/v4"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities"
)

// ThreeWayMergeManifest creates or updates a manifest based on a given YAML object.
// It performs a three-way merge between the old default template, the new default template, and the current user-modified version.
// It preserves user modifications while applying updates from the new default template.
// Contents from the current manifest are prioritized and sorted first.
func ThreeWayMergeManifest(oldDefaultYaml, newDefaultYaml, currentYaml []byte) ([]byte, error) {
	var (
		output []byte

		diff = newManifestDiff(oldDefaultYaml, newDefaultYaml, currentYaml)
	)

	for sect := range diff.current.Sections() {
		if sect.IsComment() {
			output = addWithSeparator(output, sect.Content)
			continue
		}

		current := sect.Content
		newDefault := diff.newDefault.Splits[sect.Key]
		oldDefault := diff.oldDefault.Splits[sect.Key]
		merged, err := threeWayMergeSection(oldDefault, newDefault, current)
		if err != nil {
			return nil, err
		}
		output = addWithSeparator(output, merged)
	}

	appendix := collectAppendix(diff)
	for _, extra := range appendix {
		output = addWithSeparator(output, extra)
	}

	// Ensure output ends with a newline for readability
	if len(output) > 0 && output[len(output)-1] != '\n' {
		output = append(output, '\n')
	}
	return output, nil
}

func threeWayMergeSection(oldDefaultYaml, newDefaultYaml, currentYaml []byte) ([]byte, error) {
	// Parse all three versions
	var oldDefault, newDefault, current yaml.Node
	if err := yaml.Unmarshal(newDefaultYaml, &newDefault); err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(currentYaml, &current); err != nil {
		return nil, err
	}

	// If no old default exists, use empty node (will cause all existing keys to be treated as user-added)
	if len(oldDefaultYaml) > 0 {
		if err := yaml.Unmarshal(oldDefaultYaml, &oldDefault); err != nil {
			return nil, err
		}
	}

	return encodeResult(threeWayMerge(&oldDefault, &newDefault, &current))
}

func encodeResult(merged *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	defer encoder.Close()
	encoder.SetIndent(2)
	if err := encoder.Encode(merged); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// threeWayMerge performs a three-way merge of YAML nodes
// oldDefault: the previous default template
// newDefault: the new default template
// current: the user's current version (possibly modified)
func threeWayMerge(oldDefault, newDefault, current *yaml.Node) *yaml.Node {
	// Unwrap document nodes
	if oldDefault.Kind == yaml.DocumentNode {
		oldDefault = oldDefault.Content[0]
	}
	if newDefault.Kind == yaml.DocumentNode {
		newDefault = newDefault.Content[0]
	}
	if current.Kind == yaml.DocumentNode {
		return &yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{threeWayMerge(oldDefault, newDefault, current.Content[0])},
		}
	}

	// If current equals oldDefault (including comments), no user modifications were made - use newDefault
	if nodesEqual(oldDefault, current, true) {
		return newDefault
	}

	// Build maps for easier lookup (we only handle mappings for Kubernetes manifests)
	oldMap := buildMap(oldDefault)
	currentMap := buildMap(current)
	newMap := buildMap(newDefault)

	// Create result node preserving current's comments
	result := &yaml.Node{
		Kind:        yaml.MappingNode,
		Style:       newDefault.Style,
		Tag:         newDefault.Tag,
		HeadComment: current.HeadComment,
		LineComment: current.LineComment,
		FootComment: current.FootComment,
	}

	// Process keys from newDefault
	for i := 0; i < len(newDefault.Content); i += 2 {
		newKeyNode, newValueNode := newDefault.Content[i], newDefault.Content[i+1]
		key := newKeyNode.Value
		oldValue, oldExists := oldMap[key]
		currentValue, currentExists := currentMap[key]

		var resultKeyNode, resultValue *yaml.Node

		if oldExists && !currentExists {
			// Has been dropped from current.
			continue
		}
		if !currentExists {
			// New key - add from newDefault
			resultKeyNode, resultValue = newKeyNode, newValueNode
		} else {
			resultKeyNode = findKeyNode(current, key)

			// Handle nested structures (mappings and sequences)
			switch {
			case currentValue.Kind == yaml.MappingNode && newValueNode.Kind == yaml.MappingNode:
				if !oldExists {
					oldValue = &yaml.Node{Kind: yaml.MappingNode}
				}
				resultValue = threeWayMerge(oldValue, newValueNode, currentValue)
			case currentValue.Kind == yaml.SequenceNode && newValueNode.Kind == yaml.SequenceNode:
				if !oldExists {
					oldValue = &yaml.Node{Kind: yaml.SequenceNode}
				}
				resultValue = threeWayMergeSequence(oldValue, newValueNode, currentValue)
			case oldExists && !nodesEqual(oldValue, newValueNode, false):
				resultValue = &yaml.Node{
					Kind: newValueNode.Kind, Value: newValueNode.Value, Style: newValueNode.Style, Tag: newValueNode.Tag,
					HeadComment: currentValue.HeadComment, LineComment: currentValue.LineComment, FootComment: currentValue.FootComment,
					Content: newValueNode.Content,
				}
			default:
				resultValue = currentValue
			}
		}

		result.Content = append(result.Content, resultKeyNode, resultValue)
	}

	// Then add any keys from current that don't exist in newDefault AND didn't exist in oldDefault (user-added keys)
	for i := 0; i < len(current.Content); i += 2 {
		keyNode, valueNode := current.Content[i], current.Content[i+1]
		key := keyNode.Value

		_, existsInNew := newMap[key]
		_, existedInOld := oldMap[key]

		if !existsInNew && !existedInOld {
			// Key exists only in current (user-added) - keep it at the end
			result.Content = append(result.Content, keyNode, valueNode)
		}
	}

	return result
}

// threeWayMergeSequence performs a three-way merge of YAML sequence nodes (arrays)
// Order is preserved based on newDefault, with user additions appended at the end
func threeWayMergeSequence(oldDefault, newDefault, current *yaml.Node) *yaml.Node {
	if nodesEqual(oldDefault, current, true) {
		return newDefault
	}

	result := &yaml.Node{
		Kind:        yaml.SequenceNode,
		Style:       newDefault.Style,
		Tag:         newDefault.Tag,
		HeadComment: current.HeadComment,
		LineComment: current.LineComment,
		FootComment: current.FootComment,
	}

	// Build sets for lookup
	oldSet := make(map[string]bool)
	for _, item := range oldDefault.Content {
		oldSet[nodeToString(item)] = true
	}

	currentMap := make(map[string]*yaml.Node)
	for _, item := range current.Content {
		currentMap[nodeToString(item)] = item
	}

	newSet := make(map[string]bool)
	for _, item := range newDefault.Content {
		newSet[nodeToString(item)] = true
	}

	// Process items in newDefault order to preserve order
	for _, newItem := range newDefault.Content {
		key := nodeToString(newItem)
		if !oldSet[key] {
			// New template item - add from newDefault
			result.Content = append(result.Content, newItem)
		} else if currentItem, exists := currentMap[key]; exists {
			// Unchanged item - use current version for comments
			result.Content = append(result.Content, currentItem)
		}
		// If item was in oldSet but not in currentMap, it was removed by user - skip it
	}

	// Add user-added items (items in current but not in oldDefault) at the end
	for _, item := range current.Content {
		key := nodeToString(item)
		if !oldSet[key] && !newSet[key] {
			result.Content = append(result.Content, item)
		}
	}

	return result
}

// nodeToString converts a node to a string representation for comparison
func nodeToString(node *yaml.Node) string {
	if node.Kind == yaml.ScalarNode {
		return node.Value
	}
	// For non-scalar nodes, marshal to YAML for comparison
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	utilruntime.Must(encoder.Encode(node))
	utilruntime.Must(encoder.Close())
	return buf.String()
}

// findKeyNode finds the key node for a given key in a mapping (assumes node is a mapping)
func findKeyNode(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i]
		}
	}
	return &yaml.Node{Kind: yaml.ScalarNode, Value: key}
}

// buildMap creates a map from YAML mapping node for easier lookup (assumes node is a mapping)
func buildMap(node *yaml.Node) map[string]*yaml.Node {
	result := make(map[string]*yaml.Node)
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		value := node.Content[i+1]
		result[key] = value
	}
	return result
}

// nodesEqual checks if two YAML nodes are equal
// compareComments: if true, comments must also match; if false, only values are compared
func nodesEqual(a, b *yaml.Node, compareComments bool) bool {
	if a.Kind != b.Kind {
		return false
	}

	// Check comments if requested
	if compareComments {
		if a.HeadComment != b.HeadComment ||
			a.LineComment != b.LineComment ||
			a.FootComment != b.FootComment {
			return false
		}
	}

	switch a.Kind {
	case yaml.ScalarNode:
		return a.Value == b.Value
	case yaml.SequenceNode:
		if len(a.Content) != len(b.Content) {
			return false
		}
		// For sequences, always compare in order
		for i := range a.Content {
			if !nodesEqual(a.Content[i], b.Content[i], compareComments) {
				return false
			}
		}
		return true
	case yaml.MappingNode:
		if len(a.Content) != len(b.Content) {
			return false
		}
		if compareComments {
			// When comparing comments, order matters
			for i := range a.Content {
				if !nodesEqual(a.Content[i], b.Content[i], true) {
					return false
				}
			}
		} else {
			// When ignoring comments, use map comparison (order-independent)
			aMap := buildMap(a)
			bMap := buildMap(b)
			if len(aMap) != len(bMap) {
				return false
			}
			for key, aValue := range aMap {
				bValue, exists := bMap[key]
				if !exists || !nodesEqual(aValue, bValue, false) {
					return false
				}
			}
		}
		return true
	}
	return true
}

func splitManifestFile(combinedYaml []byte) *utilities.OrderedMap {
	var values [][]byte
	if len(combinedYaml) > 0 { // Only split if there is content
		values = bytes.Split(combinedYaml, []byte("\n---\n"))
	}
	om := &utilities.OrderedMap{
		Keys:   make([]string, len(values)),
		Splits: make(map[string][]byte),
	}
	for i, v := range values {
		var t map[string]interface{}
		err := yaml.Unmarshal(v, &t)
		key := buildKey(t)
		if err != nil || key == "" {
			key = string(v)
		}
		om.Keys[i] = key
		om.Splits[key] = v
	}
	return om
}

type manifestDiff struct {
	oldDefault, newDefault, current *utilities.OrderedMap
}

func newManifestDiff(oldDefaultYaml, newDefaultYaml, currentYaml []byte) *manifestDiff {
	return &manifestDiff{
		oldDefault: splitManifestFile(oldDefaultYaml),
		newDefault: splitManifestFile(newDefaultYaml),
		current:    splitManifestFile(currentYaml),
	}
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
func collectAppendix(diff *manifestDiff) [][]byte {
	var appendix [][]byte
	for sect := range diff.newDefault.Sections() {
		isIncludedInCurrent := slices.Contains(diff.current.Keys, sect.Key)
		isIncludedInOldDefault := slices.Contains(diff.oldDefault.Keys, sect.Key)
		if !isIncludedInCurrent && !isIncludedInOldDefault {
			appendix = append(appendix, sect.Content)
		}
	}
	return appendix
}

func buildKey(t map[string]interface{}) string {
	apiVersion, _ := t["apiVersion"].(string)
	kind, _ := t["kind"].(string)
	metadata, _ := t["metadata"].(map[string]interface{})
	name, _ := metadata["name"].(string)
	namespace, _ := metadata["namespace"].(string)
	if apiVersion == "" && kind == "" && namespace == "" && name == "" {
		return ""
	}

	return apiVersion + "/" + kind + "/" + namespace + "/" + name
}
