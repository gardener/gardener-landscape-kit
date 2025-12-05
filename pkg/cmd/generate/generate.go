// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/base"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/landscape"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit generate.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generates or updates the landscape directories",
	}

	for _, subcommand := range []*cobra.Command{
		base.NewCommand(globalOpts),
		landscape.NewCommand(globalOpts),
	} {
		cmd.AddCommand(subcommand)
	}

	return cmd
}
