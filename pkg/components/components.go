// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"github.com/spf13/afero"
)

const (
	// DirName is the directory name where components are stored.
	DirName = "components"
)

// Options is an interface for options passed to components.
type Options interface {
	// GetBaseDir returns the base directory that serves as the foundation (base) for any landscape.
	GetBaseDir() string
	// GetLandscapeDir returns the landscape directory. If the returned path is empty, only the base directory should be generated.
	GetLandscapeDir() string
	// GetFilesystem returns the filesystem to use.
	GetFilesystem() afero.Afero
}

// Interface is the components interface that each component must implement.
type Interface interface {
	// Generate generates the component.
	Generate(Options) error
}

type options struct {
	baseDir      string
	landscapeDir string
	filesystem   afero.Afero
}

// GetBaseDir returns the base directory that serves as the foundation (base) for any landscape.
func (o options) GetBaseDir() string {
	return o.baseDir
}

// GetLandscapeDir returns the landscape directory. If the returned path is empty, only the base directory should be generated.
func (o options) GetLandscapeDir() string {
	return o.landscapeDir
}

// GetFilesystem returns the filesystem to use.
func (o options) GetFilesystem() afero.Afero {
	return o.filesystem
}

// NewOptions returns a new Options instance.
func NewOptions(baseDir string, landscapeDir string, fs afero.Afero) Options {
	return &options{
		baseDir:      baseDir,
		landscapeDir: landscapeDir,
		filesystem:   fs,
	}
}
