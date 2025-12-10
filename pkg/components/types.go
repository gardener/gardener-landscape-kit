// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"github.com/go-logr/logr"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
)

const (
	// DirName is the directory name where components are stored.
	DirName = "components"
)

// Options is an interface for options passed to components for generating.
type Options interface {
	// GetTargetPath returns the target directory path the components should be generated into.
	GetTargetPath() string
	// GetFilesystem returns the filesystem to use.
	GetFilesystem() afero.Afero
	// GetLogger returns the logger instance.
	GetLogger() logr.Logger
}

// LandscapeOptions is an interface for options passed to components for generating the landscape.
type LandscapeOptions interface {
	Options

	// GetGitRepository returns the git repository information.
	GetGitRepository() *v1alpha1.GitRepository
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
	targetPath string
	filesystem afero.Afero
	logger     logr.Logger
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

// NewOptions returns a new Options instance.
func NewOptions(opts *generateoptions.Options, fs afero.Afero) Options {
	return &options{
		targetPath: opts.TargetDirPath,
		filesystem: fs,
		logger:     opts.Log,
	}
}

type landscapeOptions struct {
	Options

	gitRepository *v1alpha1.GitRepository
}

// GetGitRepository returns the git repository information.
func (l *landscapeOptions) GetGitRepository() *v1alpha1.GitRepository {
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
func NewLandscapeOptions(opts *generateoptions.Options, fs afero.Afero) LandscapeOptions {
	options := NewOptions(opts, fs)

	return &landscapeOptions{
		Options:       options,
		gitRepository: opts.Config.Git,
	}
}
