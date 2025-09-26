// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit generate.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generates or updates the landscape directories",
		Long:  "Generates or updates the base or landscape specific directories.",

		Example: `# Generate the landscape base directory
gardner-landscape-kit generate --base-dir /path/to/base/dir

# Generate the landscape directory
gardner-landscape-kit generate --base-dir /path/to/base/dir --landscape-dir /path/to/landscape/dir
`,

		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.validate(); err != nil {
				return err
			}

			if err := opts.complete(); err != nil {
				return err
			}

			return run(cmd.Context(), opts)
		},
	}

	opts.addFlags(cmd.Flags())

	return cmd
}

func run(ctx context.Context, opts *Options) error {
	if opts.LandscapeDir != "" {
		fmt.Println("Generating landscape directory...")
	} else {
		fmt.Println("Generating base directory...")
	}

	return nil
}
