// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package resolveocm

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/config"
)

// Options contains options for this command.
type Options struct {
	*cmd.Options

	// LandscapeDir is the directory containing all landscape specific configuration files.
	// It is used as output directory.
	LandscapeDir string
	// ConfigPath is a path to a configuration file containing repositories and/or the root component
	ConfigPath string

	Config *config.Config
}

// Validate validates the options.
func (o *Options) validate() error {
	if o.LandscapeDir == "" {
		return fmt.Errorf("landscape dir is required")
	}

	if o.ConfigPath == "" {
		return fmt.Errorf("config option is required")
	}

	var err error
	o.Config, err = loadConfigFile(o.ConfigPath)
	if err != nil {
		return fmt.Errorf("loading config file %s failed: %w", o.ConfigPath, err)
	}

	if o.Config.RootComponent.Name == "" {
		return fmt.Errorf("root component name is required in config file")
	}
	if len(strings.Split(o.Config.RootComponent.Name, "/")) == 1 {
		return fmt.Errorf("root component name must be qualified (format 'example.com/my-org/my-root-component:1.23.4')")
	}
	if o.Config.RootComponent.Version == "" {
		return fmt.Errorf("root component version is required in config file")
	}

	if len(o.Config.Repositories) == 0 {
		return fmt.Errorf("at least one OCI repository must be specified in config file")
	}

	return nil
}

// Complete completes the options.
func (o *Options) complete() error {
	return nil
}

func (o *Options) addFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.LandscapeDir, "landscape-dir", "l", "", "Path to a directory containing the landscape specific configuration files, aka overlays.")
	fs.StringVar(&o.ConfigPath, "config", "", "Optional config file with repositories and other configuration.")
}

func (o *Options) effectiveOutputDir(subdir string) string {
	outputDir := path.Join(o.LandscapeDir, "ocm", o.Config.RootComponent.Name, o.Config.RootComponent.Version)
	if subdir != "" {
		outputDir = path.Join(outputDir, subdir)
	}
	return outputDir
}

func loadConfigFile(filename string) (*config.Config, error) {
	data, err := os.ReadFile(filename) // #nosec G304 -- Trusted file from CLI argument.
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
