// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux

import (
	"embed"
	"path"
	"strings"

	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
)

const (
	// DirName is the directory name where the cluster instances are stored.
	DirName = "flux"

	// FluxComponentsDirName is the directory name where the Flux cli generates the flux-system components into.
	FluxComponentsDirName = DirName + "/flux-system"

	// gitignoreTemplateFile is the name of the .gitignore template file.
	gitignoreTemplateFile = "flux-system/gitignore"
	// gitignoreFileName is the name of the .gitignore file.
	gitignoreFileName = ".gitignore"
)

var (
	// landscapeTemplateDir is the directory where the landscape templates are stored.
	landscapeTemplateDir = "templates/landscape"
	//go:embed templates/landscape
	landscapeTemplates embed.FS
)

type component struct{}

// NewComponent creates a new gardener-operator component.
func NewComponent() components.Interface {
	return &component{}
}

// Name returns the component name.
func (*component) Name() string {
	return "flux"
}

// GenerateBase generates the component base directory.
func (c *component) GenerateBase(_ components.Options) error {
	return nil
}

// GenerateLandscape generates the component landscape directory.
func (c *component) GenerateLandscape(options components.LandscapeOptions) error {
	for _, op := range []func(components.LandscapeOptions) error{
		writeLandscapeTemplateFiles,
		writeGitignoreFile,
		generateFirstStepsMessageIfRequired(options),
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

func writeLandscapeTemplateFiles(opts components.LandscapeOptions) error {
	var repoRef string
	switch {
	case opts.GetGitRepository().Ref.Commit != nil:
		repoRef = "commit: " + *opts.GetGitRepository().Ref.Commit
	case opts.GetGitRepository().Ref.Tag != nil:
		repoRef = "tag: " + *opts.GetGitRepository().Ref.Tag
	case opts.GetGitRepository().Ref.Branch != nil:
		repoRef = "branch: " + *opts.GetGitRepository().Ref.Branch
	default:
		repoRef = "branch: main"
	}

	fluxPath := path.Join(opts.GetRelativeLandscapePath(), DirName)
	fluxPath = strings.TrimPrefix(fluxPath, "./")

	objects, err := files.RenderTemplateFiles(landscapeTemplates, landscapeTemplateDir, map[string]any{
		"repo_url":  opts.GetGitRepository().URL,
		"repo_ref":  repoRef,
		"flux_path": fluxPath,
	})
	if err != nil {
		return err
	}

	// Remove files: gitignore is handled separately, doc.go is not needed in the filesystem.
	delete(objects, "flux-system/gitignore")
	delete(objects, "flux-system/doc.go")

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), DirName, opts.GetFilesystem())
}

func writeGitignoreFile(options components.LandscapeOptions) error {
	gitignore, err := landscapeTemplates.ReadFile(path.Join(landscapeTemplateDir, gitignoreTemplateFile))
	if err != nil {
		return err
	}
	gitignoreDefaultPath := path.Join(options.GetTargetPath(), files.GLKSystemDirName, files.DefaultDirName, FluxComponentsDirName, gitignoreFileName)

	fileDefaultExists, err := options.GetFilesystem().Exists(gitignoreDefaultPath)
	if err == nil && !fileDefaultExists {
		if err := files.WriteFileToFilesystem(gitignore, path.Join(options.GetTargetPath(), FluxComponentsDirName, gitignoreFileName), false, options.GetFilesystem()); err != nil {
			return err
		}
	}
	// Write the default gitignore file to the .glk defaults system directory.
	return files.WriteFileToFilesystem(gitignore, gitignoreDefaultPath, true, options.GetFilesystem())
}

func generateFirstStepsMessageIfRequired(options components.LandscapeOptions) func(options components.LandscapeOptions) error {
	landscapeDir := options.GetTargetPath()
	instanceFileExisted, err := options.GetFilesystem().DirExists(path.Join(landscapeDir, FluxComponentsDirName))
	return func(options components.LandscapeOptions) error {
		if err != nil || instanceFileExisted {
			return err
		}
		fluxDir := path.Join(landscapeDir, DirName)
		options.GetLogger().Info(`Initialized the landscape for an expected Flux cluster at: ` + fluxDir + `

Next steps:
1. Adjust the generated manifests to your environment, especially the Git repository reference:

   # Directory with initial flux manifests: ` + fluxDir + `

2. Target the cluster to install Flux in:

  $  KUBECONFIG=...

3. Install the Flux CRDs initially:

   $  kubectl create -f ` + path.Join(landscapeDir, FluxComponentsDirName, "gotk-components.yaml") + `

4. You might want to consider creating the Git sync credentials manually and store them separately instead of checking them into Git:

   $  kubectl create -f ` + path.Join(landscapeDir, FluxComponentsDirName, "git-sync-secret.yaml") + `

5. Commit and push the changes to your landscape git repository.

6. Deploy Flux on the cluster:

  $  kubectl apply -k ` + path.Join(landscapeDir, FluxComponentsDirName) + `
`)
		return nil
	}
}
