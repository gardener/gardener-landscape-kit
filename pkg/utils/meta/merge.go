// SPDX-FileCopyrightText: Contributors to the Gardener project
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

	oldNode, newNode, currentNode := documentContent(&oldDefault), documentContent(&newDefault), documentContent(&current)
	if newNode.Kind == 0 && oldNode.Kind == yaml.MappingNode && currentNode.Kind == yaml.MappingNode {
		// A removed mapping document still retains user-added entries.
		newNode = &yaml.Node{Kind: yaml.MappingNode, Tag: currentNode.Tag}
	}
	result, _ := mergeNode(oldNode, newNode, currentNode, mode)
	document := &yaml.Node{
		Kind:        yaml.DocumentNode,
		Content:     []*yaml.Node{result},
		HeadComment: current.HeadComment,
		LineComment: current.LineComment,
		FootComment: current.FootComment,
	}
	mergeNodeComments(&oldDefault, &newDefault, document)
	return EncodeResult(document)
}

// mergeNode merges compatible collections recursively and treats type changes as atomic replacements.
// The second return value reports a conflicting atomic replacement; collection conflicts are handled per entry.
func mergeNode(oldDefault, newDefault, current *yaml.Node, mode configv1alpha1.MergeMode) (*yaml.Node, bool) {
	oldDefault = documentContent(oldDefault)
	var result *yaml.Node
	var conflict bool
	switch {
	case canMergeCollection(yaml.MappingNode, oldDefault, newDefault, current):
		result = mergeMapping(oldDefault, newDefault, current, mode)
	case canMergeCollection(yaml.SequenceNode, oldDefault, newDefault, current):
		result = mergeSequence(oldDefault, newDefault, current, mode)
	default:
		defaultChanged := !nodesEqual(oldDefault, newDefault, false)
		userChanged := !nodesEqual(oldDefault, current, false)
		result = copyNode(current)
		if defaultChanged && !userChanged {
			result = copyNode(newDefault)
		}
		conflict = oldDefault.Kind != 0 && defaultChanged && userChanged && !nodesEqual(newDefault, current, false)
	}
	// Merge comments independently of which version supplied the value.
	result.HeadComment = current.HeadComment
	result.LineComment = current.LineComment
	result.FootComment = current.FootComment
	mergeNodeComments(oldDefault, newDefault, result)
	return result, conflict
}

func canMergeCollection(kind yaml.Kind, oldDefault, newDefault, current *yaml.Node) bool {
	return current.Kind == kind && newDefault.Kind == kind && current.Tag == newDefault.Tag &&
		(oldDefault.Kind == 0 || oldDefault.Kind == kind && oldDefault.Tag == newDefault.Tag)
}

// mergeMapping preserves surviving current keys in place and appends truly new default keys in default order.
func mergeMapping(oldDefault, newDefault, current *yaml.Node, mode configv1alpha1.MergeMode) *yaml.Node {
	oldEntries := mappingEntries(oldDefault)
	newEntries := mappingEntries(newDefault)
	currentEntries := mappingEntries(current)
	result := copyNode(current)
	result.Content = nil

	for i := 0; i < len(current.Content); i += 2 {
		currentEntry := mappingEntry{key: current.Content[i], value: current.Content[i+1]}
		key := currentEntry.key.Value
		oldEntry, existedInOld := oldEntries[key]
		newEntry, existsInNew := newEntries[key]
		if !existsInNew {
			if !existedInOld {
				result.Content = append(result.Content, currentEntry.key, currentEntry.value)
			}
			continue
		}
		keyNode, valueNode := mergeMappingEntry(oldEntry, newEntry, currentEntry, mode)
		result.Content = append(result.Content, keyNode, valueNode)
	}

	for i := 0; i < len(newDefault.Content); i += 2 {
		keyNode, valueNode := newDefault.Content[i], newDefault.Content[i+1]
		_, existsInCurrent := currentEntries[keyNode.Value]
		_, existedInOld := oldEntries[keyNode.Value]
		if !existsInCurrent && !existedInOld {
			result.Content = append(result.Content, keyNode, valueNode)
		}
	}
	return result
}

func mergeMappingEntry(oldDefault, newDefault, current mappingEntry, mode configv1alpha1.MergeMode) (*yaml.Node, *yaml.Node) {
	keyNode := copyNode(current.key)
	mergeNodeComments(documentContent(oldDefault.key), newDefault.key, keyNode)
	valueNode, conflict := mergeNode(oldDefault.value, newDefault.value, current.value, mode)
	if mode == configv1alpha1.MergeModeHint {
		switch {
		case conflict:
			annotateConflict(keyNode, valueNode, newDefault.value)
		case nodesEqual(newDefault.value, valueNode, false):
			stripGLKAnnotations(keyNode, valueNode)
		}
	}
	return keyNode, valueNode
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
	if oldComment == newComment || currentComment == newComment {
		// The default is unchanged or both versions converged — preserve current.
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

// mergeSequence selects identity, positional, or set matching. Modified sequences keep current item order;
// unchanged sequences follow the new default, while matched mapping items still preserve their key order.
func mergeSequence(oldDefault, newDefault, current *yaml.Node, mode configv1alpha1.MergeMode) *yaml.Node {
	result := copyNode(newDefault)
	identityKey := detectIdentityKey(oldDefault, newDefault, current)
	switch {
	case identityKey != "":
		result.Content = mergeKeyedSequence(oldDefault, newDefault, current, identityKey, mode)
	case len(oldDefault.Content) == len(newDefault.Content) && len(newDefault.Content) == len(current.Content) && allMappings(oldDefault, newDefault, current):
		result.Content = mergePositionalSequence(oldDefault, newDefault, current, mode)
	case nodesEqual(oldDefault, current, true):
		// Without a stable mapping-item match, an unchanged list can be replaced outright.
		result.Content = newDefault.Content
	default:
		result.Content = mergeSetSequence(oldDefault, newDefault, current)
	}
	return result
}

func sequenceByIdentity(sequence *yaml.Node, identityKey string) map[string]*yaml.Node {
	result := make(map[string]*yaml.Node, len(sequence.Content))
	for _, item := range sequence.Content {
		result[mappingValue(item, identityKey)] = item
	}
	return result
}

func mergeKeyedSequence(oldDefault, newDefault, current *yaml.Node, identityKey string, mode configv1alpha1.MergeMode) []*yaml.Node {
	oldByID := sequenceByIdentity(oldDefault, identityKey)
	newByID := sequenceByIdentity(newDefault, identityKey)
	currentByID := sequenceByIdentity(current, identityKey)
	unchanged := nodesEqual(oldDefault, current, true)
	items := current.Content
	if unchanged {
		items = newDefault.Content
	}
	var result []*yaml.Node
	for _, item := range items {
		id := mappingValue(item, identityKey)
		currentItem, newItem := currentByID[id], newByID[id]
		switch {
		case currentItem == nil:
			result = append(result, newItem)
		case newItem == nil:
			if oldByID[id] == nil {
				// Keep user additions, but drop items removed by the default.
				result = append(result, currentItem)
			}
		default:
			merged, _ := mergeNode(oldByID[id], newItem, currentItem, mode)
			result = append(result, merged)
		}
	}
	if !unchanged {
		for _, newItem := range newDefault.Content {
			id := mappingValue(newItem, identityKey)
			if currentByID[id] == nil && oldByID[id] == nil {
				// Add truly new default items, without restoring user-deleted items.
				result = append(result, newItem)
			}
		}
	}
	return result
}

func mergePositionalSequence(oldDefault, newDefault, current *yaml.Node, mode configv1alpha1.MergeMode) []*yaml.Node {
	result := make([]*yaml.Node, len(current.Content))
	for i := range current.Content {
		result[i], _ = mergeNode(oldDefault.Content[i], newDefault.Content[i], current.Content[i], mode)
	}
	return result
}

func sequenceSet(sequence *yaml.Node) map[string]bool {
	result := make(map[string]bool, len(sequence.Content))
	for _, item := range sequence.Content {
		result[nodeComparisonKey(item)] = true
	}
	return result
}

func mergeSetSequence(oldDefault, newDefault, current *yaml.Node) []*yaml.Node {
	oldSet := sequenceSet(oldDefault)
	newSet := sequenceSet(newDefault)
	currentSet := sequenceSet(current)
	var result []*yaml.Node

	for _, currentItem := range current.Content {
		key := nodeComparisonKey(currentItem)
		if !oldSet[key] || newSet[key] {
			result = append(result, currentItem)
		}
	}

	for _, newItem := range newDefault.Content {
		key := nodeComparisonKey(newItem)
		if !oldSet[key] && !currentSet[key] {
			result = append(result, newItem)
		}
	}

	return result
}

// candidateIdentityKeys is the ordered list of field names tried when detecting the identity key for keyed-sequence merging.
var candidateIdentityKeys = []string{"name", "id", "key"}

// detectIdentityKey returns the name of a field that can serve as a unique identity key for the items of a sequence, or "" when no such key can be detected.
// A key is considered valid when every mapping item in all non-empty sequences contains that field and all
// values for that field within the same sequence are unique (no duplicates within old/new/current).
func detectIdentityKey(oldDefault, newDefault, current *yaml.Node) string {
	// Only attempt identity-key detection when all items are mapping nodes.
	if !allMappings(oldDefault, newDefault, current) {
		return ""
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
