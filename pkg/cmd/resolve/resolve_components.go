// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package resolve

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

// NewCommand creates a new cobra.Command for running gardener-landscape-kit resolve.
func NewCommand(globalOpts *cmd.Options) *cobra.Command {
	opts := &Options{Options: &options.Options{Options: globalOpts}}

	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Collect all components and their versions and generate component list and image vector files",
		Long: "Collect all components by walking all dependencies of the root component descriptor (OCM) " +
			"or retrieving image vector files. " +
			"Produce the component list and generate the image vector overwrites for each component.",

		Example: `# Resolve all components starting at the root component. Writes component list, imagevector overwrite files for each component, and dumps all component descriptors.

gardener-landscape-kit resolve \
    --ocm \
    --target-dir /path/to/landscape/dir \
    --config path/to/config-file

# Create or update a default components list.

gardener-landscape-kit resolve \
    --target-dir /path/to/base/dir \
    --config path/to/config-file
`,

		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opts.complete(); err != nil {
				return err
			}

			if err := opts.validate(); err != nil {
				return err
			}

			switch {
			case opts.OCM:
				return runOCM(cmd.Context(), opts)
			default:
				return runPlain(cmd.Context(), opts)
			}
		},
	}

	opts.addFlags(cmd.Flags())

	return cmd
}

func runPlain(_ context.Context, opts *Options) error {
	var customBytes []byte
	if opts.Config != nil && opts.Config.VersionConfig != nil {
		if updateStrategy := opts.Config.VersionConfig.DefaultVersionsUpdateStrategy; updateStrategy != nil && *updateStrategy == v1alpha1.DefaultVersionsUpdateStrategyReleaseBranch {
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

	if err := utilscomponentvector.WriteComponentVectorFile(opts.fs, opts.TargetDirPath, componentVector); err != nil {
		return err
	}
	return nil
}

func runOCM(_ context.Context, opts *Options) error {
	outputDir := opts.effectiveIntermediateOutputDir()
	opts.Log.Info("Starting resolve command for ocm components", "outputDir", outputDir, "rootComponent", opts.Config.OCM.RootComponent)

	if err := writeGitIgnoreFile(opts); err != nil {
		return err
	}
	return ocm.ResolveOCMComponents(opts.Log, opts.Config, opts.TargetDirPath, outputDir, opts.Workers, opts.Debug)
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
