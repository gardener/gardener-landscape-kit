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

// commentLevelPrefix is a prefix used to mark comment lines with their indentation level during pre-processing.
const commentLevelPrefix = "###LEVEL="

// PreProcess applies preprocessing to YAML content to mark custom indented comments for re-alignment in post-processing
func PreProcess(yamlContent []byte) []byte {
	return process(yamlContent, preprocessLine)
}

// PostProcess removes prefixes added in `PreProcess` and restores the original indentation of the comments.
func PostProcess(yamlContent []byte) []byte {
	return process(yamlContent, postprocessLine)
}

// process is a helper function that processes YAML content line by line using the provided processing function.
// It also tracks indentation levels to handle nested comments correctly.
// The processing function can modify the line based on its content and indentation level.
func process(yamlContent []byte, processLineFunc func(line []byte, index int, levels []int) []byte) []byte {
	if len(yamlContent) == 0 {
		return yamlContent
	}
	buf := bytes.Buffer{}
	var oldIndent int
	lines := bytes.Split(yamlContent, []byte("\n"))
	var levels = []int{0}
	for i, line := range lines {
		index := bytes.IndexFunc(line, func(r rune) bool {
			return !unicode.IsSpace(r)
		})
		if index == -1 {
			buf.Write(line)
		} else {
			if index > oldIndent {
				levels = append(levels, index)
				oldIndent = index
			} else if index < oldIndent {
				for len(levels) > 0 && index < oldIndent {
					oldIndent = levels[len(levels)-1]
					levels = levels[:len(levels)-1]
				}
				if len(levels) == 0 {
					levels = []int{0}
				} else {
					levels = append(levels, index)
				}
			}
			buf.Write(processLineFunc(line, index, levels))
		}
		if i < len(lines)-1 {
			buf.Write([]byte("\n"))
		}
	}
	return buf.Bytes()
}

func preprocessLine(line []byte, index int, levels []int) []byte {
	trimmedLine := line[index:]
	if !bytes.HasPrefix(trimmedLine, []byte("#")) {
		return line
	}
	level := len(levels) - 1
	return append([]byte(fmt.Sprintf("%s%s%d", line[:index], commentLevelPrefix, level)), trimmedLine...)
}

func postprocessLine(line []byte, index int, levels []int) []byte {
	if !bytes.HasPrefix(line[index:], []byte(commentLevelPrefix)) {
		return line
	}

	trimmedLine, level := trimLevel(line[index+len(commentLevelPrefix):])
	if level >= len(levels) {
		level = len(levels) - 1
	}
	indent := 0
	if level >= 0 {
		indent = levels[level]
	}
	line = append(bytes.Repeat([]byte(" "), indent), trimmedLine...)
	return line
}

func trimLevel(line []byte) ([]byte, int) {
	idx := bytes.IndexByte(line, '#')
	if idx == -1 {
		return line, 0
	}
	level, err := strconv.Atoi(string(line[:idx]))
	if err != nil {
		return line, 0
	}
	return line[idx:], level
}
