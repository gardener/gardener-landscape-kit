// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package resolve

import (
	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/resolve/ocm"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/resolve/plain"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit resolve.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Resolve component versions and write the component vector file",
	}

	for _, subcommand := range []*cobra.Command{
		ocm.NewCommand(globalOpts),
		plain.NewCommand(globalOpts),
	} {
		cmd.AddCommand(subcommand)
	}

	return cmd
}
