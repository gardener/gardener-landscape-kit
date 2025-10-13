// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package resolveocm

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit resolve-ocm-components.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:   "resolve-ocm-components",
		Short: "Collects all OCM components and their versions and generates component list and image vector files.",
		Long: "Collects all OCM components by walking all dependencies of the root component descriptor. " +
			"It outputs the component list and generates the imagevector overwrites for each component.",

		Example: `# Resolve all components starting at the root component. Writes component list, imagevector overwrite files for each component, and dumps all component descriptors.

gardner-landscape-kit resolve-ocm-components \
    --landscape-dir /path/to/landscape/dir \
    --config path/to/config-file
`,

		RunE: func(cmd *cobra.Command, _ []string) error {
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

func run(_ context.Context, opts *Options) error {
	opts.Log.Info("Starting resolve-ocm-components command", "outputDir", opts.effectiveOutputDir(""), "rootComponent", opts.Config.RootComponent)

	return ocm.ResolveOCMComponents(opts.Log, opts.Config, opts.effectiveOutputDir(""))
}
