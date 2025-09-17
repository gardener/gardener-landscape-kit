// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/gardener/gardener-landscape-kit/cmd/gardener-landscape-kit/app"
)

func main() {
	if err := app.NewCommand().ExecuteContext(signals.SetupSignalHandler()); err != nil {
		os.Exit(1)
	}
}
