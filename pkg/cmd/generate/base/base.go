// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package base

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/registry"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/version"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit generate base.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &options.Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:     "base (-c CONFIG_FILE) BASE_REPO_ROOT",
		Short:   "Generate or update the base directory",
		Example: "gardener-landscape-kit generate base -c ./example/20-componentconfig-glk.yaml ./base",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(args); err != nil {
				return err
			}

			options.WarnIfTargetNotRepoRoot(opts.TargetDirPath, afero.Afero{Fs: afero.NewOsFs()}, opts.Log)

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
	fs := afero.Afero{Fs: afero.NewOsFs()}
	componentOpts, err := components.NewOptions(opts, fs)
	if err != nil {
		return fmt.Errorf("failed to create component options: %w", err)
	}

	currentComponentVector, err := utilscomponentvector.ReadComponentVectorMetadata(opts.TargetDirPath, fs)
	if err != nil {
		return fmt.Errorf("failed to read current component vector metadata: %w", err)
	}

	reg := registry.New(currentComponentVector, componentOpts.GetComponentVector())
	if err := registry.RegisterAllComponents(opts.Log, reg, opts.Config); err != nil {
		return fmt.Errorf("failed to register components: %w", err)
	}

	componentVersion, _ := componentOpts.GetComponentVector().FindComponentVersion(componentvector.NameGardenerGardenerLandscapeKit)
	if err := version.CheckGLKComponentVersion(componentVersion, opts.Config, opts.Log); err != nil {
		return fmt.Errorf("version check failed: %w", err)
	}

	if err := reg.GenerateBase(componentOpts); err != nil {
		return err
	}

	// Write version metadata after successful generation,
	// alongside the generated base content (TargetDirPath joined with base.Target).
	if err := version.WriteVersionMetadata(componentOpts.GetTargetPath(), fs); err != nil {
		return fmt.Errorf("failed to write version metadata: %w", err)
	}

	if err := utilscomponentvector.WriteComponentVectorMetadata(componentOpts.GetComponentVector(), componentOpts.GetTargetPath(), fs); err != nil {
		return fmt.Errorf("failed to write component vector metadata: %w", err)
	}

	return nil
}
