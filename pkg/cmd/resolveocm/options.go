// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package resolveocm

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	configv1alpha1validation "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1/validation"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
)

var configDecoder runtime.Decoder

func init() {
	configScheme := runtime.NewScheme()
	utilruntime.Must(configv1alpha1.AddToScheme(configScheme))
	configDecoder = serializer.NewCodecFactory(configScheme).UniversalDecoder()
}

// Options contains options for this command.
type Options struct {
	*cmd.Options

	configFilePath string

	// LandscapeDir is the directory containing all landscape specific configuration files.
	// It is used as output directory.
	LandscapeDir string

	// ConfigPath is the configuration file containing repositories and/or the root component.
	Config *configv1alpha1.OCMConfiguration
}

// Validate validates the options.
func (o *Options) validate() error {
	if o.LandscapeDir == "" {
		return fmt.Errorf("landscape dir is required")
	}

	if errs := configv1alpha1validation.ValidateOCMConfiguration(o.Config); len(errs) > 0 {
		return fmt.Errorf("invalid configuration: %v", errs.ToAggregate())
	}

	return nil
}

// Complete completes the options.
func (o *Options) complete() error {
	if o.configFilePath == "" {
		return fmt.Errorf("config option is required")
	}

	if err := o.loadConfigFile(o.configFilePath); err != nil {
		return fmt.Errorf("loading config file %s failed: %w", o.configFilePath, err)
	}

	return nil
}

func (o *Options) addFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.LandscapeDir, "landscape-dir", "l", "", "Path to a directory containing the landscape specific configuration files, aka overlays.")
	fs.StringVarP(&o.configFilePath, "config", "c", "", "Optional config file with repositories and other configuration.")
}

func (o *Options) effectiveOutputDir(subdir string) string {
	outputDir := path.Join(o.LandscapeDir, "ocm", o.Config.RootComponent.Name, o.Config.RootComponent.Version)
	if subdir != "" {
		outputDir = path.Join(outputDir, subdir)
	}
	return outputDir
}

func (o *Options) loadConfigFile(filename string) error {
	data, err := os.ReadFile(filename) // #nosec G304 -- Trusted file from CLI argument.
	if err != nil {
		return err
	}

	o.Config = &configv1alpha1.OCMConfiguration{}
	if err = runtime.DecodeInto(configDecoder, data, o.Config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	return nil
}
