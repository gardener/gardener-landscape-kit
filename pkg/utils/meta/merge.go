// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"strings"

	"go.yaml.in/yaml/v4"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
)

const (
	// GLKDefaultPrefix is the comment prefix for GLK-managed default annotations.
	// It is exported so callers can use it as the strip anchor when removing annotations.
	GLKDefaultPrefix = "# Attention - new default: "
)

// stripGLKAnnotation removes a GLK-managed annotation from a single comment string.
// If the annotation is part of a multi-line comment, only the annotation line is removed.
// If the annotation is appended to a value's line comment (e.g. "# user note  # Attention …"), the prefix and everything after it is stripped.
func stripGLKAnnotation(comment string) string {
	if !strings.Contains(comment, GLKDefaultPrefix) {
		return comment
	}
	lines := strings.Split(comment, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		before, _, found := strings.Cut(line, GLKDefaultPrefix)
		if !found {
			out = append(out, line)
			continue
		}
		stripped := strings.TrimRight(before, " \t")
		if stripped != "" {
			out = append(out, stripped)
		}
	}
	return strings.Join(out, "\n")
}

// stripGLKAnnotations removes GLK-managed annotations from all comment fields of a node.
func stripGLKAnnotations(nodes ...*yaml.Node) {
	for _, n := range nodes {
		if n == nil {
			continue
		}
		n.HeadComment = stripGLKAnnotation(n.HeadComment)
		n.LineComment = stripGLKAnnotation(n.LineComment)
		n.FootComment = stripGLKAnnotation(n.FootComment)
	}
}

// threeWayMergeSection performs a three-way merge on a single YAML section
func threeWayMergeSection(oldDefaultYaml, newDefaultYaml, currentYaml []byte, mode configv1alpha1.MergeMode) ([]byte, error) {
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

	return EncodeResult(threeWayMerge(&oldDefault, &newDefault, &current, mode))
}

// threeWayMerge performs a three-way merge of YAML nodes
// oldDefault: the previous default template
// newDefault: the new default template
// current: the user's current version (possibly modified)
func threeWayMerge(oldDefault, newDefault, current *yaml.Node, mode configv1alpha1.MergeMode) *yaml.Node {
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
			Content: []*yaml.Node{threeWayMerge(oldDefault, newDefault, current.Content[0], mode)},
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

	// Create result node preserving current's comments and style
	result := &yaml.Node{
		Kind:        yaml.MappingNode,
		Style:       current.Style,
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

			// Three-way merge the key node's comments (e.g. HeadComment above the key).
			oldKeyNode := findKeyNode(oldDefault, key)
			mergeNodeComments(oldKeyNode, newKeyNode, resultKeyNode)

			// Handle nested structures (mappings and sequences)
			switch {
			case currentValue.Kind == yaml.MappingNode && newValueNode.Kind == yaml.MappingNode:
				if !oldExists {
					oldValue = &yaml.Node{Kind: yaml.MappingNode}
				}
				resultValue = threeWayMerge(oldValue, newValueNode, currentValue, mode)
			case currentValue.Kind == yaml.SequenceNode && newValueNode.Kind == yaml.SequenceNode:
				if !oldExists {
					oldValue = &yaml.Node{Kind: yaml.SequenceNode}
				}
				resultValue = threeWayMergeSequence(oldValue, newValueNode, currentValue, mode)
			case oldExists && !nodesEqual(oldValue, newValueNode, false) && nodesEqual(oldValue, currentValue, false):
				// Default changed and current was not modified: take the new default.
				resultValue = &yaml.Node{
					Kind: newValueNode.Kind, Value: newValueNode.Value, Style: newValueNode.Style, Tag: newValueNode.Tag,
					HeadComment: currentValue.HeadComment, LineComment: currentValue.LineComment, FootComment: currentValue.FootComment,
					Content: newValueNode.Content,
				}
				mergeNodeComments(oldValue, newValueNode, resultValue)
			case oldExists && !nodesEqual(oldValue, newValueNode, false):
				// Both default and current changed: keep current (user's value wins).
				resultValue = currentValue
				mergeNodeComments(oldValue, newValueNode, resultValue)
				if mode == configv1alpha1.MergeModeHint {
					if !nodesEqual(newValueNode, currentValue, false) {
						annotateConflict(resultKeyNode, resultValue, newValueNode)
					} else {
						// Values converged — strip any lingering GLK annotation.
						stripGLKAnnotations(resultKeyNode, resultValue)
					}
				}
			default:
				resultValue = currentValue
				if oldExists {
					mergeNodeComments(oldValue, newValueNode, resultValue)
					if mode == configv1alpha1.MergeModeHint && nodesEqual(newValueNode, currentValue, false) {
						// Values converged — strip any lingering GLK annotation.
						stripGLKAnnotations(resultKeyNode, resultValue)
					}
				}
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
			// key exists only in current (user-added) - keep it at the end
			result.Content = append(result.Content, keyNode, valueNode)
		}
	}

	return result
}

// annotateConflict adds a GLK-managed annotation comment to resultValue (or resultKeyNode for complex nodes) indicating the current GLK default.
// This is used in MergeModeHint when an operator override conflicts with an updated GLK default, so the user is hinted about the divergence.
//
// For scalar nodes, the annotation is a line comment on the value node (same line as the value).
// For complex nodes (mappings/sequences), the annotation is a head comment on the key node (line above the key).
func annotateConflict(resultKeyNode, resultValue, newDefaultNode *yaml.Node) {
	switch newDefaultNode.Kind {
	case yaml.ScalarNode:
		annotation := glkManagedLineComment(newDefaultNode.Value)
		// Strip any pre-existing GLK annotation before appending the current one, so repeated runs replace rather than accumulate the comment.
		stripped := stripGLKAnnotation(resultValue.LineComment)
		if stripped != "" {
			resultValue.LineComment = stripped + "  " + annotation
		} else {
			resultValue.LineComment = annotation
		}
	default:
		// Strip existing GLK annotation from head comment before re-annotating.
		resultKeyNode.HeadComment = stripGLKAnnotation(resultKeyNode.HeadComment)
		resultKeyNode.HeadComment = glkManagedHeadComment(resultKeyNode.HeadComment)
	}
}

// glkManagedLineComment returns a GLK-managed line comment for a scalar value conflict.
func glkManagedLineComment(newValue string) string {
	return GLKDefaultPrefix + newValue
}

// glkManagedHeadComment returns a GLK-managed head comment for a complex node conflict.
func glkManagedHeadComment(existingHead string) string {
	annotation := GLKDefaultPrefix + "(complex node changed)"
	if existingHead == "" {
		return annotation
	}
	return existingHead + "\n" + annotation
}

// mergeComment performs a three-way merge on a single comment string.
// If the default comment changed (old != new), the new default is applied — unless the user
// has also modified the comment (current != old), in which case the user's comment is kept
// and the new default comment is appended on a new line (amendment).
// If the default comment did not change (old == new), the user's comment is kept as-is.
func mergeComment(oldComment, newComment, currentComment string) string {
	if oldComment == newComment {
		// Default did not change — preserve whatever the user has.
		return currentComment
	}
	// Default changed.
	if currentComment == oldComment {
		// User didn't touch the comment — apply the new default.
		return newComment
	}
	// Both default and user changed the comment — amend: keep user's, append new default.
	if currentComment == "" {
		return newComment
	}
	if newComment == "" {
		return currentComment
	}
	return currentComment + "\n" + newComment
}

// mergeNodeComments applies three-way comment merging to all comment fields of a node,
// given the corresponding old-default and new-default nodes.
func mergeNodeComments(oldNode, newNode, resultNode *yaml.Node) {
	resultNode.HeadComment = mergeComment(oldNode.HeadComment, newNode.HeadComment, resultNode.HeadComment)
	resultNode.LineComment = mergeComment(oldNode.LineComment, newNode.LineComment, resultNode.LineComment)
	resultNode.FootComment = mergeComment(oldNode.FootComment, newNode.FootComment, resultNode.FootComment)
}

// Order is preserved based on newDefault, with user additions appended at the end
func threeWayMergeSequence(oldDefault, newDefault, current *yaml.Node, mode configv1alpha1.MergeMode) *yaml.Node {
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

	// Detect a shared identity key for the sequence items (e.g. "name").
	// When present, items are matched across old/new/current by identity so that
	// value and comment changes from the new default can be merged into the user's
	// modified item, rather than treating them as separate entries.
	identityKey := detectIdentityKey(oldDefault, newDefault, current)

	if identityKey != "" {
		// Build identity-keyed maps for each version.
		oldByID := make(map[string]*yaml.Node, len(oldDefault.Content))
		for _, item := range oldDefault.Content {
			if id := mappingValue(item, identityKey); id != "" {
				oldByID[id] = item
			}
		}
		newByID := make(map[string]*yaml.Node, len(newDefault.Content))
		for _, item := range newDefault.Content {
			if id := mappingValue(item, identityKey); id != "" {
				newByID[id] = item
			}
		}

		// Process current items in order.
		for _, currentItem := range current.Content {
			id := mappingValue(currentItem, identityKey)
			if id == "" {
				result.Content = append(result.Content, currentItem)
				continue
			}
			newItem, existsInNew := newByID[id]
			if !existsInNew {
				// Item was removed in newDefault — drop it if it also existed in old.
				if _, existsInOld := oldByID[id]; existsInOld {
					continue
				}
				// User-added item (not in old or new) — keep it.
				result.Content = append(result.Content, currentItem)
				continue
			}
			// Item exists in both current and newDefault — three-way merge the mapping.
			oldItem, existsInOld := oldByID[id]
			if !existsInOld {
				oldItem = &yaml.Node{Kind: yaml.MappingNode}
			}
			result.Content = append(result.Content, threeWayMerge(oldItem, newItem, currentItem, mode))
		}

		// Append items from newDefault that are truly new (not in old) and not already in current.
		currentIDs := make(map[string]bool, len(current.Content))
		for _, item := range current.Content {
			if id := mappingValue(item, identityKey); id != "" {
				currentIDs[id] = true
			}
		}
		for _, newItem := range newDefault.Content {
			id := mappingValue(newItem, identityKey)
			if id != "" && !currentIDs[id] && oldByID[id] == nil {
				result.Content = append(result.Content, newItem)
			}
		}

		return result
	}

	// No identity key found.
	// When all three sequences have the same length and all items are mappings,
	// match items by position and three-way merge each pair. This handles cases
	// where list items are modified but their count and order stay the same
	// (e.g., a single scalar field in a mapping item is changed).
	if len(oldDefault.Content) == len(newDefault.Content) && len(newDefault.Content) == len(current.Content) && allMappings(oldDefault, newDefault, current) {
		for i := range current.Content {
			result.Content = append(result.Content, threeWayMerge(oldDefault.Content[i], newDefault.Content[i], current.Content[i], mode))
		}
		return result
	}

	// Fall back to full-string set-based merge.
	oldSet := make(map[string]bool)
	for _, item := range oldDefault.Content {
		oldSet[nodeToString(item)] = true
	}

	currentMap := make(map[string]bool)
	for _, item := range current.Content {
		currentMap[nodeToString(item)] = true
	}

	newSet := make(map[string]bool)
	for _, item := range newDefault.Content {
		newSet[nodeToString(item)] = true
	}

	for _, currentItem := range current.Content {
		key := nodeToString(currentItem)
		if !oldSet[key] || newSet[key] {
			result.Content = append(result.Content, currentItem)
		}
	}

	for _, newItem := range newDefault.Content {
		key := nodeToString(newItem)
		if !oldSet[key] && !currentMap[key] {
			result.Content = append(result.Content, newItem)
		}
	}

	return result
}

// candidateIdentityKeys is the ordered list of field names tried when detecting the identity key for keyed-sequence merging.
var candidateIdentityKeys = []string{"name", "id", "key"}

// detectIdentityKey returns the name of a field that can serve as a unique identity key for the items of a sequence, or "" when no such key can be detected.
// A key is considered valid when every mapping item in at least one of the three sequences contains that field and all
// values for that field within the same sequence are unique (no duplicates within old/new/current).
func detectIdentityKey(oldDefault, newDefault, current *yaml.Node) string {
	// Only attempt identity-key detection when all items are mapping nodes.
	for _, seq := range []*yaml.Node{oldDefault, newDefault, current} {
		for _, item := range seq.Content {
			if item.Kind != yaml.MappingNode {
				return ""
			}
		}
	}

	for _, candidate := range candidateIdentityKeys {
		if isValidIdentityKey(candidate, oldDefault, newDefault, current) {
			return candidate
		}
	}
	return ""
}

// isValidIdentityKey reports whether candidate is a valid identity key across the three sequences.
func isValidIdentityKey(candidate string, seqs ...*yaml.Node) bool {
	found := false
	for _, seq := range seqs {
		if len(seq.Content) == 0 {
			continue
		}
		seen := make(map[string]bool, len(seq.Content))
		for _, item := range seq.Content {
			val := mappingValue(item, candidate)
			if val == "" {
				return false // not all items have the key
			}
			if seen[val] {
				return false // duplicate identity value within the same sequence
			}
			seen[val] = true
			found = true
		}
	}
	return found
}

// mappingValue returns the value of field key in a YAML mapping node, or "" when absent.
func mappingValue(node *yaml.Node, key string) string {
	if node.Kind != yaml.MappingNode {
		return ""
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1].Value
		}
	}
	return ""
}

// allMappings reports whether every item in each of the given sequences is a mapping node.
func allMappings(seqs ...*yaml.Node) bool {
	for _, seq := range seqs {
		for _, item := range seq.Content {
			if item.Kind != yaml.MappingNode {
				return false
			}
		}
	}
	return true
}
