// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package plain

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

var configDecoder runtime.Decoder

func init() {
	configScheme := runtime.NewScheme()
	utilruntime.Must(configv1alpha1.AddToScheme(configScheme))
	configDecoder = serializer.NewCodecFactory(configScheme).UniversalDecoder()
}

// Options contains options for the resolve plain subcommand.
type Options struct {
	*cmd.Options

	fs afero.Afero

	// ConfigFilePath is the path to the GLK configuration file.
	ConfigFilePath string
	// TargetDirPath is the target directory where the component vector file will be written.
	TargetDirPath string
	// Config is the decoded GLK configuration.
	Config *configv1alpha1.LandscapeKitConfiguration
}

// NewCommand creates a new cobra.Command for running gardener-landscape-kit resolve plain.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &Options{Options: globalOpts}

	cmd := &cobra.Command{
		Use:   "plain (-c CONFIG_FILE) (-d TARGET_DIR)",
		Short: "Write the default component vector file to the target directory",
		Long: "Write the default component vector file (components.yaml) to TARGET_DIR, " +
			"applying any user overrides from an existing components.yaml in the same directory. " +
			"Version pins in the existing file are preserved across runs via three-way merge.",
		Example: "gardener-landscape-kit resolve plain -c ./example/20-componentconfig-glk.yaml -d ./base",

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opts.complete(); err != nil {
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

func (o *Options) addFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.TargetDirPath, "target-dir", "d", "", "Path to a target directory where the component vector file will be written.")
	fs.StringVarP(&o.ConfigFilePath, "config", "c", o.ConfigFilePath, "Path to configuration file.")
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
	if opts.Config != nil && opts.Config.VersionConfig != nil {
		if updateStrategy := opts.Config.VersionConfig.DefaultVersionsUpdateStrategy; updateStrategy != nil && *updateStrategy == configv1alpha1.DefaultVersionsUpdateStrategyReleaseBranch {
			opts.Log.Info("Updating default component vector file from the release branch", "branch", utilscomponentvector.GetReleaseBranchName())
			var err error
			// The componentvector.DefaultComponentsYAML is intentionally overridden, so that subsequently it can be used to extract the updated default component vector versions.
			componentvector.DefaultComponentsYAML, err = utilscomponentvector.GetDefaultComponentVectorFromGitHub()
			if err != nil {
				return fmt.Errorf("failed to update default component vector file: %w", err)
			}
		}
	}

	compVectorFile := path.Join(opts.TargetDirPath, utilscomponentvector.ComponentVectorFilename)
	opts.Log.Info("Writing component vector file", "file", compVectorFile)

	var customBytes []byte
	var err error
	if customBytes, err = opts.fs.ReadFile(compVectorFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to read component vector override file: %w", err)
		}
	}

	componentVector, err := utilscomponentvector.NewWithOverride(componentvector.DefaultComponentsYAML, customBytes)
	if err != nil {
		return fmt.Errorf("failed to create component vector: %w", err)
	}

	return utilscomponentvector.WriteComponentVectorFile(opts.fs, opts.TargetDirPath, componentVector)
}
