// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components_test

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	descriptorruntime "ocm.software/open-component-model/bindings/go/descriptor/runtime"
	descriptorv2 "ocm.software/open-component-model/bindings/go/descriptor/v2"

	"github.com/gardener/gardener-landscape-kit/pkg/ocm/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/ociaccess"
)

const resourcesDir = "../../../test/ocm/testdata"

const (
	refShootCertService = components.ComponentReference("github.com/gardener/gardener-extension-shoot-cert-service:v1.53.0")
	refGardener         = components.ComponentReference("github.com/gardener/gardener:v1.128.3")
	refRoot             = components.ComponentReference("example.com/kubernetes-root-example:0.1499.0")
)

var _ = Describe("Components", func() {
	var (
		c *components.Components

		loadDescriptor = func(cref components.ComponentReference) *descriptorruntime.Descriptor {
			filename := cref.ToFilename(resourcesDir)
			data, err := os.ReadFile(filename)
			Expect(err).NotTo(HaveOccurred())
			dv2 := &descriptorv2.Descriptor{}
			Expect(json.Unmarshal(data, dv2)).To(Succeed(), filename)
			desc, err := descriptorruntime.ConvertFromV2(dv2)
			Expect(err).NotTo(HaveOccurred())
			return desc
		}

		loadWithDep = func(levels int, roots ...components.ComponentReference) {
			Expect(levels).To(BeNumerically(">=", 0))
			loadWithDepRecursive(c, loadDescriptor, levels, roots...)
		}
	)
	BeforeEach(func() {
		c = components.NewComponents()
	})

	It("should produce correct image vector for shoot-cert-service", func() {
		loadWithDep(1, refShootCertService)
		Expect(c.ComponentsCount()).To(Equal(1))
		roots := c.GetRootComponents()
		Expect(roots).To(ConsistOf(refShootCertService))

		By("resolved references")
		imageVector, err := c.GetImageVector(refShootCertService, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(imageVector).To(HaveLen(1))
		Expect(imageVector, "").To(ConsistOf(
			imagevector.ImageSource{
				Name:       "cert-management",
				Repository: ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/cert-controller-manager"),
				Tag:        ptr.To("v0.17.7@sha256:6f55f7bf5a6498dc0d138e5cde33eb39a090ceeee1fe80647008cb8e04676d8c"),
				Version:    ptr.To("v0.17.7"),
			}))

		By("original references")
		imageVector, err = c.GetImageVector(refShootCertService, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(imageVector).To(HaveLen(1))
		Expect(imageVector, "").To(ConsistOf(
			imagevector.ImageSource{
				Name:       "cert-management",
				Repository: ptr.To("europe-docker.pkg.dev/gardener-project/releases/cert-controller-manager"),
				Tag:        ptr.To("v0.17.7"),
				Version:    ptr.To("v0.17.7"),
			}))

		By("check resources for ociImage, helmChart and helmchart-imagemap types")
		resources := c.GetResources(refShootCertService)
		Expect(resources).To(HaveLen(4))
		Expect(resources).To(ContainElements(
			components.Resource{
				Name:    "gardener-extension-shoot-cert-service",
				Version: "v1.53.0",
				Type:    "ociImage",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/extensions/shoot-cert-service:v1.53.0@sha256:73d1016d52140655c444d1189ad90826a81eb2418126fbbae365b9c9ee0ddcfd",
			},
			components.Resource{
				Name:    "shoot-cert-service",
				Version: "v1.53.0",
				Type:    "helmChart/v1",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/charts/gardener/extensions/shoot-cert-service:v1.53.0@sha256:1236fb136e6951d2c438d6ae315721425f866fc494e2d811582b43c0a579e90e",
			},
			components.Resource{
				Name:    "shoot-cert-service",
				Version: "v1.53.0",
				Type:    "helmchart-imagemap",
				Value:   "fake content",
			},
			components.Resource{
				Name:    "cert-management",
				Version: "v0.17.7",
				Type:    "ociImage",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/cert-controller-manager:v0.17.7@sha256:6f55f7bf5a6498dc0d138e5cde33eb39a090ceeee1fe80647008cb8e04676d8c",
			},
		))
	})

	It("should produce correct image vector for gardener/gardener for rewritten images in OCM components", func() {
		loadWithDep(3, refRoot)
		Expect(c.ComponentsCount()).To(Equal(23))
		roots := c.GetRootComponents()
		Expect(roots).To(ConsistOf(refRoot))
		imageVector, err := c.GetImageVector(refGardener, false)
		Expect(err).NotTo(HaveOccurred())

		By("check target versions / referenced resources")
		Expect(filterByNamePrefix(imageVector, "cluster-autoscaler")).To(HaveLen(5))
		Expect(filterByNamePrefix(imageVector, "cluster-autoscaler")).To(ContainElements(
			imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To("1.31.x"),
				Repository:    ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.31.0@sha256:501e13c151f6d655bc38bc6f9c08e4f327f074c4a958e9e0a36f1d0819e091ff"),
				Version:       ptr.To("v1.31.0"),
			},
			imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To("1.32.x"),
				Repository:    ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.32.1@sha256:8c7080fa391ba569c67d4f2470dad4de0f6f1a83aa1a539f7df6fadad1fb6240"),
				Version:       ptr.To("v1.32.1"),
			}, imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To(">= 1.33"),
				Repository:    ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.33.0@sha256:4932021e763b81c2679dda43af369619976a1a177726c8a507aa9003e84c18e3"),
				Version:       ptr.To("v1.33.0"),
			}))

		By("check gardener images")
		Expect(filterByNamePrefix(imageVector, "gardener-")).To(ContainElements(
			imagevector.ImageSource{
				Name:       "gardener-admission-controller",
				Repository: ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/admission-controller"),
				Tag:        ptr.To("v1.128.3@sha256:1b4d5332ebe78b9e4970361230ec043aa967ea70ea6e53b2c3a8538e2e4a496d"),
				Version:    ptr.To("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-apiserver",
				Repository: ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/apiserver"),
				Tag:        ptr.To("v1.128.3@sha256:d8679b8760f77e540c28d1e32e938b082d3dfdd3b7603666d474726940bb8942"),
				Version:    ptr.To("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-controller-manager",
				Repository: ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/controller-manager"),
				Tag:        ptr.To("v1.128.3@sha256:f1509f9f7d43902d319a87757612bd369439739fc4381ef77698d3e5447896f7"),
				Version:    ptr.To("v1.128.3"),
			}))

		By("check images from a referenced components")
		Expect(filterByNamePrefix(imageVector, "vpn-")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "vpn-client",
				Repository: ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/vpn-client"),
				Tag:        ptr.To("0.41.1@sha256:1871708ac9d09183b11d4f6d55548052db89e075faa0eddb1eb8bd5bb8ee956f"),
				Version:    ptr.To("0.41.1"),
			},
			imagevector.ImageSource{
				Name:       "vpn-server",
				Repository: ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/vpn-server"),
				Tag:        ptr.To("0.41.1@sha256:25b166baba9426d77929dc6cd38ab1e3d6dd846e3e9f847365dd216a1d9dd1ab"),
				Version:    ptr.To("0.41.1"),
			}))

		By("check images referenced from Kubernetes component by label `imagevector.gardener.cloud/images`")
		Expect(filterByNamePrefix(imageVector, "hyperkube")).To(ContainElements(
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.32.6@sha256:7ea3d70fa985db8b1577cbae3af775af304eec06432518cc4e69d1fb5e48459f"),
				Version:       ptr.To("1.32.6"),
				TargetVersion: ptr.To("1.32.6"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.32.8@sha256:6cd7f4cf3eaabfd34eb0c500b92c1bca0c88e23b851cca6e076f0e2b6f3a18e5"),
				Version:       ptr.To("1.32.8"),
				TargetVersion: ptr.To("1.32.8"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.31.10@sha256:54ba6d6336e7ce9585db3ec3bba789bfc0980e1e514b03eec6d69d193bd15c55"),
				Version:       ptr.To("1.31.10"),
				TargetVersion: ptr.To("1.31.10"),
			}))
		kubernetesVersionCount := 5
		Expect(countImagesByName(imageVector, "hyperkube")).To(Equal(kubernetesVersionCount))
		Expect(countImagesByName(imageVector, "kube-apiserver")).To(Equal(kubernetesVersionCount))
		Expect(countImagesByName(imageVector, "kube-controller-manager")).To(Equal(kubernetesVersionCount))
		Expect(countImagesByName(imageVector, "kube-proxy")).To(Equal(kubernetesVersionCount))
		Expect(countImagesByName(imageVector, "kube-scheduler")).To(Equal(kubernetesVersionCount))

		By("check resolution if resourceID == component name")
		Expect(filterByNamePrefix(imageVector, "ingress-default-backend")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "ingress-default-backend",
				Repository: ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/ingress-default-backend"),
				Tag:        ptr.To("0.24.0@sha256:4c9b83da61f44d255d128c56421450de95fbdd2c9f77d8ff8e9a15b8ca41a924"),
				Version:    ptr.To("0.24.0"),
			}))

		By("check resolution if resourceID != component name")
		Expect(filterByNamePrefix(imageVector, "apiserver-proxy-sidecar")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "apiserver-proxy-sidecar",
				Repository: ptr.To("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/apiserver-proxy"),
				Tag:        ptr.To("v0.19.0@sha256:18fd91eb57eef02cfb2c0f9943deeefc4e44c4a1577863d808250538af8a6e03"),
				Version:    ptr.To("v0.19.0"),
			}))

		By("check resources for ociImage, helmChart and helmchart-imagemap types")
		resources := c.GetResources(refGardener)
		Expect(resources).To(HaveLen(52))
		Expect(resources).To(ContainElements(
			components.Resource{
				Name:    "resource-manager",
				Version: "v1.128.3",
				Type:    "helmChart/v1",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/charts/gardener/resource-manager:v1.128.3@sha256:ce1e87bde456347a364035314092ff699a7522bb3f90c65a2f21a88915ad4e7e",
			},
			components.Resource{
				Name:    "resource-manager",
				Version: "v1.128.3",
				Type:    "helmchart-imagemap",
				Value:   "fake content",
			},
			components.Resource{
				Name:    "gardenlet",
				Version: "v1.128.3",
				Type:    "helmChart/v1",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/charts/gardener/gardenlet:v1.128.3@sha256:a5880e6933465e58536fdfb381acee013905ecd6888d94f0d484dff081ab0b11",
			},
			components.Resource{
				Name:    "gardenlet",
				Version: "v1.128.3",
				Type:    "helmchart-imagemap",
				Value:   "fake content",
			},
			components.Resource{
				Name:    "gardenlet",
				Version: "v1.128.3",
				Type:    "ociImage",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/gardenlet:v1.128.3@sha256:a5880e6933465e58536fdfb381acee013905ecd6888d94f0d484dff081ab0b11",
			},
		))
	})

	It("should produce correct image vector for gardener/gardener with original reference", func() {
		loadWithDep(3, refRoot)
		Expect(c.ComponentsCount()).To(Equal(23))
		roots := c.GetRootComponents()
		Expect(roots).To(ConsistOf(refRoot))
		imageVector, err := c.GetImageVector(refGardener, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(filterByNamePrefix(imageVector, "cluster-autoscaler")).To(HaveLen(5))
		Expect(filterByNamePrefix(imageVector, "cluster-autoscaler")).To(ContainElements(
			imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To("1.31.x"),
				Repository:    ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.31.0"),
				Version:       ptr.To("v1.31.0"),
			},
			imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To("1.32.x"),
				Repository:    ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.32.1"),
				Version:       ptr.To("v1.32.1"),
			}, imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To(">= 1.33"),
				Repository:    ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.33.0"),
				Version:       ptr.To("v1.33.0"),
			}))
		Expect(filterByNamePrefix(imageVector, "gardener-")).To(ContainElements(
			imagevector.ImageSource{
				Name:       "gardener-admission-controller",
				Repository: ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/admission-controller"),
				Tag:        ptr.To("v1.128.3"),
				Version:    ptr.To("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-apiserver",
				Repository: ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/apiserver"),
				Tag:        ptr.To("v1.128.3"),
				Version:    ptr.To("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-controller-manager",
				Repository: ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/controller-manager"),
				Tag:        ptr.To("v1.128.3"),
				Version:    ptr.To("v1.128.3"),
			}))

		Expect(filterByNamePrefix(imageVector, "vpn-")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "vpn-client",
				Repository: ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/vpn-client"),
				Tag:        ptr.To("0.41.1"),
				Version:    ptr.To("0.41.1"),
			},
			imagevector.ImageSource{
				Name:       "vpn-server",
				Repository: ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/vpn-server"),
				Tag:        ptr.To("0.41.1"),
				Version:    ptr.To("0.41.1"),
			}))

		Expect(filterByNamePrefix(imageVector, "hyperkube")).To(ContainElements(
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("europe-docker.pkg.dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.32.6"),
				Version:       ptr.To("1.32.6"),
				TargetVersion: ptr.To("1.32.6"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("europe-docker.pkg.dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.32.8"),
				Version:       ptr.To("1.32.8"),
				TargetVersion: ptr.To("1.32.8"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("europe-docker.pkg.dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.31.10"),
				Version:       ptr.To("1.31.10"),
				TargetVersion: ptr.To("1.31.10"),
			}))
		kubernetesVersionCount := 5
		Expect(countImagesByName(imageVector, "hyperkube")).To(Equal(kubernetesVersionCount))
		Expect(countImagesByName(imageVector, "kube-apiserver")).To(Equal(kubernetesVersionCount))
		Expect(countImagesByName(imageVector, "kube-controller-manager")).To(Equal(kubernetesVersionCount))
		Expect(countImagesByName(imageVector, "kube-proxy")).To(Equal(kubernetesVersionCount))
		Expect(countImagesByName(imageVector, "kube-scheduler")).To(Equal(kubernetesVersionCount))

		Expect(filterByNamePrefix(imageVector, "ingress-default-backend")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "ingress-default-backend",
				Repository: ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/ingress-default-backend"),
				Tag:        ptr.To("0.24.0"),
				Version:    ptr.To("0.24.0"),
			}))

		// check resolution by resourceID
		Expect(filterByNamePrefix(imageVector, "apiserver-proxy-sidecar")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "apiserver-proxy-sidecar",
				Repository: ptr.To("europe-docker.pkg.dev/gardener-project/releases/gardener/apiserver-proxy"),
				Tag:        ptr.To("v0.19.0"),
				Version:    ptr.To("v0.19.0"),
			}))
	})
})

func countImagesByName(images []imagevector.ImageSource, name string) int {
	var count int
	for _, image := range images {
		if image.Name == name {
			count++
		}
	}
	return count
}

func loadWithDepRecursive(c *components.Components, loadDescriptor func(cref components.ComponentReference) *descriptorruntime.Descriptor, levels int, roots ...components.ComponentReference) {
	Expect(levels).To(BeNumerically(">=", 0))
	for _, root := range roots {
		desc := loadDescriptor(root)
		blobs := addFakeLocalBlobs(desc)
		deps, err := c.AddComponentDependencies(desc, blobs)
		Expect(err).NotTo(HaveOccurred())
		if levels > 0 && len(deps) > 0 {
			loadWithDepRecursive(c, loadDescriptor, levels-1, deps...)
		}
	}
}

func addFakeLocalBlobs(desc *descriptorruntime.Descriptor) components.Blobs {
	var blobs components.Blobs
	for _, res := range desc.Component.Resources {
		if res.Type == components.ResourceTypeHelmChartImageMap {
			// simulate that we have the ociImage blobs available locally
			if blobs == nil {
				blobs = components.Blobs{}
			}
			blobs[ociaccess.ResourceToBlobKey(res)] = []byte("fake content")
		}
	}
	return blobs
}

func filterByNamePrefix(images []imagevector.ImageSource, namePrefix string) []imagevector.ImageSource {
	var result []imagevector.ImageSource
	for _, img := range images {
		if strings.HasPrefix(img.Name, namePrefix) {
			result = append(result, img)
		}
	}
	return result
}
