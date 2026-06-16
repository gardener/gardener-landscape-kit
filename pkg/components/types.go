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
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

const (
	// DirName is the directory name where components are stored.
	DirName = "components"
)

// Options is an interface for options passed to components for generating.
type Options interface {
	// GetComponentVector returns the component vector.
	GetComponentVector() utilscomponentvector.Interface
	// GetRepoRoot returns the path on disk to the root of the repository being generated into
	// (the value the user passed as TARGET_DIR).
	GetRepoRoot() string
	// GetTargetPath returns the path the component should write its content into.
	// This is the repository root joined with the repository-relative target (base.target or landscape.target).
	GetTargetPath() string
	// GetFilesystem returns the filesystem to use.
	GetFilesystem() afero.Afero
	// GetLogger returns the logger instance.
	GetLogger() logr.Logger
	// GetMergeMode returns the configured mode to solve merge conflicts.
	GetMergeMode() configv1alpha1.MergeMode
}

// LandscapeOptions is an interface for options passed to components for generating the landscape.
type LandscapeOptions interface {
	Options

	// GetLandscapeURL returns the URL of the landscape git repository.
	GetLandscapeURL() string
	// GetLandscapeRef returns the git reference of the landscape repository.
	GetLandscapeRef() configv1alpha1.GitRepositoryRef
	// GetRelativeBasePath returns the path inside the landscape repository
	// where the base repository's generated content lives, i.e., landscape.baseLink.
	GetRelativeBasePath() string
	// GetRelativeLandscapePath returns landscape.target — i.e. the
	// landscape directory within the landscape repository.
	GetRelativeLandscapePath() string
	// GetRelativeBaseComponentPath returns the path from a landscape
	// component directory to the corresponding base component directory,
	// suitable for kustomize "resources:" entries. componentDir is the
	// component-specific path beneath DirName (e.g. "gardener/garden").
	GetRelativeBaseComponentPath(componentDir string) string
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
	repoRoot        string
	targetPath      string
	filesystem      afero.Afero
	logger          logr.Logger
	mergeMode       configv1alpha1.MergeMode
}

// GetComponentVector returns the component vector.
func (o *options) GetComponentVector() utilscomponentvector.Interface {
	return o.componentVector
}

// GetRepoRoot returns the on-disk path to the root of the repository being generated into.
func (o *options) GetRepoRoot() string {
	return o.repoRoot
}

// GetTargetPath returns the path the component should write its content into.
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

// NewOptions returns a new Options instance for `glk generate base`.
//
// opts.TargetDirPath is treated as the on-disk root of the base repository being generated into.
// The component target directory is repoRoot/<base.target>, which is also where the base components.yaml override is read from.
// For landscape generation, NewLandscapeOptions overrides this with landscape-specific paths.
func NewOptions(opts *generateoptions.Options, fs afero.Afero) (Options, error) {
	repoRoot := path.Clean(opts.TargetDirPath)
	targetPath := path.Join(repoRoot, opts.Config.Repositories.Base.Target)

	componentVector, err := loadComponentVector(opts, fs, path.Join(targetPath, utilscomponentvector.ComponentVectorFilename))
	if err != nil {
		return nil, err
	}
	return newOptions(opts, fs, repoRoot, targetPath, componentVector), nil
}

func newOptions(opts *generateoptions.Options, fs afero.Afero, repoRoot, targetPath string, componentVector utilscomponentvector.Interface) *options {
	return &options{
		componentVector: componentVector,
		repoRoot:        repoRoot,
		targetPath:      targetPath,
		filesystem:      fs,
		logger:          opts.Log,
		mergeMode:       *opts.Config.MergeMode,
	}
}

// loadComponentVector reads zero or more components.yaml override files (later paths override earlier ones) on top of the default component vector embedded in the binary.
func loadComponentVector(opts *generateoptions.Options, fs afero.Afero, overridePaths ...string) (utilscomponentvector.Interface, error) {
	var customComponentVectors [][]byte
	for _, p := range overridePaths {
		componentsBytes, err := readCustomComponentsFile(opts, fs, p)
		if err != nil {
			return nil, err
		}
		customComponentVectors = append(customComponentVectors, componentsBytes)
	}
	componentVector, err := utilscomponentvector.NewWithOverride(componentvector.DefaultComponentsYAML, customComponentVectors...)
	if err != nil {
		return nil, fmt.Errorf("failed to create component vector: %w", err)
	}
	return componentVector, nil
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

	landscape  *configv1alpha1.LandscapeRepositoryConfig
	targetPath string
}

// GetTargetPath overrides Options.GetTargetPath: for landscape generation the
// content directory is the landscape repository root joined with landscape.target.
func (l *landscapeOptions) GetTargetPath() string {
	return l.targetPath
}

// GetLandscapeURL returns the URL of the landscape git repository.
func (l *landscapeOptions) GetLandscapeURL() string {
	return l.landscape.URL
}

// GetLandscapeRef returns the git reference of the landscape repository.
func (l *landscapeOptions) GetLandscapeRef() configv1alpha1.GitRepositoryRef {
	return l.landscape.Ref
}

// GetRelativeBasePath returns the path inside the landscape repository where the base repository's generated content lives.
func (l *landscapeOptions) GetRelativeBasePath() string {
	return l.landscape.BaseLink
}

// GetRelativeLandscapePath returns landscape.target
// This is the landscape directory within the landscape repository.
func (l *landscapeOptions) GetRelativeLandscapePath() string {
	return l.landscape.Target
}

// GetRelativeBaseComponentPath returns the path from a landscape component
// directory to the corresponding base component directory, suitable for kustomize "resources:" entries.
// Both endpoints are relative to the landscape repository root:
// the landscape side at landscape.target/components/<dir>,
// the base side at landscape.baseLink/components/<dir>.
func (l *landscapeOptions) GetRelativeBaseComponentPath(componentDir string) string {
	// The leading "/" provides a guaranteed common anchor to filepath.Rel, which makes both inputs absolute paths.
	from := path.Join("/", l.landscape.Target, DirName, componentDir)
	to := path.Join("/", l.landscape.BaseLink, DirName, componentDir)
	rel, err := filepath.Rel(from, to)
	if err != nil {
		// from/to are both absolute and well-formed; this should never error.
		return path.Join(l.landscape.BaseLink, DirName, componentDir)
	}
	return rel
}

// NewLandscapeOptions returns a new LandscapeOptions instance.
//
// opts.TargetDirPath is the on-disk root of the landscape repository.
// Both the base and the landscape components.yaml override files are read from inside this repository:
// the former at landscape.baseLink (where the base content is mounted), the latter at landscape.target.
// The landscape override is applied last so it takes precedence.
func NewLandscapeOptions(opts *generateoptions.Options, fs afero.Afero) (LandscapeOptions, error) {
	repoRoot := path.Clean(opts.TargetDirPath)
	landscape := opts.Config.Repositories.Landscape

	componentVector, err := loadComponentVector(opts, fs,
		// Base components.yaml override, mounted inside the landscape repo at landscape.baseLink.
		path.Join(repoRoot, landscape.BaseLink, utilscomponentvector.ComponentVectorFilename),
		// Landscape-specific components.yaml override.
		path.Join(repoRoot, landscape.Target, utilscomponentvector.ComponentVectorFilename),
	)
	if err != nil {
		return nil, err
	}

	basePath := path.Join(repoRoot, landscape.BaseLink)
	return &landscapeOptions{
		Options:    newOptions(opts, fs, repoRoot, basePath, componentVector),
		landscape:  landscape,
		targetPath: path.Clean(path.Join(repoRoot, landscape.Target)),
	}, nil
}
