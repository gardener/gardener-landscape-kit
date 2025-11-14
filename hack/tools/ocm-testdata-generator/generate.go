// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"k8s.io/utils/set"
	"ocm.software/open-component-model/bindings/go/descriptor/runtime"
	descriptorv2 "ocm.software/open-component-model/bindings/go/descriptor/v2"
	accessv1 "ocm.software/open-component-model/bindings/go/oci/spec/access/v1"

	"github.com/gardener/gardener-landscape-kit/pkg/ocm/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/ociaccess"
)

const gardenerRepositoryURL = "europe-docker.pkg.dev/gardener-project/releases"

type generator struct {
	config *Config
}

// OCM represents the structure of an OCM component descriptor.
type OCM struct {
	Meta      OCMMeta      `json:"meta"`
	Component OCMComponent `json:"component"`
}

// OCMMeta contains metadata about the OCM component descriptor.
type OCMMeta struct {
	SchemaVersion string `json:"schemaVersion"`
}

// OCMComponent represents an OCM component with its details.
type OCMComponent struct {
	Name                string           `json:"name"`
	Version             string           `json:"version"`
	Labels              []OCMLabel       `json:"labels"`
	Provider            string           `json:"provider"`
	Resources           []map[string]any `json:"resources"`
	ComponentReferences []map[string]any `json:"componentReferences"`
}

// OCMLabel represents a label in the OCM component descriptor.
type OCMLabel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// NewGenerator creates a new generator for OCM component descriptors test data with the given config.
func NewGenerator(config *Config) *generator {
	return &generator{
		config: config,
	}
}

// Generate generates the OCM component descriptor test data and writes it to the target directory.
func (g *generator) Generate(targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	if err := g.generateKubernetesRootComponent(targetDir); err != nil {
		return fmt.Errorf("failed to generate kubernetes root component: %w", err)
	}

	repoAccess, err := ociaccess.NewRepoAccess(gardenerRepositoryURL)
	if err != nil {
		return fmt.Errorf("failed to create repo access: %w", err)
	}

	toBeProcessed := slices.Clone(g.config.Components)
	processedComponents := set.New[components.ComponentReference]()
	for len(toBeProcessed) > 0 {
		current := toBeProcessed[0]
		toBeProcessed = toBeProcessed[1:]
		cr := components.ComponentReferenceFromNameAndVersion(current.ComponentName, current.Version)
		if processedComponents.Has(cr) {
			continue
		}
		componentRefs, err := g.processComponent(repoAccess, current, targetDir)
		if err != nil {
			return fmt.Errorf("failed to process component %s:%s: %w", current.ComponentName, current.Version, err)
		}
		processedComponents.Insert(cr)
		toBeProcessed = append(toBeProcessed, componentRefs...)
	}

	return nil
}

func (g *generator) processComponent(repoAccess *ociaccess.RepoAccess, comp Component, targetDir string) ([]Component, error) {
	println("Processing component:", comp.ComponentName, comp.Version)
	descriptor, err := repoAccess.GetComponentVersion(context.Background(), comp.ComponentName, comp.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get component version %s:%s: %w", comp.ComponentName, comp.Version, err)
	}
	dv2, err := runtime.ConvertToV2(ociaccess.DefaultScheme, descriptor)

	// Modify the resource access to point to the target repository with the calculated SHA256 digest.
	for i := range dv2.Component.Resources {
		res := &dv2.Component.Resources[i]
		if res.Type == components.ResourceTypeOCIImage || res.Type == components.ResourceTypeHelmChart {
			var spec accessv1.OCIImage
			if err := ociaccess.DefaultScheme.Convert(res.Access, &spec); err != nil {
				return nil, fmt.Errorf("failed to convert from access of resource %s to OCIImage: %w", res.Name, err)
			}
			res.Labels = append(res.Labels, descriptorv2.Label{
				Name:  components.LabelNameOriginalRef,
				Value: []byte(strconv.Quote(spec.ImageReference)),
			})
			parts := strings.SplitN(spec.ImageReference, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid image reference format: %s", spec.ImageReference)
			}
			spec.ImageReference = fmt.Sprintf("%s/%s:%s@%s",
				g.config.TargetRepositoryURL,
				strings.ReplaceAll(parts[0], ".", "_"),
				res.Version,
				calcSampleSHA256(res.Name, res.Version),
			)
			if err := ociaccess.DefaultScheme.Convert(&spec, res.Access); err != nil {
				return nil, fmt.Errorf("failed to convert to access of resource %s to OCIImage: %w", res.Name, err)
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to convert to v2: %w", err)
	}

	cr := components.ComponentReferenceFromNameAndVersion(comp.ComponentName, comp.Version)
	if err := writeComponentDescriptor(cr, dv2, targetDir); err != nil {
		return nil, err
	}

	// collect component references
	componentRefs := []Component{comp}
	for _, ref := range dv2.Component.References {
		componentRefs = append(componentRefs, Component{
			Name:          ref.Name,
			Version:       ref.Version,
			ComponentName: ref.Component,
		})
	}

	return componentRefs, nil
}

func (g *generator) generateKubernetesRootComponent(targetDir string) error {
	resourceSlice := make([]map[string]any, 0)
	resourceSlice = g.addKubernetesVersionResources(resourceSlice)
	resourceSlice = g.addExternalResources(resourceSlice)

	componentSlice := make([]map[string]any, 0)
	for _, comp := range g.config.Components {
		component := map[string]any{
			"name":          comp.Name,
			"version":       comp.Version,
			"componentName": comp.ComponentName,
		}
		componentSlice = append(componentSlice, component)
	}

	ocm := OCM{
		Meta: OCMMeta{
			SchemaVersion: "v2",
		},
		Component: OCMComponent{
			Name:    g.config.RootComponent.ComponentName,
			Version: g.config.RootComponent.Version,
			Labels: []OCMLabel{
				{
					Name:  components.LabelImageVectorApplication,
					Value: "kubernetes",
				},
			},
			Provider:            "test-resources",
			Resources:           resourceSlice,
			ComponentReferences: componentSlice,
		},
	}

	data, err := json.MarshalIndent(ocm, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OCM component: %w", err)
	}

	cr := components.ComponentReferenceFromNameAndVersion(g.config.RootComponent.ComponentName, g.config.RootComponent.Version)
	return os.WriteFile(cr.ToFilename(targetDir), data, 0644)
}

func (g *generator) addKubernetesVersionResources(resources []map[string]any) []map[string]any {
	for _, version := range g.config.KubernetesVersions {
		resources = append(resources, g.createKubernetesVersionHyperkube(version))
		resources = append(resources, g.createKubernetesVersionKubeapiserver(version))
		resources = append(resources, g.createKubernetesVersionKubeControllerManager(version))
		resources = append(resources, g.createKubernetesVersionKubeScheduler(version))
		resources = append(resources, g.createKubernetesVersionKubeProxy(version))
	}
	return resources
}

func (g *generator) createKubernetesVersionHyperkube(version string) map[string]any {
	return g.createKubernetesVersionResource("hyperkube", version, "europe-docker.pkg.dev/gardener-project/releases/hyperkube")
}

func (g *generator) createKubernetesVersionKubeapiserver(version string) map[string]any {
	return g.createKubernetesVersionResource("kube-apiserver", version, "registry.k8s.io/kube-apiserver")
}

func (g *generator) createKubernetesVersionKubeControllerManager(version string) map[string]any {
	return g.createKubernetesVersionResource("kube-controller-manager", version, "registry.k8s.io/kube-controller-manager")
}

func (g *generator) createKubernetesVersionKubeScheduler(version string) map[string]any {
	return g.createKubernetesVersionResource("kube-scheduler", version, "registry.k8s.io/kube-scheduler")
}

func (g *generator) createKubernetesVersionKubeProxy(version string) map[string]any {
	return g.createKubernetesVersionResource("kube-proxy", version, "registry.k8s.io/kube-proxy")
}

func (g *generator) createKubernetesVersionResource(name, version, origin string) map[string]any {
	return map[string]any{
		"name":    name,
		"version": version,
		"labels": []map[string]any{
			{
				"name":  components.LabelNameImageVectorName,
				"value": origin,
			},
			{
				"name":  components.LabelNameOriginalRef,
				"value": fmt.Sprintf("%s:v%s", origin, version),
			},
		},
		"extraIdentity": map[string]any{
			"version": version,
		},
		"type":     "ociImage",
		"relation": "external",
		"access": map[string]any{
			"imageReference": fmt.Sprintf("%s/%s:v%s@%s",
				g.config.TargetRepositoryURL, strings.ReplaceAll(origin, ".", "_"), version, calcSampleSHA256(name, version)),
			"type": "ociRegistry",
		},
	}
}

func (g *generator) addExternalResources(resources []map[string]any) []map[string]any {
	for _, extRes := range g.config.ExternalResources {
		resources = append(resources, g.createExternalResource(extRes))
	}
	return resources
}

func (g *generator) createExternalResource(externalResource ExternalResource) map[string]any {
	return map[string]any{
		"name":    externalResource.Name,
		"version": externalResource.Version,
		"labels": []map[string]any{
			{
				"name":  components.LabelNameOriginalRef,
				"value": fmt.Sprintf("%s:%s", externalResource.RepositoryURL, externalResource.Version),
			},
		},
		"type":     "ociImage",
		"relation": "external",
		"access": map[string]any{
			"imageReference": fmt.Sprintf("%s/%s:%s@%s",
				g.config.TargetRepositoryURL, strings.ReplaceAll(externalResource.RepositoryURL, ".", "_"), externalResource.Version, calcSampleSHA256(externalResource.Name, externalResource.Version)),
			"type": "ociRegistry",
		},
	}
}

func writeComponentDescriptor(cr components.ComponentReference, dv2 *descriptorv2.Descriptor, targetDir string) error {
	filename := cr.ToFilename(targetDir)
	data, err := json.MarshalIndent(dv2, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write component descriptor to file %s: %w", filename, err)
	}
	return nil
}

func calcSampleSHA256(name, version string) string {
	hash := sha256.Sum256([]byte(name + ":" + version))
	return "sha256:" + hex.EncodeToString(hash[:])
}
