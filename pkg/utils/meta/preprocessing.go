// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"bytes"
	"strconv"
	"unicode"
)

const (
	// commentLevelPrefix marks a comment with its position in the indentation stack.
	// PostProcess maps the stored level back to a column via the rebuilt stack.
	commentLevelPrefix = "###LEVEL="
)

// PreProcess marks standalone comment lines with indentation metadata so that
// PostProcess can restore their original alignment after a YAML encode round-trip.
//
// Each comment is tagged with ###LEVEL=N where N is the comment's depth in the
// indentation stack. PostProcess maps N back to a column via the rebuilt stack.
func PreProcess(yamlContent []byte) []byte {
	return processLines(yamlContent, func(line, trimmed []byte, indent int, levels []int) []byte {
		if trimmed[0] != '#' {
			return line
		}
		return appendLevel(indent, len(levels)-1, trimmed)
	})
}

// PostProcess strips markers added by PreProcess and restores comment indentation.
func PostProcess(yamlContent []byte) []byte {
	return processLines(yamlContent, func(line, trimmed []byte, _ int, levels []int) []byte {
		if !bytes.HasPrefix(trimmed, []byte(commentLevelPrefix)) {
			return line
		}

		comment, level := parseLevelMarker(trimmed[len(commentLevelPrefix):])
		if level >= len(levels) {
			level = len(levels) - 1
		}
		if level < 0 {
			level = 0
		}
		return indentBytes(levels[level], comment)
	})
}

// processLines iterates YAML lines, maintains an indentation stack, and calls
// fn for each non-blank line. Comments do not permanently alter the stack:
// they compute their level from a temporary stack state, which is discarded.
func processLines(yamlContent []byte, fn func(line, trimmed []byte, indent int, levels []int) []byte) []byte {
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
			continue
		}
		trimmed := line[indent:]

		newLevels, newOld := pushPop(levels, oldIndent, indent)
		lines[i] = fn(line, trimmed, indent, newLevels)

		if trimmed[0] != '#' {
			levels = newLevels
			oldIndent = newOld
		}
	}

	return bytes.Join(lines, []byte("\n"))
}

// pushPop returns an updated indentation stack after incorporating indent.
// Always returns a new slice to avoid aliasing.
func pushPop(levels []int, oldIndent, indent int) ([]int, int) {
	out := make([]int, len(levels))
	copy(out, levels)

	if indent < oldIndent {
		for len(out) > 0 && indent < oldIndent {
			oldIndent = out[len(out)-1]
			out = out[:len(out)-1]
		}
		if len(out) == 0 {
			out = []int{0}
		} else {
			out = append(out, indent)
		}
	}
	if indent > oldIndent {
		out = append(out, indent)
		oldIndent = indent
	}
	return out, oldIndent
}

// lineIndent returns the column of the first non-space character, or -1 for blank lines.
func lineIndent(line []byte) int {
	return bytes.IndexFunc(line, func(r rune) bool { return !unicode.IsSpace(r) })
}

// indentBytes builds a line: <indent spaces> + content.
func indentBytes(indent int, content []byte) []byte {
	out := make([]byte, indent+len(content))
	for j := range indent {
		out[j] = ' '
	}
	copy(out[indent:], content)
	return out
}

// appendLevel builds a marker line: <indent spaces> + ###LEVEL=N + comment.
func appendLevel(indent, level int, comment []byte) []byte {
	prefix := []byte(commentLevelPrefix)
	levelStr := strconv.AppendInt(nil, int64(level), 10)
	out := make([]byte, indent+len(prefix)+len(levelStr)+len(comment))
	for j := range indent {
		out[j] = ' '
	}
	n := indent
	n += copy(out[n:], prefix)
	n += copy(out[n:], levelStr)
	copy(out[n:], comment)
	return out
}

// parseLevelMarker extracts the level number and the original comment text.
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
