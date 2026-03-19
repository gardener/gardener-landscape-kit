// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package landscape

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/registry"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/kustomization"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/version"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit generate landscape.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &options.Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:     "landscape (-c CONFIG_FILE) TARGET_DIR",
		Short:   "Generate or update landscape specific directories",
		Example: "gardener-landscape-kit generate landscape -c ./example/20-componentconfig-glk.yaml ./",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(args); err != nil {
				return err
			}

			// general config validation
			if err := opts.Validate(); err != nil {
				return err
			}
			// specific validation for landscape generation
			if err := validate(opts); err != nil {
				return err
			}

			return run(cmd.Context(), opts)
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func validate(opts *options.Options) error {
	if opts.Config.Git == nil {
		return fmt.Errorf("git config is required")
	}

	// Calculate the path to the base directory from the landscape directory
	relPathToRoot := files.RelativePathFromDirDepth(opts.Config.Git.Paths.Landscape)
	pathToBase := filepath.Join(opts.TargetDirPath, relPathToRoot, opts.Config.Git.Paths.Base)

	// Validate version compatibility
	fs := afero.Afero{Fs: afero.NewOsFs()}
	if err := version.ValidateLandscapeVersionCompatibility(pathToBase, fs); err != nil {
		return fmt.Errorf("version compatibility check failed: %w", err)
	}

	return nil
}

func run(_ context.Context, opts *options.Options) error {
	componentOpts, err := components.NewLandscapeOptions(opts, afero.Afero{Fs: afero.NewOsFs()})
	if err != nil {
		return fmt.Errorf("failed to create component options: %w", err)
	}

	reg := registry.New()
	if err := registry.RegisterAllComponents(reg, opts.Config); err != nil {
		return fmt.Errorf("failed to register components: %w", err)
	}

	if err := reg.GenerateLandscape(componentOpts); err != nil {
		return fmt.Errorf("failed to generate landscape components: %w", err)
	}

	return kustomization.WriteLandscapeComponentsKustomizations(componentOpts)
}
