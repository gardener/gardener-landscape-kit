// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"bytes"

	"go.yaml.in/yaml/v4"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

type mappingEntry struct {
	key, value *yaml.Node
}

// mappingEntries indexes both nodes of each entry. An absent or non-mapping node has no entries.
func mappingEntries(node *yaml.Node) map[string]mappingEntry {
	result := make(map[string]mappingEntry)
	if node == nil || node.Kind != yaml.MappingNode {
		return result
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		result[node.Content[i].Value] = mappingEntry{key: node.Content[i], value: node.Content[i+1]}
	}
	return result
}

// documentContent unwraps a document and represents absent content as an empty node.
func documentContent(node *yaml.Node) *yaml.Node {
	if node == nil {
		return &yaml.Node{}
	}
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			return &yaml.Node{}
		}
		return node.Content[0]
	}
	return node
}

// copyNode shallow-copies a node; callers rebuilding a collection replace Content separately.
func copyNode(node *yaml.Node) *yaml.Node {
	result := *node
	return &result
}

// nodesEqual compares node kinds, YAML tags, and values.
// Comparing comments also makes mapping order significant; otherwise mapping order is ignored.
func nodesEqual(a, b *yaml.Node, compareComments bool) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind || a.Tag != b.Tag {
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
			aMap := mappingEntries(a)
			bMap := mappingEntries(b)
			if len(aMap) != len(bMap) {
				return false
			}
			for key, aEntry := range aMap {
				bEntry, exists := bMap[key]
				if !exists || !nodesEqual(aEntry.key, bEntry.key, false) || !nodesEqual(aEntry.value, bEntry.value, false) {
					return false
				}
			}
		}
		return true
	}
	return true
}

// nodeComparisonKey compares scalars by type and value, and complex nodes by their serialized YAML.
func nodeComparisonKey(node *yaml.Node) string {
	if node.Kind == yaml.ScalarNode {
		return node.Tag + "\x00" + node.Value
	}
	// For non-scalar nodes, marshal to YAML for comparison
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	utilruntime.Must(encoder.Encode(node))
	utilruntime.Must(encoder.Close())
	return buf.String()
}

// EncodeResult encodes a YAML node to bytes
func EncodeResult(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	defer encoder.Close()
	encoder.SetIndent(2)
	encoder.CompactSeqIndent()
	if err := encoder.Encode(node); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
