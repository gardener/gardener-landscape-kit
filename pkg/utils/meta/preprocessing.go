// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"bytes"
)

// keepLeftAlignedMarker is a marker to identify left-aligned comment lines during pre- and post-processing.
const keepLeftAlignedMarker = "###KEEP_LEFT_ALIGNED###"

type lineProcessor func([]byte, *bytes.Buffer)

// process applies a line processor to each line of YAML content
func process(yamlContent []byte, processLine lineProcessor) []byte {
	if len(yamlContent) == 0 {
		return yamlContent
	}
	buf := bytes.Buffer{}
	lines := bytes.Split(yamlContent, []byte("\n"))
	for i, line := range lines {
		processLine(line, &buf)
		if i < len(lines)-1 {
			buf.Write([]byte("\n"))
		}
	}
	return buf.Bytes()
}

// preProcessLine adds a marker to a left aligned comment line.
// As the "go.yaml.in/yaml/v4" package does not store the original indentation of comments in the node model,
// they are indented during marshaling. This marker helps to identify such lines for left-alignment during post-processing.
func preProcessLine(line []byte, buf *bytes.Buffer) {
	buf.Write(line)
	if bytes.HasPrefix(line, []byte("#")) {
		buf.Write([]byte(keepLeftAlignedMarker))
	}
}

// postProcessLine removes the marker added during pre-processing and left-aligns the comment line again after marshaling.
func postProcessLine(line []byte, buf *bytes.Buffer) {
	if before, ok := bytes.CutSuffix(line, []byte(keepLeftAlignedMarker)); ok {
		line = before
		line = bytes.TrimLeft(line, " ")
	}
	buf.Write(line)
}

// preProcess applies preprocessing to YAML content
func preProcess(yamlContent []byte) []byte {
	return process(yamlContent, preProcessLine)
}

// postProcess applies postprocessing to YAML content
func postProcess(yamlContent []byte) []byte {
	return process(yamlContent, postProcessLine)
}
