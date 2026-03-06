// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"bytes"

	"go.yaml.in/yaml/v4"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

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

// findKeyNode finds the key node for a given key in a mapping (assumes node is a mapping)
func findKeyNode(node *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i]
		}
	}
	return &yaml.Node{Kind: yaml.ScalarNode, Value: key}
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
