// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/gardener/gardener-landscape-kit/pkg/registry"
)

func main() {
	templatePath := flag.String("template", "example/20-componentconfig-glk.yaml.tpl", "Path to the template file")
	filePath := flag.String("file", "example/20-componentconfig-glk.yaml", "Path to the output file")
	flag.Parse()

	names, err := componentNames()
	if err != nil {
		fatalf("%v", err)
	}

	tplData, err := os.ReadFile(*templatePath)
	if err != nil {
		fatalf("cannot read template %q: %v", *templatePath, err)
	}

	tpl, err := template.New("example").Parse(string(tplData))
	if err != nil {
		fatalf("cannot parse template %q: %v", *templatePath, err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, struct{ Components []string }{Components: names}); err != nil {
		fatalf("cannot execute template: %v", err)
	}

	info, err := os.Stat(*filePath)
	if err != nil && !os.IsNotExist(err) {
		fatalf("cannot stat file %q: %v", *filePath, err)
	}
	mode := os.FileMode(0644)
	if info != nil {
		mode = info.Mode()
	}

	if err := os.WriteFile(*filePath, buf.Bytes(), mode); err != nil {
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

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
