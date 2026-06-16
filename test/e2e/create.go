// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	fluxv1 "github.com/fluxcd/kustomize-controller/api/v1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	operatorv1alpha1 "github.com/gardener/gardener/pkg/apis/operator/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	operatorclient "github.com/gardener/gardener/pkg/operator/client"
	. "github.com/gardener/gardener/test/e2e/gardener"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	componentbaseconfigv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/registry"
)

var _ = Describe("Create Gardener Landscape", Label("Garden", "default"), Ordered, func() {
	var (
		runtimeClusterClient kubernetes.Interface

		s *GardenContext
	)

	It("Should generate Gardener components", func(ctx SpecContext) {
		git := newGitCommand(BasePath)
		forgejo := newForgejoCommand(GitServerURL, GitUserName, GitUserPassword)

		// patch GLK config
		config := &v1alpha1.LandscapeKitConfiguration{}
		configBytes, err := os.ReadFile(ConfigPath) // #nosec G304 -- controlled by test.
		Expect(err).NotTo(HaveOccurred())

		Expect(yaml.Unmarshal(configBytes, config)).To(Succeed())

		Expect(config.Components).NotTo(BeNil())
		// activate all components
		config.Components.Include = nil

		configBytes, err = yaml.Marshal(config)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(ConfigPath, configBytes, 0600)).To(Succeed())

		By("Committing config changes")
		branchName := fmt.Sprintf("e2e/generate-%d", time.Now().Unix())
		git.checkout(branchName, true)
		DeferCleanup(func() {
			git.checkout("main", false)
			git.deleteBranch(branchName)
		})

		git.add(".")
		git.commit("Update components")

		By("Opening PR to trigger base generation action")
		baseBranch, basePRIndex := forgejo.pushAndCreatePR(branchName, GLKBaseRepoName, BasePath)

		By("Waiting for base generation action to succeed")
		forgejo.waitForActionSuccess(ctx, GLKBaseRepoName, baseBranch, git.headCommit())

		By("Verifying action committed generated content")
		forgejo.verifyActionCommit(GLKBaseRepoName, baseBranch)

		By("Merging base PR")
		forgejo.mergePR(GLKBaseRepoName, basePRIndex)
	}, SpecTimeout(20*time.Minute))

	It("Should prepare the Garden resource", func(ctx SpecContext) {
		git := newGitCommand(LandscapePath)
		forgejo := newForgejoCommand(GitServerURL, GitUserName, GitUserPassword)

		By("Updating base submodule in landscape")
		branchName := fmt.Sprintf("e2e/generate-%d", time.Now().Unix())
		git.checkout(branchName, true)

		DeferCleanup(func() {
			git.checkout("main", false)
			git.deleteBranch(branchName)
		})

		git.submoduleUpdate()

		By("Committing base submodule update")
		git.add("base")
		git.commit("Update base submodule")

		By("Opening PR to trigger landscape generation action")
		landscapeBranch, landscapePRIndex := forgejo.pushAndCreatePR(branchName, GLKLandscapeRepoName, LandscapePath)

		By("Waiting for landscape generation action to succeed")
		forgejo.waitForActionSuccess(ctx, GLKLandscapeRepoName, landscapeBranch, git.headCommit())

		By("Verifying action committed generated content")
		forgejo.verifyActionCommit(GLKLandscapeRepoName, landscapeBranch)

		By("Configuring the Garden resource")
		git.rebase(branchName)

		gardenComponentDir := filepath.Join(LandscapePath, "components", "gardener", "garden")
		gardenYamlPath := filepath.Join(gardenComponentDir, "garden.yaml")
		gardenBytes, err := os.ReadFile(gardenYamlPath) // #nosec G304 -- controlled by test.
		Expect(err).NotTo(HaveOccurred())

		// Unmarshal into map to preserve structure and avoid empty fields
		gardenMap := make(map[string]any)
		Expect(yaml.Unmarshal(gardenBytes, &gardenMap)).To(Succeed())

		// Get spec
		spec, ok := gardenMap["spec"].(map[string]any)
		Expect(ok).To(BeTrue(), "spec field should exist")

		// Set DNS
		spec["dns"] = map[string]any{
			"providers": []map[string]any{
				{
					"name": "primary",
					"type": "local",
					"secretRef": map[string]any{
						"name": "garden-dns-local",
					},
				},
			},
		}

		// Set runtimeCluster
		spec["runtimeCluster"] = map[string]any{
			"ingress": map[string]any{
				"controller": map[string]any{
					"kind": "nginx",
				},
				"domains": []map[string]any{
					{
						"name":     "ingress.runtime-garden.local.gardener.cloud",
						"provider": "primary",
					},
				},
			},
			"networking": map[string]any{
				"ipFamilies": []string{"IPv4"},
				"pods":       []string{"10.1.0.0/16"},
				"nodes":      []string{"172.18.0.0/24"},
				"services":   []string{"10.2.0.0/16"},
			},
		}

		// Set virtualGarden
		spec["virtualCluster"] = map[string]any{
			"dns": map[string]any{
				"domains": []map[string]any{
					{
						"name":     "virtual-garden.local.gardener.cloud",
						"provider": "primary",
					},
				},
			},
			"gardener": map[string]any{
				"clusterIdentity": "test-landscape-123456",
			},
		}

		// Marshal back to YAML
		patchedGardenBytes, err := yaml.Marshal(gardenMap)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(gardenYamlPath, patchedGardenBytes, 0600)).To(Succeed())

		By("Create DNS secret")
		secretYamlPath := filepath.Join(gardenComponentDir, "secret-garden-dns-local.yaml")
		secretYaml := `apiVersion: v1
kind: Secret
metadata:
  name: garden-dns-local
  namespace: garden
type: Opaque
data: {}
` // #nosec G101 -- No credential.
		Expect(os.WriteFile(secretYamlPath, []byte(secretYaml), 0600)).To(Succeed())

		By("Patch kustomization.yaml")
		kustomizationPath := filepath.Join(gardenComponentDir, "kustomization.yaml")
		kustomizationBytes, err := os.ReadFile(kustomizationPath) // #nosec G304 -- controlled by test.
		Expect(err).NotTo(HaveOccurred())

		kustomizationMap := make(map[string]any)
		Expect(yaml.Unmarshal(kustomizationBytes, &kustomizationMap)).To(Succeed())

		// Get or create resources array
		var resources []any
		if resourcesRaw, ok := kustomizationMap["resources"]; ok {
			if resourcesSlice, ok := resourcesRaw.([]any); ok {
				resources = resourcesSlice
			}
		}

		// Add secret-garden-dns-local.yaml if not present
		secretResource := "secret-garden-dns-local.yaml"
		found := false
		for _, r := range resources {
			if r == secretResource {
				found = true
				break
			}
		}
		if !found {
			resources = append(resources, secretResource)
		}
		kustomizationMap["resources"] = resources

		// Marshal back to YAML
		patchedKustomizationBytes, err := yaml.Marshal(kustomizationMap)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(kustomizationPath, patchedKustomizationBytes, 0600)).To(Succeed())

		By("Committing Garden changes")
		git.add("components")
		git.commit("Prepare garden component")
		git.push(branchName)

		By("Waiting for landscape generation action to succeed")
		forgejo.waitForActionSuccess(ctx, GLKLandscapeRepoName, landscapeBranch, git.headCommit())

		By("Verifying action committed generated content")
		forgejo.verifyActionCommit(GLKLandscapeRepoName, landscapeBranch)

		By("Merging landscape PR")
		forgejo.mergePR(GLKLandscapeRepoName, landscapePRIndex)
	}, SpecTimeout(20*time.Minute))

	It("Create Kubernetes client", func() {
		runtimeScheme := runtime.NewScheme()
		Expect(fluxv1.AddToScheme(runtimeScheme)).To(Succeed())
		Expect(operatorclient.AddRuntimeSchemeToScheme(runtimeScheme)).To(Succeed())

		var err error
		runtimeClusterClient, err = kubernetes.NewClientFromFile("", os.Getenv("KUBECONFIG"),
			kubernetes.WithClientOptions(client.Options{Scheme: runtimeScheme}),
			kubernetes.WithClientConnectionOptions(
				componentbaseconfigv1alpha1.ClientConnectionConfiguration{QPS: 100, Burst: 130}),
			kubernetes.WithAllowedUserFields([]string{kubernetes.AuthTokenFile}),
			kubernetes.WithDisabledCachedClient(),
		)
		Expect(err).ToNot(HaveOccurred())

		s = &GardenContext{}
		s.WithVirtualClusterClientSet(runtimeClusterClient)
	})

	It("Reconcile Garden", func(ctx SpecContext) {
		garden := &operatorv1alpha1.Garden{ObjectMeta: metav1.ObjectMeta{Name: "garden"}}
		Eventually(ctx, func(g Gomega) {
			g.Expect(runtimeClusterClient.Client().Get(ctx, client.ObjectKeyFromObject(garden), garden)).To(Succeed())
			g.Expect(garden.Status.LastOperation).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"State":    Equal(gardencorev1beta1.LastOperationStateSucceeded),
				"Progress": BeEquivalentTo(100),
			})))
		}).Should(Succeed())
	}, SpecTimeout(20*time.Minute))

	It("Ensure that the configured operator extensions have been installed", func(ctx SpecContext) {
		var (
			extOps             operatorv1alpha1.ExtensionList
			expectedExtensions []types.GomegaMatcher
			extensionNames     = []string{"provider-local"}
		)

		// Iterate over all components and identify extensions
		for _, newComponent := range registry.ComponentList {
			component := newComponent()
			pkgPath := reflect.TypeOf(component).Elem().PkgPath()

			// Consider the component as an extension if the package path contains "gardener-extensions"
			if strings.Contains(pkgPath, "gardener-extensions") {
				extensionNames = append(extensionNames, component.Name())
			}
		}

		// Construct the expected extensions matchers based on the identified extension names
		for _, extension := range extensionNames {
			expectedExtensions = append(expectedExtensions, MatchFields(IgnoreExtras, Fields{
				"ObjectMeta": MatchFields(IgnoreExtras, Fields{
					"Name": Equal(extension),
				}),
				"Status": MatchFields(IgnoreExtras, Fields{
					"Conditions": ContainElement(MatchFields(IgnoreExtras, Fields{
						"Type":   Equal(operatorv1alpha1.ExtensionInstalled),
						"Status": BeEquivalentTo("True"),
					})),
				}),
			}))
		}

		Eventually(ctx, func(g Gomega) {
			g.Eventually(s.VirtualClusterKomega.List(&extOps)).To(Succeed())
			g.Expect(extOps.Items).To(ConsistOf(expectedExtensions))
		}).Should(Succeed())
	})

	It("Ensure that all Flux Kustomizations have been reconciled successfully", func(ctx SpecContext) {
		Eventually(ctx, func(g Gomega) {
			var ksList fluxv1.KustomizationList
			g.Expect(runtimeClusterClient.Client().List(ctx, &ksList)).To(Succeed())
			g.Expect(ksList.Items).ToNot(BeEmpty())

			for _, ks := range ksList.Items {
				readyCond := apimeta.FindStatusCondition(ks.Status.Conditions, fluxmeta.ReadyCondition)
				if !g.Expect(readyCond).ToNot(BeNil(),
					"Kustomization %s/%s has no Ready condition", ks.Namespace, ks.Name) {
					continue
				}
				g.Expect(readyCond.Status).To(Equal(metav1.ConditionTrue),
					"Kustomization %s/%s is not ready: %s: %s", ks.Namespace, ks.Name,
					readyCond.Reason, readyCond.Message)
			}
		}).Should(Succeed())
	})
})
