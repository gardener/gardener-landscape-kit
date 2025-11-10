//go:build tools
// +build tools

// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// This package imports things required by build scripts, to force `go mod` to see them as dependencies
package hack

import (
	_ "github.com/ahmetb/gen-crd-api-reference-docs"
	_ "github.com/ironcore-dev/vgopath"
	_ "golang.org/x/tools/cmd/goimports"
	_ "k8s.io/code-generator"
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
)
