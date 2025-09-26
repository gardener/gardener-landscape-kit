// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
)

// Options contains options for this command.
type Options struct {
	*cmd.Options

	// BaseDir is the base directory containing the landscape configuration files.
	BaseDir string
	// LandscapeDir is the directory containing all landscape specific configuration files.
	LandscapeDir string
}

// Validate validates the options.
func (o *Options) validate() error {
	if o.BaseDir == "" {
		return fmt.Errorf("base dir is required")
	}
	return nil
}

// Complete completes the options.
func (o *Options) complete() error {
	return nil
}

func (o *Options) addFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.BaseDir, "base-dir", "b", "", "Path to a directory containing the landscape base configuration files.")
	fs.StringVarP(&o.LandscapeDir, "landscape-dir", "l", "", "Path to a directory containing the landscape specific configuration files, aka overlays.")
}
