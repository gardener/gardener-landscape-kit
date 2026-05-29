// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"io/fs"
	"path"

	dotgithub "github.com/gardener/gardener-landscape-kit/.github"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

// ComponentName is the name of the github component.
const ComponentName = "github"

// dotGitHubDir is the destination directory at the repository root.
const dotGitHubDir = ".github"

// DisclaimerHeader is prepended to every file written by this component.
const DisclaimerHeader = `# This file is managed by gardener-landscape-kit (github component) and will be
# overwritten on every 'glk generate' invocation. To stop glk from managing
# this file, exclude the 'github' component in your glk configuration:
#
#   components:
#     exclude:
#     - github
#
`

type component struct{}

// NewComponent creates a new github component.
func NewComponent() components.Interface {
	return &component{}
}

// Name returns the component name.
func (*component) Name() string {
	return ComponentName
}

// GenerateBase materializes the embedded .github/ assets into the repository root during base generation.
func (*component) GenerateBase(opts components.Options) error {
	return writeDotGitHub(opts)
}

// GenerateLandscape materializes the embedded .github/ assets into the repository root during landscape generation.
func (*component) GenerateLandscape(opts components.LandscapeOptions) error {
	return writeDotGitHub(opts)
}

// writeDotGitHub walks the embedded sources and writes each file directly to the repository root with the disclaimer header prepended, overwriting any existing content.
func writeDotGitHub(opts components.Options) error {
	dotGitHubRoot := path.Join(opts.GetRepoRoot(), dotGitHubDir)
	for _, src := range []struct {
		fs   fs.FS
		root string
	}{
		{dotgithub.DotGitHubActions, "actions"},
		{dotgithub.DotGitHubWorkflows, "workflows"},
	} {
		if err := writeEmbedded(src.fs, src.root, dotGitHubRoot, opts); err != nil {
			return err
		}
	}
	return nil
}

// writeEmbedded walks srcRoot in srcFS and writes each regular file's contents (with the disclaimer header prepended) to destRoot joined with the file's path relative to the embed root.
// Files are always overwritten. No template rendering is performed (action/workflow YAML uses ${{ ... }} which would collide with Go template delimiters).
func writeEmbedded(srcFS fs.FS, srcRoot, destRoot string, opts components.Options) error {
	return fs.WalkDir(srcFS, srcRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		contents, err := fs.ReadFile(srcFS, p)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", p, err)
		}
		withHeader := append([]byte(DisclaimerHeader), contents...)
		destPath := path.Join(destRoot, p)
		return files.WriteFileToFilesystem(withHeader, destPath, true, opts.GetFilesystem())
	})
}
