// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package landscape

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/registry"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/kustomization"
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

	return nil
}

func run(_ context.Context, opts *options.Options) error {
	componentOpts := components.NewLandscapeOptions(opts, afero.Afero{Fs: afero.NewOsFs()})

	reg := registry.New()
	if err := registry.RegisterAllComponents(reg, opts.Config); err != nil {
		return fmt.Errorf("failed to register components: %w", err)
	}

	if err := reg.GenerateLandscape(componentOpts); err != nil {
		return fmt.Errorf("failed to generate landscape components: %w", err)
	}

	return kustomization.WriteLandscapeComponentsKustomizations(componentOpts)
}
