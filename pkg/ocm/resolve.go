// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package ocm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/sets"
	"ocm.software/open-component-model/bindings/go/descriptor/runtime"
	"sigs.k8s.io/yaml"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/ociaccess"
)

const (
	// CustomOCMComponentNameFilename is the filename containing a custom OCM component name.
	CustomOCMComponentNameFilename = "ocm-component-name"
)

type ocmComponentsResolver struct {
	log          logr.Logger
	cfg          *configv1alpha1.LandscapeKitConfiguration
	landscapeDir string
	outputDir    string
	debug        bool
	workers      int
	components   *components.Components
	repos        []*ociaccess.RepoAccess
}

// ResolveOCMComponents resolves OCM components starting from a root component, processes their dependencies,
// and writes component descriptors and image vectors to the specified output directory.
func ResolveOCMComponents(log logr.Logger, cfg *configv1alpha1.LandscapeKitConfiguration, landscapeDir, outputDir string,
	workers int, debug bool) error {
	// TODO (MartinWeindel): This is a temporary workaround to inform users about potential authentication issues.
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		log.Info("Warning: Environment variable GOOGLE_APPLICATION_CREDENTIALS is not set. Accessing private GCR repositories may fail.")
	}

	repos, err := createRepoAccesses(cfg.OCM)
	if err != nil {
		return err
	}

	resolver := &ocmComponentsResolver{
		log:          log,
		cfg:          cfg,
		landscapeDir: landscapeDir,
		outputDir:    outputDir,
		debug:        debug,
		workers:      workers,
		components:   components.NewComponents(),
		repos:        repos,
	}

	ctx := context.Background()
	return resolver.resolve(ctx)
}

func (r *ocmComponentsResolver) resolve(ctx context.Context) error {
	r.printRepositories()

	if err := r.ensureOutputDirectories(); err != nil {
		return err
	}

	if err := r.walkComponents(ctx); err != nil {
		return err
	}

	if r.debug {
		r.log.Info("Debug mode is enabled, writing additional debug files.")
		if err := r.writeAllImageVectors(); err != nil {
			return err
		}
		if err := r.writeComponentResources(); err != nil {
			return err
		}
	}

	if err := r.writeComponentList(); err != nil {
		return err
	}

	if err := r.writeLandscapeKitComponents(); err != nil {
		return err
	}

	r.log.Info(fmt.Sprintf("Component count: %d", r.components.ComponentsCount()))
	return nil
}

func (r *ocmComponentsResolver) printRepositories() {
	for _, repo := range r.repos {
		r.log.Info("Using repository", "name", repo.Name, "url", repo.RepositoryURL)
	}
}

func (r *ocmComponentsResolver) ensureOutputDirectories() error {
	descriptorDir := path.Join(r.outputDir, "descriptors")
	if err := os.MkdirAll(descriptorDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", descriptorDir, err)
	}
	return nil
}

func (r *ocmComponentsResolver) walkComponents(ctx context.Context) error {
	itemFunc := func(cref components.ComponentReference) ([]components.ComponentReference, error) {
		name, version, err := cref.ExtractNameAndVersion()
		if err != nil {
			return nil, err
		}
		descriptor, blobs, err := ociaccess.FindComponentVersion(ctx, r.log, r.repos, name, version, components.ResourceTypeHelmChartImageMap)
		if err != nil {
			return nil, fmt.Errorf("failed to find component version %s: %w", cref, err)
		}
		r.log.Info("Processing component", "component", cref)

		dv2, err := runtime.ConvertToV2(ociaccess.DefaultScheme, descriptor)
		if err != nil {
			return nil, fmt.Errorf("failed to convert to v2: %w", err)
		}
		data, err := json.MarshalIndent(dv2, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal json: %w", err)
		}
		filename := cref.ToFilename(path.Join(r.outputDir, "descriptors"))
		if err := os.WriteFile(filename, data, 0600); err != nil {
			return nil, fmt.Errorf("failed to write file %s: %w", filename, err)
		}

		return r.components.AddComponentDependencies(descriptor, blobs)
	}

	walker := components.NewComponentWalker(r.log, r.components, r.workers, itemFunc)
	rootComponentReference := components.ComponentReferenceFromNameAndVersion(r.cfg.OCM.RootComponent.Name, r.cfg.OCM.RootComponent.Version)

	if err := walker.Walk(rootComponentReference); err != nil {
		return fmt.Errorf("failed to walk components: %w", err)
	}
	r.log.Info("Finished walking components successfully.", "count", r.components.ComponentsCount())
	return nil
}

func (r *ocmComponentsResolver) writeAllImageVectors() error {
	imagevectorDir := path.Join(r.outputDir, "imagevectors")
	r.log.Info("Writing image vectors to directory", "dir", imagevectorDir)
	for _, cref := range r.components.GetSortedComponents() {
		images, err := r.components.GetImageVector(cref, r.cfg.OCM.OriginalRefs)
		if len(images) == 0 {
			continue
		}
		r.log.Info("Write image vector", "component", cref, "imageCount", len(images))
		if err != nil {
			return fmt.Errorf("failed get image vector for component %s: %w", cref, err)
		}
		if err := writeImageVector(imagevectorDir, cref, images); err != nil {
			return fmt.Errorf("failed to write image vector for component %s: %w", cref, err)
		}
	}
	return nil
}

func (r *ocmComponentsResolver) writeComponentResources() error {
	resourcesDir := path.Join(r.outputDir, "resources")
	r.log.Info("Writing component resources to directory", "dir", resourcesDir)
	for _, cref := range r.components.GetSortedComponents() {
		resources := r.components.GetResources(cref)
		if len(resources) == 0 {
			continue
		}
		r.log.Info("Write resources", "component", cref, "resourceCount", len(resources))
		if err := writeResources(resourcesDir, cref, resources); err != nil {
			return fmt.Errorf("failed to write resources for component %s: %w", cref, err)
		}
	}
	return nil
}

func (r *ocmComponentsResolver) writeComponentList() error {
	listFilename := path.Join(r.outputDir, "component-list.yaml")
	listData, err := r.components.DumpComponentRefListAsYAML()
	if err != nil {
		return fmt.Errorf("failed to dump component list as YAML: %w", err)
	}
	if err := os.WriteFile(listFilename, []byte(listData), 0600); err != nil {
		return fmt.Errorf("failed to write component list file %s: %w", listFilename, err)
	}
	r.log.Info(fmt.Sprintf("Wrote component list to %s", listFilename))
	return nil
}

func (r *ocmComponentsResolver) writeLandscapeKitComponents() error {
	customComponents, err := r.findCustomComponents()
	if err != nil {
		return err
	}
	componentVersions, err := r.components.GetGLKComponents(customComponents, false)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(componentVersions)
	if err != nil {
		return fmt.Errorf("failed to marshal component versions to YAML: %w", err)
	}
	filename := path.Join(r.landscapeDir, "ocm-components.yaml")
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write components file %s: %w", filename, err)
	}
	r.log.Info(fmt.Sprintf("Wrote components file to %s", filename))
	return nil
}

func (r *ocmComponentsResolver) findCustomComponents() (sets.Set[string], error) {
	customComponents := sets.New[string]()

	if err := filepath.WalkDir(r.landscapeDir, func(path string, d os.DirEntry, err error) error {
		if d.IsDir() || d.Name() != CustomOCMComponentNameFilename {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read custom OCM component name file %s: %w", path, err)
		}
		name := strings.TrimSpace(string(content))
		customComponents.Insert(name)
		r.log.Info("Found custom OCM component", "name", name, "file", path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to walk landscape directory %s: %w", r.landscapeDir, err)
	}
	return customComponents, nil
}

func writeImageVector(outputDir string, cref components.ComponentReference, images []imagevector.ImageSource) error {
	if len(images) == 0 {
		return nil
	}
	return writeObject(outputDir, cref, components.ImageVectorOverwriteOutput{
		Images: images,
	})
}

func writeResources(outputDir string, cref components.ComponentReference, resources []components.Resource) error {
	if len(resources) == 0 {
		return nil
	}
	return writeObject(outputDir, cref, components.ResourcesOutput{
		Resources: resources,
	})
}

func writeObject(outputDir string, cref components.ComponentReference, obj any) error {
	output, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}
	outputFile := cref.ToFilename(outputDir)
	return os.WriteFile(outputFile, output, 0600)
}

func createRepoAccesses(cfg *configv1alpha1.OCMConfig) ([]*ociaccess.RepoAccess, error) {
	var repos []*ociaccess.RepoAccess

	for _, url := range cfg.Repositories {
		repo, err := ociaccess.NewRepoAccess(url)
		if err != nil {
			return nil, fmt.Errorf("failed to create RepoAccess for %s repository: %w", url, err)
		}
		repos = append(repos, repo)
	}
	return repos, nil
}
