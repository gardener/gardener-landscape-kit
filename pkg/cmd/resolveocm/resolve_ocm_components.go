// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package resolveocm

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit resolve-ocm-components.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:   "resolve-ocm-components",
		Short: "Collect all OCM components and their versions and generate component list and image vector files",
		Long: "Collect all OCM components by walking all dependencies of the root component descriptor. " +
			"Produce the component list and generate the image vector overwrites for each component.",

		Example: `# Resolve all components starting at the root component. Writes component list, imagevector overwrite files for each component, and dumps all component descriptors.

gardner-landscape-kit resolve-ocm-components \
    --landscape-dir /path/to/landscape/dir \
    --config path/to/config-file
`,

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opts.complete(); err != nil {
				return err
			}

			if err := opts.validate(); err != nil {
				return err
			}

			return run(cmd.Context(), opts)
		},
	}

	opts.addFlags(cmd.Flags())

	return cmd
}

func run(_ context.Context, opts *Options) error {
	outputDir := opts.effectiveOutputDir()
	opts.Log.Info("Starting resolve-ocm-components command", "outputDir", outputDir, "rootComponent", opts.Config.OCM.RootComponent)

	if err := writeGitIgnoreFile(opts); err != nil {
		return err
	}

	return ocm.ResolveOCMComponents(opts.Log, opts.Config, opts.LandscapeDir, outputDir, opts.Workers, opts.Debug)
}

func writeGitIgnoreFile(opts *Options) error {
	baseDir := opts.baseDir()
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return fmt.Errorf("failed to create output base directory %s: %w", baseDir, err)
	}
	if err := os.WriteFile(path.Join(baseDir, ".gitignore"), []byte("/*"), 0600); err != nil {
		return fmt.Errorf("failed to write .gitignore file to output base directory %s: %w", baseDir, err)
	}
	return nil
}
