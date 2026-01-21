// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"

	flag "github.com/spf13/pflag"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/meta"
)

func main() {
	inline := flag.BoolP("inline", "i", false, "inline the output")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  prettify [-i] <yaml-file>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	filename := flag.Args()[0]
	content, err := os.ReadFile(filename) // #nosec: G304 -- using any file is a feature
	if err != nil {
		log.Fatalf("Error reading file: %s", err)
	}
	prettified, err := meta.ThreeWayMergeManifest(nil, content, nil)
	if err != nil {
		log.Fatalf("Marshalling failed: %s", err)
	}

	if *inline {
		fileInfo, err := os.Stat(filename)
		if err != nil {
			log.Fatalf("Reading stats failed: %s", err)
		}
		err = os.WriteFile(filename, prettified, fileInfo.Mode())
		if err != nil {
			log.Fatalf("Overwriting %s failed: %s", filename, err)
		}
	} else {
		_, err = os.Stdout.Write(prettified)
		if err != nil {
			log.Fatalf("Writing failed: %s", err)
		}
	}
}
