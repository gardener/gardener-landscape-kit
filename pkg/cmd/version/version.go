// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"fmt"

	"github.com/spf13/cobra"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/component-base/version"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
)

// NewCommand creates a new cobra.Command.
func NewCommand(opts *cmd.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the program version information",
		Long:  "Print the program version information",

		Run: func(_ *cobra.Command, _ []string) {
			_, err := fmt.Fprintf(opts.Out, "gardener-landscape-kit Version: %s\n", version.Get())
			utilruntime.Must(err)
		},
	}

	return cmd
}
