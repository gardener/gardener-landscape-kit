// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package files

import (
	"bytes"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/meta"
)

const (
	// GLKSystemDirName is the name of the directory that contains system files for gardener-landscape-kit.
	GLKSystemDirName = ".glk"

	// DefaultDirName is the name of the directory within the GLK system directory that contains the default generated configuration files.
	DefaultDirName = "defaults"

	// glkGitAttributesContent marks all files under .glk/ as generated so that GitHub collapses them in PR diffs.
	glkGitAttributesContent = "** linguist-generated=true\n"

	secretEncryptionDisclaimer = `#
# SECURITY ADVISORY
#
# Gardener-Landscape-Kit has detected that this manifest may contain sensitive data (e.g., a Kubernetes Secret).
# It is strongly recommended not to store unencrypted sensitive data in Git repositories.
# Please ensure that you:
# - Either create the secrets manually in the target cluster and store them securely (e.g., in a vault or password manager),
# - Or use an encryption provider supported by Flux, such as [SOPS](https://fluxcd.io/flux/guides/mozilla-sops/) or [Sealed Secrets](https://fluxcd.io/flux/guides/sealed-secrets/),
# - Or use your own encryption solution.
#
` // #nosec G101 -- No credential.
)

func isSecret(contents []byte) bool {
	kindSecret := "kind: Secret"
	return bytes.HasPrefix(contents, []byte(kindSecret)) || bytes.Contains(contents, []byte("\n"+kindSecret))
}

// stripGLKManagedAnnotations removes GLK-managed annotation comments from YAML bytes.
// A line is annotated when it contains both GLKDefaultPrefix and GLKManagedMarker.
func stripGLKManagedAnnotations(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, meta.GLKManagedMarker) {
			out = append(out, line)
			continue
		}
		// Strip from the start of the GLK annotation prefix.
		if idx := strings.Index(line, meta.GLKDefaultPrefix); idx >= 0 {
			stripped := strings.TrimRight(line[:idx], " \t")
			// If the entire line was a head comment annotation, drop the line entirely.
			if strings.TrimSpace(stripped) == "" {
				continue
			}
			out = append(out, stripped)
		}
		// If for some reason the marker is present but the prefix is not, drop the line.
	}
	return []byte(strings.Join(out, "\n"))
}

// WriteObjectsToFilesystem writes the given objects to the filesystem at the specified rootDir and relativeFilePath.
// If the manifest file already exists, it patches changes from the new default.
// Additionally, it maintains a default version of the manifest in a separate directory for future diff checks.
func WriteObjectsToFilesystem(objects map[string][]byte, rootDir, relativeFilePath string, fs afero.Afero, mode configv1alpha1.MergeMode) error {
	if err := fs.MkdirAll(path.Join(rootDir, relativeFilePath), 0700); err != nil {
		return err
	}
	if err := WriteFileToFilesystem([]byte(glkGitAttributesContent), path.Join(rootDir, GLKSystemDirName, ".gitattributes"), false, fs); err != nil {
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

		if isSecret(object) {
			object = append([]byte(secretEncryptionDisclaimer), object...)
		}

		// Strip GLK-managed annotations from the current file before merging, so they are always re-evaluated (idempotency: the annotation reflects the *current* GLK default).
		currentYaml = stripGLKManagedAnnotations(currentYaml)

		output, err := meta.ThreeWayMergeManifest(oldDefaultYaml, object, currentYaml, mode)
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

// CalculatePathToComponentBase calculates the relative path from the given path splits (from the repository root) to a component base directory.
func CalculatePathToComponentBase(pathSplitsFromRepoRoot ...string) string {
	return RelativePathFromDirDepth(path.Join(pathSplitsFromRepoRoot...))
}
