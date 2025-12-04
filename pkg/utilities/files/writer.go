// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/meta"
)

const (
	// GLKSystemDirName is the name of the directory that contains system files for gardener-landscape-kit.
	GLKSystemDirName = ".glk"

	// DefaultDirName is the name of the directory within the GLK system directory that contains the default generated configuration files.
	DefaultDirName = "defaults"
)

// WriteObjectsToFilesystem writes the given objects to the filesystem at the specified rootDir and relativeFilePath.
// If the manifest file already exists, it patches changes from the new default.
// Additionally, it maintains a default version of the manifest in a separate directory for future diff checks.
func WriteObjectsToFilesystem(objects map[string][]byte, rootDir, relativeFilePath string, fs afero.Afero) error {
	if err := fs.MkdirAll(path.Join(rootDir, relativeFilePath), 0700); err != nil {
		return err
	}

	for fileName, object := range objects {
		filePath := path.Join(relativeFilePath, fileName)

		filePathCurrent := path.Join(rootDir, filePath)
		currentYaml, err := fs.ReadFile(filePathCurrent)
		isCurrentNotExistsErr := os.IsNotExist(err)
		if err != nil && !isCurrentNotExistsErr {
			return err
		}

		filePathDefault := path.Join(rootDir, GLKSystemDirName, DefaultDirName, filePath)
		oldDefaultYaml, err := fs.ReadFile(filePathDefault)
		isDefaultNotExistsErr := os.IsNotExist(err)
		if err != nil && !isDefaultNotExistsErr {
			return err
		}

		if !isDefaultNotExistsErr && len(oldDefaultYaml) > 0 && isCurrentNotExistsErr {
			// File has been deleted by the user. Do not recreate until the default file within the .glk directory is deleted.
			continue
		}

		output, err := meta.ThreeWayMergeManifest(oldDefaultYaml, object, currentYaml)
		if err != nil {
			return err
		}
		// write new manifest
		if err := WriteFileToFilesystem(output, filePathCurrent, true, fs); err != nil {
			return err
		}
		// write new default
		if err := WriteFileToFilesystem(object, filePathDefault, true, fs); err != nil {
			return err
		}
	}

	return nil
}

// WriteFileToFilesystem writes the given file to the filesystem at the specified filePathDir.
// If overwriteExisting is false and the file already exists, it does nothing.
func WriteFileToFilesystem(contents []byte, filePathDir string, overwriteExisting bool, fs afero.Afero) error {
	exists, err := fs.Exists(filePathDir)
	if err != nil {
		return err
	}
	if !exists || overwriteExisting {
		if err := fs.MkdirAll(path.Dir(filePathDir), 0700); err != nil {
			return err
		}
		return fs.WriteFile(filePathDir, contents, 0600)
	}

	return nil
}

// RelativePathFromDirDepth returns a relative path that goes up the directory tree
// based on the depth of the given relativePath.
// If the passed path is already a relative path, it will log a fatal error.
func RelativePathFromDirDepth(relativePath string) string {
	relativePath, err := filepath.Rel("./"+relativePath, "./")
	if err != nil {
		log.Fatal(err)
	}
	return relativePath
}
