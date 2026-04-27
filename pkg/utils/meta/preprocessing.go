// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode"
)

const (
	// commentLevelPrefix marks a comment with its position in the indentation stack.
	// PostProcess maps the stored level back to a column via the rebuilt stack.
	commentLevelPrefix = "###LEVEL="

	// commentSeqPeerPrefix marks a comment that precedes a sequence item ("- ").
	// PostProcess resolves it by peeking at the next content line's indentation,
	// so it is immune to indentation-stack drift caused by the YAML encoder.
	commentSeqPeerPrefix = "###SEQPEER#"
)

// PreProcess marks standalone comment lines with indentation metadata so that
// PostProcess can restore their original alignment after a YAML encode round-trip.
//
// Two marker types are used:
//   - ###SEQPEER#  — for comments immediately before a sequence item ("- ").
//     Resolved by matching the next content line's indentation.
//   - ###LEVEL=N   — for all other comments.
//     N is the comment's depth in the indentation stack.
func PreProcess(yamlContent []byte) []byte {
	return processLines(yamlContent, func(lines [][]byte, i, indent int, levels []int) []byte {
		line := lines[i]
		trimmed := line[indent:]
		if !bytes.HasPrefix(trimmed, []byte("#")) {
			return line
		}

		if nextContentIsSeqItem(lines, i) {
			return indentBytes(indent, commentSeqPeerPrefix, trimmed)
		}
		level := len(levels) - 1
		return indentBytes(indent, fmt.Sprintf("%s%d", commentLevelPrefix, level), trimmed)
	})
}

// PostProcess strips markers added by PreProcess and restores comment indentation.
func PostProcess(yamlContent []byte) []byte {
	return processLines(yamlContent, func(lines [][]byte, i, indent int, levels []int) []byte {
		line := lines[i]
		trimmed := line[indent:]

		switch {
		case bytes.HasPrefix(trimmed, []byte(commentSeqPeerPrefix)):
			comment := trimmed[len(commentSeqPeerPrefix):]
			if peerIndent := nextContentIndent(lines, i); peerIndent >= 0 {
				indent = peerIndent
			}
			return indentBytes(indent, "", comment)

		case bytes.HasPrefix(trimmed, []byte(commentLevelPrefix)):
			comment, level := parseLevelMarker(trimmed[len(commentLevelPrefix):])
			if level >= len(levels) {
				level = len(levels) - 1
			}
			col := 0
			if level >= 0 {
				col = levels[level]
			}
			return indentBytes(col, "", comment)

		default:
			return line
		}
	})
}

// lineCallback is called for each non-blank line. It may return a modified line.
type lineCallback func(lines [][]byte, i, indent int, levels []int) []byte

// processLines iterates YAML lines, maintains the indentation stack, and
// dispatches each non-blank line to fn. This is the shared loop that both
// PreProcess and PostProcess rely on — the stack-tracking logic is identical.
func processLines(yamlContent []byte, fn lineCallback) []byte {
	if len(yamlContent) == 0 {
		return yamlContent
	}

	lines := bytes.Split(yamlContent, []byte("\n"))
	var (
		oldIndent int
		levels    = []int{0}
	)

	for i, line := range lines {
		indent := lineIndent(line)
		if indent < 0 {
			continue // blank line
		}

		// Update the indentation stack. Must mirror the original algorithm exactly.
		if indent < oldIndent {
			for len(levels) > 0 && indent < oldIndent {
				oldIndent = levels[len(levels)-1]
				levels = levels[:len(levels)-1]
			}
			if len(levels) == 0 {
				levels = []int{0}
			} else {
				levels = append(levels, indent)
			}
		}
		if indent > oldIndent {
			levels = append(levels, indent)
			oldIndent = indent
		}

		lines[i] = fn(lines, i, indent, levels)
	}

	return bytes.Join(lines, []byte("\n"))
}

// lineIndent returns the column of the first non-space character, or -1 for blank lines.
func lineIndent(line []byte) int {
	return bytes.IndexFunc(line, func(r rune) bool { return !unicode.IsSpace(r) })
}

// indentBytes builds a line: <indent spaces> + prefix + content.
func indentBytes(indent int, prefix string, content []byte) []byte {
	out := make([]byte, 0, indent+len(prefix)+len(content))
	out = append(out, bytes.Repeat([]byte(" "), indent)...)
	out = append(out, prefix...)
	out = append(out, content...)
	return out
}

// nextContentIsSeqItem reports whether the next non-blank, non-comment line starts with "- ".
func nextContentIsSeqItem(lines [][]byte, after int) bool {
	for j := after + 1; j < len(lines); j++ {
		indent := lineIndent(lines[j])
		if indent < 0 {
			continue
		}
		trimmed := lines[j][indent:]
		if bytes.HasPrefix(trimmed, []byte("#")) {
			continue
		}
		return bytes.HasPrefix(trimmed, []byte("- "))
	}
	return false
}

// nextContentIndent returns the indentation of the next non-blank, non-marker line.
// Returns -1 if no such line exists.
func nextContentIndent(lines [][]byte, after int) int {
	for j := after + 1; j < len(lines); j++ {
		indent := lineIndent(lines[j])
		if indent < 0 {
			continue
		}
		trimmed := lines[j][indent:]
		if bytes.HasPrefix(trimmed, []byte(commentLevelPrefix)) || bytes.HasPrefix(trimmed, []byte(commentSeqPeerPrefix)) {
			continue
		}
		return indent
	}
	return -1
}

// parseLevelMarker extracts the level number and the original comment text from a LEVEL marker.
// Input is everything after the "###LEVEL=" prefix, e.g. "3# some comment".
func parseLevelMarker(after []byte) (comment []byte, level int) {
	idx := bytes.IndexByte(after, '#')
	if idx < 0 {
		return after, 0
	}
	n, err := strconv.Atoi(string(after[:idx]))
	if err != nil {
		return after, 0
	}
	return after[idx:], n
}
