// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux

import (
	"embed"
	"path"
	"strings"

	"github.com/gardener/gardener/pkg/utils"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

const (
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
	//go:embed meta.yaml
	metadataYAML []byte
)

type component struct {
	*components.Metadata

	fluxComponentsDirName string
}

// NewComponent creates a new gardener-operator component.
func NewComponent() (components.Interface, error) {
	metadata, err := components.NewMetadata(metadataYAML)
	if err != nil {
		return nil, err
	}

	return &component{
		Metadata:              metadata,
		fluxComponentsDirName: metadata.Directory + "/flux-system",
	}, nil
}

// GenerateBase generates the component base directory.
func (c *component) GenerateBase(_ components.Context, _ components.Options) error {
	return nil
}

// GenerateLandscape generates the component landscape directory.
func (c *component) GenerateLandscape(_ components.Context, options components.LandscapeOptions) error {
	for _, op := range []func(components.LandscapeOptions) error{
		c.writeLandscapeTemplateFiles,
		c.writeGitignoreFile,
		c.generateFirstStepsMessageIfRequired(options),
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

func (c *component) writeLandscapeTemplateFiles(opts components.LandscapeOptions) error {
	ref := opts.GetLandscapeRef()
	var repoRef string
	switch {
	case ref.Commit != nil:
		repoRef = "commit: " + *ref.Commit
	case ref.Tag != nil:
		repoRef = "tag: " + *ref.Tag
	case ref.Branch != nil:
		repoRef = "branch: " + *ref.Branch
	default:
		repoRef = "branch: main"
	}

	fluxPath := path.Join(opts.GetRelativeLandscapePath(), c.Directory)
	fluxPath = strings.TrimPrefix(fluxPath, "./")

	cvValues, err := components.GetComponentVectorTemplateValues(opts, componentvector.NameGardenerGardenerLandscapeKit)
	if err != nil {
		return err
	}

	values := utils.MergeMaps(cvValues, map[string]any{
		"sourceKind":      opts.GetSourceKind(),
		"source_url":      opts.GetLandscapeURL(),
		"source_ref":      repoRef,
		"flux_path":       fluxPath,
		"components_path": path.Join(opts.GetRelativeLandscapePath(), components.DirName),
	})

	objects, err := files.RenderTemplateFiles(landscapeTemplates, landscapeTemplateDir, values)
	if err != nil {
		return err
	}

	// Remove files: gitignore is handled separately, doc.go is not needed in the filesystem.
	delete(objects, "flux-system/gitignore")
	delete(objects, "flux-system/doc.go")

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), c.Directory, opts.GetFilesystem(), opts.GetMergeMode())
}

func (c *component) writeGitignoreFile(options components.LandscapeOptions) error {
	gitignore, err := landscapeTemplates.ReadFile(path.Join(landscapeTemplateDir, gitignoreTemplateFile))
	if err != nil {
		return err
	}
	gitignoreDefaultPath := path.Join(options.GetTargetPath(), files.GLKSystemDirName, files.DefaultDirName, c.fluxComponentsDirName, gitignoreFileName)

	fileDefaultExists, err := options.GetFilesystem().Exists(gitignoreDefaultPath)
	if err == nil && !fileDefaultExists {
		if err := files.WriteFileToFilesystem(gitignore, path.Join(options.GetTargetPath(), c.fluxComponentsDirName, gitignoreFileName), false, options.GetFilesystem()); err != nil {
			return err
		}
	}
	// Write the default gitignore file to the .glk defaults system directory.
	return files.WriteFileToFilesystem(gitignore, gitignoreDefaultPath, true, options.GetFilesystem())
}

func (c *component) generateFirstStepsMessageIfRequired(options components.LandscapeOptions) func(options components.LandscapeOptions) error {
	landscapeDir := options.GetTargetPath()
	instanceFileExisted, err := options.GetFilesystem().DirExists(path.Join(landscapeDir, c.fluxComponentsDirName))
	return func(options components.LandscapeOptions) error {
		if err != nil || instanceFileExisted {
			return err
		}
		fluxDir := path.Join(landscapeDir, c.Directory)
		options.GetLogger().Info(`Initialized the landscape for an expected Flux cluster at: ` + fluxDir + `

Next steps:
1. Adjust the generated manifests to your environment, especially the Git repository reference:

   # Directory with initial flux manifests: ` + fluxDir + `

2. Target the cluster to install Flux in:

  $  KUBECONFIG=...

3. Install the Flux CRDs initially:

   $  kubectl create -f ` + path.Join(landscapeDir, c.fluxComponentsDirName, "gotk-components.yaml") + `

4. You might want to consider creating the Git sync credentials manually and store them separately instead of checking them into Git:

   $  kubectl create -f ` + path.Join(landscapeDir, c.fluxComponentsDirName, "git-sync-secret.yaml") + `

5. Commit and push the changes to your landscape git repository.

6. Deploy Flux on the cluster:

  $  kubectl apply -k ` + path.Join(landscapeDir, c.fluxComponentsDirName) + `
`)
		return nil
	}
}
