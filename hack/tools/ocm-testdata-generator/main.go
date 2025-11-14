// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"
	"path"
)

var configFilePath string

func init() {
	flag.StringVar(&configFilePath, "config", "", "Path to the config file")
}

func main() {
	flag.Parse()

	if configFilePath == "" {
		flag.Usage()
		os.Exit(1)
	}

	println("Starting OCM testdata generation...")
	config, err := LoadConfig(configFilePath)
	if err != nil {
		panic(err)
	}

	generator := NewGenerator(config)
	println("Generating testdata in", path.Dir(configFilePath))
	if err := generator.Generate(path.Dir(configFilePath)); err != nil {
		panic(err)
	}

	println("OCM testdata generation completed successfully.")
}
