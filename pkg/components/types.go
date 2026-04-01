// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

const (
	// DirName is the directory name where components are stored.
	DirName = "components"
)

// Options is an interface for options passed to components for generating.
type Options interface {
	// GetComponentVector returns the component vector.
	GetComponentVector() utilscomponentvector.Interface
	// GetTargetPath returns the target directory path the components should be generated into.
	GetTargetPath() string
	// GetFilesystem returns the filesystem to use.
	GetFilesystem() afero.Afero
	// GetLogger returns the logger instance.
	GetLogger() logr.Logger
	// GetMergeMode returns the configured merge mode for three-way merges.
	GetMergeMode() configv1alpha1.MergeMode
}

// LandscapeOptions is an interface for options passed to components for generating the landscape.
type LandscapeOptions interface {
	Options

	// GetGitRepository returns the git repository information.
	GetGitRepository() *configv1alpha1.GitRepository
	// GetRelativeBasePath returns the base directory that is relative to the target path.
	GetRelativeBasePath() string
	// GetRelativeLandscapePath returns the landscape directory that is relative to the target path.
	GetRelativeLandscapePath() string
}

// Interface is the components interface that each component must implement.
type Interface interface {
	// Name returns the component name.
	Name() string
	// GenerateBase generates the component base dir.
	GenerateBase(Options) error
	// GenerateLandscape generates the component landscape dir.
	GenerateLandscape(LandscapeOptions) error
}

type options struct {
	componentVector utilscomponentvector.Interface
	targetPath      string
	filesystem      afero.Afero
	logger          logr.Logger
	mergeMode       configv1alpha1.MergeMode
}

// GetComponentVector returns the component vector.
func (o *options) GetComponentVector() utilscomponentvector.Interface {
	return o.componentVector
}

// GetTargetPath returns the target directory path the components should be generated into.
func (o *options) GetTargetPath() string {
	return o.targetPath
}

// GetFilesystem returns the filesystem to use.
func (o *options) GetFilesystem() afero.Afero {
	return o.filesystem
}

// GetLogger returns the logger instance.
func (o *options) GetLogger() logr.Logger {
	return o.logger
}

// GetMergeMode returns the configured merge mode for three-way merges.
func (o *options) GetMergeMode() configv1alpha1.MergeMode {
	return o.mergeMode
}

// NewOptions returns a new Options instance.
func NewOptions(opts *generateoptions.Options, fs afero.Afero) (Options, error) {
	var customComponentVectors [][]byte
	if opts.Config != nil && opts.Config.Git != nil {
		// isTargetLandscapeDir: when the target dir is the landscape dir, also check for a components.yaml in the base dir.
		isTargetLandscapeDir := path.Join(opts.TargetDirPath, files.CalculatePathToComponentBase(opts.Config.Git.Paths.Landscape), opts.Config.Git.Paths.Landscape) == path.Clean(opts.TargetDirPath)
		if isTargetLandscapeDir {
			baseCompVectorFile := path.Join(opts.TargetDirPath, files.CalculatePathToComponentBase(opts.Config.Git.Paths.Landscape), opts.Config.Git.Paths.Base, utilscomponentvector.ComponentVectorFilename)
			componentsBytes, err := readCustomComponentsFile(opts, fs, baseCompVectorFile)
			if err != nil {
				return nil, err
			}
			customComponentVectors = append(customComponentVectors, componentsBytes)
		}
	}

	componentsBytes, err := readCustomComponentsFile(opts, fs, path.Join(opts.TargetDirPath, utilscomponentvector.ComponentVectorFilename))
	if err != nil {
		return nil, err
	}
	customComponentVectors = append(customComponentVectors, componentsBytes)

	componentVector, err := utilscomponentvector.NewWithOverride(componentvector.DefaultComponentsYAML, customComponentVectors...)
	if err != nil {
		return nil, fmt.Errorf("failed to create component vector: %w", err)
	}

	return &options{
		componentVector: componentVector,
		targetPath:      path.Clean(opts.TargetDirPath),
		filesystem:      fs,
		logger:          opts.Log,
		mergeMode:       opts.Config.GetMergeMode(),
	}, nil
}

func readCustomComponentsFile(opts *generateoptions.Options, fs afero.Afero, filePath string) ([]byte, error) {
	customBytes, err := fs.ReadFile(filePath)
	if err == nil {
		opts.Log.Info("Found custom component vector override file", "file", filePath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to read component vector override file: %w", err)
	}
	return customBytes, nil
}

type landscapeOptions struct {
	Options

	gitRepository *configv1alpha1.GitRepository
}

// GetGitRepository returns the git repository information.
func (l *landscapeOptions) GetGitRepository() *configv1alpha1.GitRepository {
	return l.gitRepository
}

// GetRelativeBasePath returns the base directory that is relative to the target path.
func (l *landscapeOptions) GetRelativeBasePath() string {
	return l.gitRepository.Paths.Base
}

// GetRelativeLandscapePath returns the landscape directory that is relative to the target path.
func (l *landscapeOptions) GetRelativeLandscapePath() string {
	return l.gitRepository.Paths.Landscape
}

// NewLandscapeOptions returns a new LandscapeOptions instance.
func NewLandscapeOptions(opts *generateoptions.Options, fs afero.Afero) (LandscapeOptions, error) {
	options, err := NewOptions(opts, fs)
	if err != nil {
		return nil, err
	}

	return &landscapeOptions{
		Options:       options,
		gitRepository: opts.Config.Git,
	}, nil
}
