// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

var configDecoder runtime.Decoder

func init() {
	configScheme := runtime.NewScheme()
	utilruntime.Must(configv1alpha1.AddToScheme(configScheme))
	configDecoder = serializer.NewCodecFactory(configScheme).UniversalDecoder()
}

// Options contains options for the resolve ocm subcommand.
type Options struct {
	*cmd.Options

	fs afero.Afero

	// ConfigFilePath is the path to the GLK configuration file.
	ConfigFilePath string
	// TargetDirPath is the target directory for the resolved output files.
	TargetDirPath string
	// Config is the decoded GLK configuration.
	Config *configv1alpha1.LandscapeKitConfiguration

	// Debug enables additional debug output files like resources and image vectors.
	Debug bool
	// Workers is the number of concurrent workers to use for resolving OCM components.
	Workers int
	// IgnoreMissingComponents indicates whether to ignore missing components during resolution.
	IgnoreMissingComponents bool
}

// NewCommand creates a new cobra.Command for running gardener-landscape-kit resolve ocm.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:   "ocm (-c CONFIG_FILE) (-d TARGET_DIR)",
		Short: "Resolve OCM component descriptor dependencies and generate image vector overwrite files",
		Long: "Walk all dependencies of the root component descriptor (OCM), " +
			"produce the component list, and generate image vector overwrite files for each component.",
		Example: "gardener-landscape-kit resolve ocm -c ./example/20-componentconfig-glk.yaml -d ./",

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opts.complete(); err != nil {
				return err
			}

			if err := opts.validate(); err != nil {
				return err
			}

			return run(cmd.Context(), opts)
		},
	}

	opts.addFlags(cmd.Flags())

	return cmd
}

func (o *Options) complete() error {
	o.fs = afero.Afero{Fs: afero.NewOsFs()}

	if o.TargetDirPath == "" {
		return fmt.Errorf("target dir is required")
	}

	if o.ConfigFilePath == "" {
		return fmt.Errorf("config option is required")
	}

	if err := o.loadConfigFile(o.ConfigFilePath); err != nil {
		return fmt.Errorf("loading config file %s failed: %w", o.ConfigFilePath, err)
	}

	return nil
}

func (o *Options) validate() error {
	if o.Config == nil || o.Config.OCM == nil {
		return fmt.Errorf("OCM configuration is required")
	}

	return nil
}

func (o *Options) addFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.TargetDirPath, "target-dir", "d", "", "Path to a target directory containing the landscape specific configuration files.")
	fs.StringVarP(&o.ConfigFilePath, "config", "c", o.ConfigFilePath, "Path to configuration file.")
	fs.BoolVar(&o.Debug, "debug", false, "Enable debug output files like resources and imagevectors.")
	fs.IntVar(&o.Workers, "workers", 10, "Number of concurrent workers to use for resolving OCM components.")
	fs.BoolVar(&o.IgnoreMissingComponents, "ignore-missing-components", false, "Ignore missing components during resolution. By default, the command will fail if a component cannot be resolved or is not referenced.")
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

func run(_ context.Context, opts *Options) error {
	outputDir := opts.effectiveIntermediateOutputDir()
	opts.Log.Info("Starting resolve ocm command", "outputDir", outputDir, "rootComponent", opts.Config.OCM.RootComponent)

	if err := writeGitIgnoreFile(opts); err != nil {
		return err
	}
	return ocm.ResolveOCMComponents(opts.Log, opts.Config, opts.TargetDirPath, outputDir, opts.Workers, opts.Debug, opts.IgnoreMissingComponents)
}

func writeGitIgnoreFile(opts *Options) error {
	intermediateResultDir := opts.intermediateResultDir()
	if err := os.MkdirAll(intermediateResultDir, 0700); err != nil {
		return fmt.Errorf("failed to create intermediate result directory %s: %w", intermediateResultDir, err)
	}
	if err := os.WriteFile(path.Join(intermediateResultDir, ".gitignore"), []byte("/*"), 0600); err != nil {
		return fmt.Errorf("failed to write .gitignore file to intermediate result directory %s: %w", intermediateResultDir, err)
	}
	return nil
}
