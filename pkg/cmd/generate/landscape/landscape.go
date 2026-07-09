// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package landscape

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/registry"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/kustomization"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/version"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit generate landscape.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &options.Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:     "landscape (-c CONFIG_FILE) LANDSCAPE_REPO_ROOT",
		Short:   "Generate or update landscape specific directories",
		Example: "gardener-landscape-kit generate landscape -c ./example/20-componentconfig-glk.yaml ./",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(args); err != nil {
				return err
			}

			options.WarnIfTargetNotRepoRoot(opts.TargetDirPath, afero.Afero{Fs: afero.NewOsFs()}, opts.Log)

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
	if opts.Config.Repositories.Landscape == nil {
		return fmt.Errorf("repositories.landscape config is required")
	}
	if opts.Config.Repositories.Base == nil {
		return fmt.Errorf("repositories.base config is required for landscape generation")
	}
	if strings.TrimSpace(opts.Config.Repositories.Base.Target) == "" {
		return fmt.Errorf("repositories.base.target must be set for landscape generation")
	}

	landscape := opts.Config.Repositories.Landscape

	// opts.TargetDirPath is the on-disk landscape repository root.
	// landscape.BaseLink is the mount point of the base repository root inside the landscape repo;
	// base.target is joined to reach the actual base content.
	pathToBase := filepath.Join(opts.TargetDirPath, landscape.BaseLink, opts.Config.Repositories.Base.Target)

	// Validate version compatibility
	opts.Log.V(1).Info("Validating version compatibility", "pathToBase", pathToBase)
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

	if err := version.CheckGLKComponentVersion(componentOpts.GetComponentVector(), opts.Config, opts.Log); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
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
