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
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
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

	// Config is the path to the landscape kit configuration file.
	Config *configv1alpha1.LandscapeKitConfiguration
}

// Validate validates the options.
func (o *Options) validate() error {
	if o.LandscapeDir == "" {
		return fmt.Errorf("landscape dir is required")
	}

	if o.Config == nil || o.Config.OCM == nil {
		return fmt.Errorf("OCM configuration is required")
	}

	if errs := configv1alpha1validation.ValidateLandscapeKitConfiguration(o.Config); len(errs) > 0 {
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
	fs.StringVarP(&o.configFilePath, "config", "c", o.configFilePath, "Path to configuration file.")
}

func (o *Options) effectiveOutputDir() string {
	return path.Join(o.baseDir(), o.Config.OCM.RootComponent.Name, o.Config.OCM.RootComponent.Version)
}

func (o *Options) baseDir() string {
	return path.Join(o.LandscapeDir, files.GLKSystemDirName, "ocm")
}

func (o *Options) loadConfigFile(filename string) error {
	data, err := os.ReadFile(filename) // #nosec G304 -- Trusted file from CLI argument.
	if err != nil {
		return err
	}

	o.Config = &configv1alpha1.LandscapeKitConfiguration{}
	if err = runtime.DecodeInto(configDecoder, data, o.Config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	return nil
}
