// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package generate

import (
	"fmt"
	"path"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/components"
)

// GLKSystemDirName is the name of the directory that contains system files for gardener-landscape-kit.
const GLKSystemDirName = ".glk"

// CreateBaseDirStructure creates the base directory structure.
func CreateBaseDirStructure(log logr.Logger, baseDir string, fs afero.Afero) error {
	log.Info("Creating base directory", "baseDir", baseDir)

	for _, dirName := range []string{
		GLKSystemDirName,
		components.DirName,
	} {
		if err := fs.MkdirAll(path.Join(baseDir, dirName), 0744); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dirName, err)
		}
	}

	return nil
}

// CreateLandscapeDirStructure creates the landscape directory structure.
func CreateLandscapeDirStructure(log logr.Logger, landscapeDir string, fs afero.Afero) error {
	log.Info("Creating landscape directory", "landscapeDir", landscapeDir)

	for _, dirName := range []string{
		GLKSystemDirName,
		components.DirName,
	} {
		if err := fs.MkdirAll(path.Join(landscapeDir, dirName), 0744); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dirName, err)
		}
	}

	return nil
}
