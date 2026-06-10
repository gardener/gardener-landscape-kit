// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"
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

	// ConfigFilePath is the path to the landscape kit configuration file.
	ConfigFilePath string

	// TargetDirPath is the target directory for generation.
	TargetDirPath string
	// Config is the path to the landscape kit configuration file.
	Config *configv1alpha1.LandscapeKitConfiguration
}

// Validate validates the options.
func (o *Options) Validate() error {
	if o.TargetDirPath == "" {
		return fmt.Errorf("target path is required")
	}

	if errs := configv1alpha1validation.ValidateLandscapeKitConfiguration(o.Config); len(errs) > 0 {
		return fmt.Errorf("invalid configuration: %v", errs.ToAggregate())
	}

	return nil
}

// Complete completes the options.
func (o *Options) Complete(args []string) error {
	if len(args) != 1 {
		return errors.New("requires exactly one argument")
	}
	o.TargetDirPath = args[0]

	data, err := os.ReadFile(o.ConfigFilePath) // #nosec G304 -- Trusted file from CLI argument.
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	o.Config = &configv1alpha1.LandscapeKitConfiguration{}
	if err = runtime.DecodeInto(configDecoder, data, o.Config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	return nil
}

// AddFlags adds flags for the options to the given FlagSet.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.ConfigFilePath, "config", "c", o.ConfigFilePath, "Path to configuration file.")
}

// WarnIfTargetNotRepoRoot logs a warning if TargetDirPath does not contain a `.git` directory.
// As of the `repositories:` config migration, generate commands expect the *repository root*, not the inner content directory.
// This catches users still passing the old-style inner path.
//
// TODO(LucaBernstein): remove a few releases after the `repositories:` config rollout.
func WarnIfTargetNotRepoRoot(targetDirPath string, fs afero.Afero, log logr.Logger) {
	gitDir := filepath.Join(targetDirPath, ".git")
	exists, err := fs.DirExists(gitDir)
	if err != nil || exists {
		return
	}
	log.Info(
		"WARNING: target directory does not look like a git repository root (no .git directory found). "+
			"As of the `repositories:` config migration, `generate base` and `generate landscape` expect the repository root, "+
			"not the inner content directory. The content sub-paths are now taken from `repositories.base.target` and "+
			"`repositories.landscape.target` in the config. If this directory is intentionally not a git repo yet, ignore this warning.",
		"targetDirPath", targetDirPath,
	)
}
