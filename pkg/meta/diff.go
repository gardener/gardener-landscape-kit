// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"fmt"
	"os"
	"path"
	"reflect"

	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	// GLKSystemDirName is the name of the directory that contains system files for gardener-landscape-kit.
	GLKSystemDirName = ".glk"

	// DefaultDirName is the name of the directory within the GLK system directory that contains the default generated configuration files.
	DefaultDirName = "defaults"
)

// CreateOrUpdateManifest creates or updates a manifest file at the given filePath within the landscapeDir.
// If the manifest file already exists, it performs a strategic merge patch with the newDefaultObj.
// Additionally, it maintains a default version of the manifest in a separate directory for future diff checks.
func CreateOrUpdateManifest[T client.Object](newDefaultObj T, landscapeDir, filePath string, fs afero.Afero) error {
	filePathManifest := path.Join(landscapeDir, filePath)
	oldManifest, err := fs.ReadFile(filePathManifest)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	filePathDefault := path.Join(landscapeDir, GLKSystemDirName, DefaultDirName, filePath)
	oldDefaultManifest, err := fs.ReadFile(filePathDefault)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	newDefaultManifest, err := yaml.Marshal(newDefaultObj)
	if err != nil {
		return err
	}
	newManifest := newDefaultManifest // default in case no existing manifest is present
	for _, dir := range []string{
		filePathManifest,
		filePathDefault,
	} {
		if err := fs.MkdirAll(path.Dir(dir), 0700); err != nil {
			return err
		}
	}
	if len(oldManifest) > 0 {
		// Output file already exists, perform diff check and write only patched output
		var oldDefaultObj T
		if err := yaml.Unmarshal(oldDefaultManifest, &oldDefaultObj); err != nil {
			return err
		}
		patch, err := client.StrategicMergeFrom(oldDefaultObj).Data(newDefaultObj)
		if err != nil {
			return err
		}
		oldJSON, err := yaml.YAMLToJSON(oldManifest)
		if err != nil {
			return err
		}
		var patchTarget interface{}
		t := reflect.TypeOf(newDefaultObj)
		if t.Kind() == reflect.Pointer {
			patchTarget = reflect.New(t.Elem()).Elem().Interface()
		} else {
			patchTarget = reflect.New(t).Elem().Interface()
		}
		newJSON, err := strategicpatch.StrategicMergePatch(oldJSON, patch, patchTarget)
		if err != nil {
			return fmt.Errorf("could not patch existing manifest file at '%s': %w", filePathManifest, err)
		}
		newManifest, err = yaml.JSONToYAML(newJSON)
		if err != nil {
			return err
		}
	}

	if err := fs.WriteFile(filePathManifest, newManifest, 0600); err != nil {
		return err
	}

	// Write new metadata file
	if err := fs.WriteFile(filePathDefault, newDefaultManifest, 0600); err != nil {
		return err
	}
	return nil
}
