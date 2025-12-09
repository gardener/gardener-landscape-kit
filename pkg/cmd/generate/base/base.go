// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/registry"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit generate base.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &options.Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:     "base (-c CONFIG_FILE) TARGET_DIR",
		Short:   "Generate or update the base directory",
		Example: "gardener-landscape-kit generate base -c ./example/20-componentconfig-glk.yaml ./base",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(args); err != nil {
				return err
			}

			if err := opts.Validate(); err != nil {
				return err
			}

			return run(cmd.Context(), opts)
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func run(_ context.Context, opts *options.Options) error {
	componentOpts := components.NewOptions(opts, afero.Afero{Fs: afero.NewOsFs()})

	reg := registry.New()
	registry.RegisterAllComponents(reg)

	return reg.GenerateBase(componentOpts)
}
