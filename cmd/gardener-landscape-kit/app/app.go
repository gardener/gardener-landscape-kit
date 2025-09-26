// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/component-base/version/verflag"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate"
)

// Name is a const for the name of this component.
const Name = "gardener-landscape-kit"

// NewCommand creates a new cobra.Command for running gardener-landscape-kit.
func NewCommand() *cobra.Command {
	opts := &cmd.Options{
		IOStreams: genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
	}

	cmd := &cobra.Command{
		Use:   Name,
		Short: Name + " generates and manages manifests for Gardener landscapes.",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if err := opts.Validate(); err != nil {
				return err
			}

			if err := opts.Complete(); err != nil {
				return err
			}

			return nil
		},
	}

	// don't output usage on further errors raised during execution
	cmd.SilenceUsage = true

	for _, subcommand := range []*cobra.Command{
		generate.NewCommand(opts),
	} {
		cmd.AddCommand(subcommand)
	}

	flags := cmd.PersistentFlags()
	opts.AddFlags(flags)
	verflag.AddFlags(flags)

	return cmd
}
