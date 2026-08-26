// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	_ "embed"
	"fmt"
	"io/fs"
	"path"

	dotgithub "github.com/gardener/gardener-landscape-kit/.github"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

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

var (
	//go:embed meta.yaml
	metadataYAML []byte
)

type component struct {
	*components.Metadata
}

// NewComponent creates a new github component.
func NewComponent() (components.Interface, error) {
	metadata, err := components.NewMetadata(metadataYAML)
	if err != nil {
		return nil, err
	}
	return &component{Metadata: metadata}, nil
}

// GenerateBase materializes the embedded .github/ assets into the repository root during base generation.
func (c *component) GenerateBase(_ components.Context, opts components.Options) error {
	return c.writeDotGitHub(opts)
}

// GenerateLandscape materializes the embedded .github/ assets into the repository root during landscape generation.
func (c *component) GenerateLandscape(_ components.Context, opts components.LandscapeOptions) error {
	return c.writeDotGitHub(opts)
}

// writeDotGitHub walks the embedded sources and writes each file directly to the repository root with the disclaimer header prepended, overwriting any existing content.
func (c *component) writeDotGitHub(opts components.Options) error {
	dotGitHubRoot := path.Join(opts.GetRepoRoot(), c.Directory)
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
