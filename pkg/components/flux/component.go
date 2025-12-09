// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/Masterminds/sprig/v3"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
	"github.com/gardener/gardener-landscape-kit/pkg/utilities/kustomization"
)

const (
	// DirName is the directory name where the cluster instances are stored.
	DirName = "flux"

	// FluxComponentsDirName is the directory name where the Flux cli generates the flux-system components into.
	FluxComponentsDirName = DirName + "/flux-system"

	// glkComponentsName is the name of the Flux Kustomize component that serves as the root for all subsequent components.
	glkComponentsName = "glk-components"

	// gitignoreTemplateFile is the name of the .gitignore template file.
	gitignoreTemplateFile = "gitignore"
	// gitignoreFileName is the name of the .gitignore file.
	gitignoreFileName = ".gitignore"
	// gitSecretFileName is the name of the template file for the Git sync secret which should be created manually and not checked into the landscape Git repo.
	gitSecretFileName = "git-sync-secret.yaml"
	// gotkSyncFileName is the name of the template file for the gotk sync manifest.
	gotkSyncFileName = "gotk-sync.yaml.tpl"
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
		writeFluxTemplateFilesAndKustomization,
		writeGitignoreFile,
		writeGardenNamespaceManifest, // The `garden` namespace will hold all Flux resources (related to gardener components) in the cluster and must be created as soon as possible.
		writeFluxKustomization,
		logFluxInitializationFirstSteps,
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

func writeFluxTemplateFilesAndKustomization(options components.LandscapeOptions) error {
	var (
		objects                    = make(map[string][]byte)
		kustomizationObjectEntries []string
	)
	dir, err := landscapeTemplates.ReadDir(landscapeTemplateDir)
	if err != nil {
		return err
	}
	for _, file := range dir {
		fileName := file.Name()
		if fileName == gitignoreTemplateFile {
			continue
		}

		fileContents, err := landscapeTemplates.ReadFile(path.Join(landscapeTemplateDir, fileName))
		if err != nil {
			return err
		}

		if fileName == gotkSyncFileName {
			fileContents, fileName, err = renderGOTKTemplate(options, fileContents, fileName)
			if err != nil {
				return err
			}
		}

		if fileName != gitSecretFileName {
			kustomizationObjectEntries = append(kustomizationObjectEntries, fileName)
		}
		objects[fileName] = fileContents
	}

	kustomizationManifest := kustomization.NewKustomization(kustomizationObjectEntries, nil)
	objects[kustomization.KustomizationFileName], err = yaml.Marshal(kustomizationManifest)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, options.GetTargetPath(), FluxComponentsDirName, options.GetFilesystem())
}

func renderGOTKTemplate(options components.LandscapeOptions, fileContents []byte, fileName string) ([]byte, string, error) {
	gotkTemplate, err := template.New("gotk-sync").Funcs(sprig.TxtFuncMap()).Parse(string(fileContents))
	if err != nil {
		return nil, "", fmt.Errorf("error parsing gotk sync template: %w", err)
	}

	var repoRef string
	switch {
	case options.GetGitRepository().Ref.Commit != nil:
		repoRef = "commit: " + *options.GetGitRepository().Ref.Commit
	case options.GetGitRepository().Ref.Tag != nil:
		repoRef = "tag: " + *options.GetGitRepository().Ref.Tag
	case options.GetGitRepository().Ref.Branch != nil:
		repoRef = "branch: " + *options.GetGitRepository().Ref.Branch
	default:
		repoRef = "branch: main"
	}

	fluxPath := path.Join(options.GetRelativeLandscapePath(), DirName)
	fluxPath = strings.TrimPrefix(fluxPath, "./")

	var gotkResult bytes.Buffer
	if err := gotkTemplate.Execute(&gotkResult, map[string]any{
		"repo_url":  options.GetGitRepository().URL,
		"repo_ref":  repoRef,
		"flux_path": fluxPath,
	}); err != nil {
		return nil, "", fmt.Errorf("error executing gotk sync template: %w", err)
	}

	fileContents = gotkResult.Bytes()
	fileName = strings.TrimSuffix(fileName, ".tpl")
	return fileContents, fileName, nil
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

func writeFluxKustomization(options components.LandscapeOptions) error {
	fluxKustomization, err := yaml.Marshal(&kustomizev1.Kustomization{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kustomizev1.GroupVersion.String(),
			Kind:       kustomizev1.KustomizationKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      glkComponentsName,
			Namespace: FluxSystemNamespaceName,
		},
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: SourceRef,
			Path:      path.Join(options.GetRelativeLandscapePath(), components.DirName),
		},
	})
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(
		map[string][]byte{
			glkComponentsName + ".yaml": fluxKustomization,
		},
		options.GetTargetPath(),
		DirName,
		options.GetFilesystem(),
	)
}

func logFluxInitializationFirstSteps(options components.LandscapeOptions) error {
	landscapeDir := options.GetTargetPath()
	if instanceFileExisted, err := options.GetFilesystem().DirExists(path.Join(landscapeDir, FluxComponentsDirName)); err != nil || instanceFileExisted {
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

const (
	// NamespaceKind is the kind of the namespace resource.
	NamespaceKind = "Namespace"
	// GardenNamespaceFileName is the name of the namespace manifest file.
	GardenNamespaceFileName = "garden-namespace.yaml"
	// GardenNamespaceName is the name of the namespace created by this component.
	GardenNamespaceName = "garden"
)

// writeGardenNamespaceManifest generates the garden namespace in the given landscape directory.
func writeGardenNamespaceManifest(options components.LandscapeOptions) error {
	objects := make(map[string][]byte)

	var err error
	objects[GardenNamespaceFileName], err = yaml.Marshal(&corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       NamespaceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: GardenNamespaceName,
		},
	})
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, options.GetTargetPath(), DirName, options.GetFilesystem())
}
