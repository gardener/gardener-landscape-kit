// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package glk

import (
	"os"
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

	"github.com/gardener/gardener-landscape-kit/pkg/registry"
)

var _ = Describe("Garden Reconciliation", Label("Garden", "default"), Ordered, func() {
	var (
		runtimeClusterClient kubernetes.Interface

		s *GardenContext
	)

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
