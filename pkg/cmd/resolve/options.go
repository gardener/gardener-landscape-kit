// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package resolve

import (
	"fmt"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	configv1alpha1validation "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1/validation"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
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
	*options.Options

	fs afero.Afero

	// OCM indicates whether to resolve OCM components based on the root component descriptor or only generate the (default) component list.
	OCM bool

	// Debug enables additional debug output files like resources and image vectors.
	Debug bool

	// Workers is the number of concurrent workers to use for resolving OCM components.
	Workers int
}

// Validate validates the options.
func (o *Options) validate() error {
	if o.TargetDirPath == "" {
		return fmt.Errorf("target dir is required")
	}

	if o.OCM && (o.Config == nil || o.Config.OCM == nil) {
		return fmt.Errorf("OCM configuration is required")
	}

	if errs := configv1alpha1validation.ValidateLandscapeKitConfiguration(o.Config); len(errs) > 0 {
		return fmt.Errorf("invalid configuration: %v", errs.ToAggregate())
	}

	return nil
}

// Complete completes the options.
func (o *Options) complete() error {
	o.fs = afero.Afero{Fs: afero.NewOsFs()}

	if o.ConfigFilePath == "" {
		return fmt.Errorf("config option is required")
	}

	if err := o.loadConfigFile(o.ConfigFilePath); err != nil {
		return fmt.Errorf("loading config file %s failed: %w", o.ConfigFilePath, err)
	}

	return nil
}

func (o *Options) addFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.TargetDirPath, "target-dir", "d", "", "Path to a target directory containing the base or landscape specific configuration files.")
	fs.StringVarP(&o.ConfigFilePath, "config", "c", o.ConfigFilePath, "Path to configuration file.")
	fs.BoolVar(&o.Debug, "debug", false, "Enable debug output files like resources and imagevectors.")
	fs.IntVar(&o.Workers, "workers", 10, "Number of concurrent workers to use for resolving OCM components.")
	fs.BoolVar(&o.OCM, "ocm", false, "Whether to resolve OCM components based on the root component descriptor or only generate the (default) component list.")
}

func (o *Options) effectiveIntermediateOutputDir() string {
	return path.Join(o.intermediateResultDir(), o.Config.OCM.RootComponent.Name, o.Config.OCM.RootComponent.Version)
}

func (o *Options) intermediateResultDir() string {
	return path.Join(o.TargetDirPath, files.GLKSystemDirName, "ocm")
}

func (o *Options) loadConfigFile(filename string) error {
	data, err := o.fs.ReadFile(filename) // #nosec G304 -- Trusted file from CLI argument.
	if err != nil {
		return err
	}

	o.Config = &configv1alpha1.LandscapeKitConfiguration{}
	if err = runtime.DecodeInto(configDecoder, data, o.Config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	return nil
}
