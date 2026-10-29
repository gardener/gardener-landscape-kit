// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/gardener/gardener-landscape-kit/pkg/registry"
)

const (
	beginPin = "## --- BEGIN generated component names (managed by hack/tools/example-components-generator) ---"
	endPin   = "## --- END generated component names ---"
)

func main() {
	filePath := flag.String("file", "example/20-componentconfig-glk.yaml", "Path to the example YAML file to update")
	flag.Parse()

	names, err := componentNames()
	if err != nil {
		fatalf("%v", err)
	}

	info, err := os.Stat(*filePath)
	if err != nil {
		fatalf("cannot stat file %q: %v", *filePath, err)
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		fatalf("cannot read file %q: %v", *filePath, err)
	}
	lines := strings.Split(string(data), "\n")

	beginIdx, endIdx, err := pinRange(lines)
	if err != nil {
		fatalf("%s: %v", *filePath, err)
	}

	// Rebuild the file: keep everything up to and including the BEGIN pin, replace the managed region with one
	// commented entry per component, then keep everything from the END pin onward.
	result := slices.Clone(lines[:beginIdx+1])
	for _, name := range names {
		result = append(result, "# - name: "+name)
	}
	result = append(result, lines[endIdx:]...)

	if err := os.WriteFile(*filePath, []byte(strings.Join(result, "\n")), info.Mode()); err != nil {
		fatalf("cannot write file %q: %v", *filePath, err)
	}
}

// componentNames returns the names of all registered components, in registry order.
func componentNames() ([]string, error) {
	names := make([]string, 0, len(registry.ComponentList))
	for _, newComponent := range registry.ComponentList {
		comp, err := newComponent()
		if err != nil {
			return nil, fmt.Errorf("failed to create component: %w", err)
		}
		names = append(names, comp.GetComponentMetadata().Name)
	}
	return names, nil
}

// pinRange locates the BEGIN and END pin lines that delimit the generated region and validates their ordering.
func pinRange(lines []string) (beginIdx, endIdx int, err error) {
	beginIdx = slices.Index(lines, beginPin)
	if beginIdx == -1 {
		return 0, 0, fmt.Errorf("BEGIN pin not found")
	}
	endIdx = slices.Index(lines, endPin)
	if endIdx == -1 {
		return 0, 0, fmt.Errorf("END pin not found")
	}
	if endIdx <= beginIdx {
		return 0, 0, fmt.Errorf("END pin (line %d) must come after BEGIN pin (line %d)", endIdx+1, beginIdx+1)
	}
	return beginIdx, endIdx, nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
