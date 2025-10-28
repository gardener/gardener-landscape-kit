// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"encoding/json"
	"maps"
	"os"
	"path"
	"slices"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gardener/gardener/pkg/utils"
	goccyyaml "github.com/goccy/go-yaml"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	// GLKSystemDirName is the name of the directory that contains system files for gardener-landscape-kit.
	GLKSystemDirName = ".glk"

	// DefaultDirName is the name of the directory within the GLK system directory that contains the default generated configuration files.
	DefaultDirName = "defaults"
)

// CreateOrUpdateManifest creates or updates a manifest file at the given filePath within the baseDir based on a given YAML object.
// If the manifest file already exists, it patches changes from the newDefaultYaml.
// Additionally, it maintains a default version of the manifest in a separate directory for future diff checks.
func CreateOrUpdateManifest(newDefaultYaml []byte, baseDir, filePath string, fs afero.Afero) error {
	newManifest := newDefaultYaml // default in case no existing manifest is present
	var newDefaultObj map[string]interface{}
	newDefaultComments := goccyyaml.CommentMap{}
	if err := goccyyaml.UnmarshalWithOptions(newDefaultYaml, &newDefaultObj, goccyyaml.CommentToMap(newDefaultComments)); err != nil {
		return err
	}

	filePathManifest := path.Join(baseDir, filePath)
	oldManifest, err := fs.ReadFile(filePathManifest)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	filePathDefault := path.Join(baseDir, GLKSystemDirName, DefaultDirName, filePath)
	oldDefaultManifest, err := fs.ReadFile(filePathDefault)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for _, dir := range []string{
		filePathManifest,
		filePathDefault,
	} {
		if err := fs.MkdirAll(path.Dir(dir), 0700); err != nil {
			return err
		}
	}
	if len(oldManifest) > 0 {
		var oldDefaultObject map[string]interface{}
		oldDefaultComments := goccyyaml.CommentMap{}
		if err := goccyyaml.UnmarshalWithOptions(oldDefaultManifest, &oldDefaultObject, goccyyaml.CommentToMap(oldDefaultComments)); err != nil {
			return err
		}

		var oldObject map[string]interface{}
		oldComments := goccyyaml.CommentMap{}
		if err := goccyyaml.UnmarshalWithOptions(oldManifest, &oldObject, goccyyaml.CommentToMap(oldComments)); err != nil {
			return err
		}

		oldDefaultJson, err := json.Marshal(oldDefaultObject)
		if err != nil {
			return err
		}
		newDefaultJson, err := json.Marshal(newDefaultObj)
		if err != nil {
			return err
		}
		oldJson, err := json.Marshal(oldObject)
		if err != nil {
			return err
		}

		patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(oldJson, newDefaultJson, oldDefaultJson)
		if err != nil {
			return err
		}

		newJson, err := jsonpatch.MergePatch(oldJson, patch)
		if err != nil {
			return err
		}

		var newObject map[string]interface{}
		if err := json.Unmarshal(newJson, &newObject); err != nil {
			return err
		}
		newManifest, err = goccyyaml.MarshalWithOptions(newObject, goccyyaml.WithComment(mergeComments(oldDefaultComments, newDefaultComments, oldComments)))
		if err != nil {
			return err
		}
	}

	if err := fs.WriteFile(filePathManifest, newManifest, 0600); err != nil {
		return err
	}

	// Write new metadata file
	if err := fs.WriteFile(filePathDefault, newDefaultYaml, 0600); err != nil {
		return err
	}
	return nil
}

func mergeComments(oldDefault, newDefault, manifest goccyyaml.CommentMap) goccyyaml.CommentMap {
	var (
		oldDefaultComments = make(map[string]map[goccyyaml.CommentPosition]*goccyyaml.Comment, len(oldDefault))
		newDefaultComments = make(map[string]map[goccyyaml.CommentPosition]*goccyyaml.Comment, len(newDefault))
		manifestComments   = make(map[string]map[goccyyaml.CommentPosition]*goccyyaml.Comment, len(manifest))
	)

	for commentPath, comments := range oldDefault {
		oldDefaultComments[commentPath] = utils.CreateMapFromSlice(comments, func(cm *goccyyaml.Comment) goccyyaml.CommentPosition {
			return cm.Position
		})
	}
	for commentPath, comments := range newDefault {
		newDefaultComments[commentPath] = utils.CreateMapFromSlice(comments, func(cm *goccyyaml.Comment) goccyyaml.CommentPosition {
			return cm.Position
		})
	}
	for commentPath, comments := range manifest {
		manifestComments[commentPath] = utils.CreateMapFromSlice(comments, func(cm *goccyyaml.Comment) goccyyaml.CommentPosition {
			return cm.Position
		})
	}

	oldDefaultCommentsSet := sets.New(slices.Collect(maps.Keys(oldDefaultComments))...)
	newDefaultCommentsSet := sets.New(slices.Collect(maps.Keys(newDefaultComments))...)
	sameCommentPaths := oldDefaultCommentsSet.Intersection(newDefaultCommentsSet)
	addedCommentPaths := newDefaultCommentsSet.Difference(oldDefaultCommentsSet)

	for commentPath := range sameCommentPaths.Union(addedCommentPaths) {
		for position, comment := range newDefaultComments[commentPath] {
			if oldDefaultComment, exists := oldDefaultComments[commentPath][position]; exists {
				if newDefaultComment, exists := newDefaultComments[commentPath][position]; exists && slices.Equal(oldDefaultComment.Texts, newDefaultComment.Texts) {
					continue
				}
			}
			if _, exists := manifestComments[commentPath]; !exists {
				manifestComments[commentPath] = map[goccyyaml.CommentPosition]*goccyyaml.Comment{
					position: comment,
				}
				continue
			}
			if _, exists := manifestComments[commentPath][position]; exists {
				// Append new comment texts before existing comment in manifest.
				manifestComments[commentPath][position].Texts = append(comment.Texts, manifestComments[commentPath][position].Texts...)
			} else {
				// Add new comment at this position.
				manifestComments[commentPath][position] = comment
			}
		}
	}

	for commentPath, comments := range manifestComments {
		manifest[commentPath] = nil
		for _, c := range comments {
			manifest[commentPath] = append(manifest[commentPath], c)
		}
	}

	return manifest
}
