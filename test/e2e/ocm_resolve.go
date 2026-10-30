// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	descriptorruntime "ocm.software/open-component-model/bindings/go/descriptor/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/ociaccess"
	"github.com/gardener/gardener-landscape-kit/pkg/registry"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

// These synthetic components are pushed into the local OCI registry and resolved by `glk resolve ocm`.
// The names are intentionally outside the embedded default component vector so the resolve output depends
// solely on the descriptors we seed here and the custom-component marker files we drop.
const (
	rootComponentName  = "example.com/glk-e2e/root"
	leafAComponentName = "example.com/glk-e2e/leaf-a"
	leafBComponentName = "example.com/glk-e2e/leaf-b"
	syntheticVersion   = "v0.1.0"
)

var _ = Describe("Resolve OCM components", Label("Garden", "default", "ocm"), func() {
	It("Should push a synthetic component vector and resolve it via `glk resolve ocm`", func(ctx SpecContext) {
		log := logf.Log

		By("Seeding synthetic OCM components into the local registry")
		repo, err := ociaccess.NewRepoAccess(RegistryURL)
		Expect(err).NotTo(HaveOccurred())

		// Two leaf components referenced by the root component. They carry no resources, so resolve has no
		// local blobs to fetch and the assertions depend only on the component graph.
		leafA := newDescriptor(leafAComponentName, syntheticVersion)
		leafB := newDescriptor(leafBComponentName, syntheticVersion)
		Expect(repo.AddComponentVersion(ctx, leafA)).To(Succeed())
		Expect(repo.AddComponentVersion(ctx, leafB)).To(Succeed())

		root := newDescriptor(rootComponentName, syntheticVersion)
		root.Component.References = []descriptorruntime.Reference{
			newReference("leaf-a", leafAComponentName, syntheticVersion),
			newReference("leaf-b", leafBComponentName, syntheticVersion),
		}
		Expect(repo.AddComponentVersion(ctx, root)).To(Succeed())

		By("Writing a GLK config and custom-component markers pointing at the seeded components")
		landscapeDir := GinkgoT().TempDir()
		outputDir := GinkgoT().TempDir()

		cfg := &configv1alpha1.LandscapeKitConfiguration{
			OCM: &configv1alpha1.OCMConfig{
				Repositories: []string{RegistryURL},
				RootComponent: configv1alpha1.OCMComponent{
					Name:    rootComponentName,
					Version: syntheticVersion,
				},
				IgnoreMissingComponents: new(true),
			},
		}

		// Create a custom component for every component we expect to appear in components.yaml. Without a
		// marker (and since these names are not in the embedded default vector) GetGLKComponents would skip them.
		for name, dir := range map[string]string{
			rootComponentName:  "root",
			leafAComponentName: "leaf-a",
			leafBComponentName: "leaf-b",
		} {
			markerDir := filepath.Join(landscapeDir, dir)
			Expect(os.MkdirAll(markerDir, 0700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(markerDir, registry.CustomComponentNameFilename), []byte(name), 0600)).To(Succeed())
		}

		By("Running resolve ocm")
		Expect(ocm.ResolveOCMComponents(log, cfg, landscapeDir, outputDir, 1, false)).To(Succeed())

		By("Verifying the component vector was written")
		vectorBytes, err := os.ReadFile(filepath.Join(landscapeDir, componentvector.ComponentVectorFilename)) // #nosec G304 -- controlled by test.
		Expect(err).NotTo(HaveOccurred())
		vector := &componentvector.Components{}
		Expect(yaml.Unmarshal(vectorBytes, vector)).To(Succeed())

		vectorByName := map[string]string{}
		for _, c := range vector.Components {
			vectorByName[c.Name] = c.Version
		}
		Expect(vectorByName).To(HaveKeyWithValue(rootComponentName, syntheticVersion))
		Expect(vectorByName).To(HaveKeyWithValue(leafAComponentName, syntheticVersion))
		Expect(vectorByName).To(HaveKeyWithValue(leafBComponentName, syntheticVersion))

		By("Verifying the component list contains the full resolved graph")
		listBytes, err := os.ReadFile(filepath.Join(outputDir, "component-list.yaml")) // #nosec G304 -- controlled by test.
		Expect(err).NotTo(HaveOccurred())
		Expect(string(listBytes)).To(SatisfyAll(
			ContainSubstring(rootComponentName),
			ContainSubstring(leafAComponentName),
			ContainSubstring(leafBComponentName),
		))

		By("Verifying a descriptor file was written per resolved component")
		descriptorDir := path.Join(outputDir, "descriptors")
		entries, err := os.ReadDir(descriptorDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(3), "expected one descriptor file per resolved component")
	})
})

// newDescriptor builds a minimal v2-schema component descriptor with no resources or references.
func newDescriptor(name, version string) *descriptorruntime.Descriptor {
	return &descriptorruntime.Descriptor{
		Meta: descriptorruntime.Meta{Version: "v2"},
		Component: descriptorruntime.Component{
			ComponentMeta: descriptorruntime.ComponentMeta{
				ObjectMeta: descriptorruntime.ObjectMeta{
					Name:    name,
					Version: version,
				},
			},
			Provider: descriptorruntime.Provider{Name: "glk-e2e"},
		},
	}
}

// newReference builds a component reference. localName is the reference's own name within the component,
// component/version identify the referenced component version.
func newReference(localName, component, version string) descriptorruntime.Reference {
	return descriptorruntime.Reference{
		Name:      localName,
		Version:   version,
		Component: component,
	}
}
