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

const resourcesDir = "../../../test/ocm/resources"

const (
	refShootCertService = components.ComponentReference("github.com/gardener/gardener-extension-shoot-cert-service:v1.53.0")
	refGardener         = components.ComponentReference("github.com/gardener/gardener:v1.128.3")
	refLSS              = components.ComponentReference("example.com/kubernetes/landscape-setup:0.6914.2")
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
		imageVector, err := c.GetImageVector(refShootCertService, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(imageVector).To(HaveLen(1))
		Expect(imageVector, "").To(ConsistOf(
			imagevector.ImageSource{
				Name:       "cert-management",
				Repository: ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/cert-controller-manager"),
				Tag:        ptr.To("v0.17.7@sha256:958007e219e1cafe9c86ca7c438f60a4e44e0924d080803cf7a9cfe692a5ce55"),
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
		resources := c.GetResources(refShootCertService)
		Expect(resources).To(HaveLen(4))
		Expect(resources).To(ContainElements(
			components.Resource{
				Name:    "gardener-extension-shoot-cert-service",
				Version: "v1.53.0",
				Type:    "ociImage",
				Value:   "repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/extensions/shoot-cert-service:v1.53.0@sha256:5d5231cc5c2782c20a70237e2343844c4f11f8bd7761cf39341a119bb3d2b47c",
			},
			components.Resource{
				Name:    "shoot-cert-service",
				Version: "v1.53.0",
				Type:    "helmChart/v1",
				Value:   "repo.example.com/staging/europe-docker-pkg-dev/gardener-project/releases/charts/gardener/extensions/shoot-cert-service:v1.53.0@sha256:e566cf677cca3b8d6c8dd1cbd48ebfa658e9683d7982913873697d337dd018bd",
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
				Value:   "repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/cert-controller-manager:v0.17.7@sha256:958007e219e1cafe9c86ca7c438f60a4e44e0924d080803cf7a9cfe692a5ce55",
			},
		))
	})

	It("should produce correct image vector for gardener/gardener", func() {
		loadWithDep(3, refLSS)
		Expect(c.ComponentsCount()).To(Equal(75))
		roots := c.GetRootComponents()
		Expect(roots).To(ConsistOf(refLSS))
		imageVector, err := c.GetImageVector(refGardener, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(filterByNamePrefix(imageVector, "cluster-autoscaler")).To(HaveLen(5))
		Expect(filterByNamePrefix(imageVector, "cluster-autoscaler")).To(ContainElements(
			imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To("1.31.x"),
				Repository:    ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.31.0@sha256:739cddcb7f1803c2172ca9c87e2debd83b4002187f3a585b7b5d0eda2613e565"),
				Version:       ptr.To("v1.31.0"),
			},
			imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To("1.32.x"),
				Repository:    ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.32.1@sha256:4b3622812f05a01c9eda6520193d04a86a16cfd7736e34aca51fd3e0df8c3ad4"),
				Version:       ptr.To("v1.32.1"),
			}, imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: ptr.To(">= 1.33"),
				Repository:    ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           ptr.To("v1.33.0@sha256:2b569c1c016b09ac16204f6ac80eef477828b9771422051236e45c61681e30a4"),
				Version:       ptr.To("v1.33.0"),
			}))
		Expect(filterByNamePrefix(imageVector, "gardener-")).To(ContainElements(
			imagevector.ImageSource{
				Name:       "gardener-admission-controller",
				Repository: ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/admission-controller"),
				Tag:        ptr.To("v1.128.3@sha256:2b52910820b68b1d1d48733073576b9805fcb5e9d1e7d23954dd5e9cd0644b27"),
				Version:    ptr.To("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-apiserver",
				Repository: ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/apiserver"),
				Tag:        ptr.To("v1.128.3@sha256:6de833f0513d4ae4e49910eebfe925f45c9b19d7b1fa47bc3959ad914d2092bd"),
				Version:    ptr.To("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-controller-manager",
				Repository: ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/controller-manager"),
				Tag:        ptr.To("v1.128.3@sha256:4ce9a0ffb0210e25feb0090053d77ab4ef49a8219b6c5a3aa40760f4375761c4"),
				Version:    ptr.To("v1.128.3"),
			}))

		Expect(filterByNamePrefix(imageVector, "vpn-")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "vpn-client",
				Repository: ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/vpn-client"),
				Tag:        ptr.To("0.41.1@sha256:b9c8192128008a59498e9a3f117f9c7c2a4811ce7307e91f6776858168b2b6a3"),
				Version:    ptr.To("0.41.1"),
			},
			imagevector.ImageSource{
				Name:       "vpn-server",
				Repository: ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/vpn-server"),
				Tag:        ptr.To("0.41.1@sha256:fd0535996e332e977eeb2ddc0490bda74b23102e61e21db6ca59bc329a3a1762"),
				Version:    ptr.To("0.41.1"),
			}))

		Expect(filterByNamePrefix(imageVector, "hyperkube")).To(ContainElements(
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.32.2@sha256:24ca4f3e8b5730afadb3cf17589c929230b0342e6d316d11ccaec76540fed2c3"),
				Version:       ptr.To("1.32.2"),
				TargetVersion: ptr.To("1.32.2"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.32.3@sha256:1a253fcac9dbca37dbd2ca4143c55e3a69694abd9b860d48eae00ae22f1190d1"),
				Version:       ptr.To("1.32.3"),
				TargetVersion: ptr.To("1.32.3"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.31.6@sha256:49927020b5d27d8ea3bb3b5f8b66476a644cecc9ae0e978b3f1b13d00338f212"),
				Version:       ptr.To("1.31.6"),
				TargetVersion: ptr.To("1.31.6"),
			}))
		Expect(countImagesByName(imageVector, "hyperkube")).To(Equal(35))
		Expect(countImagesByName(imageVector, "kube-apiserver")).To(Equal(35))
		Expect(countImagesByName(imageVector, "kube-controller-manager")).To(Equal(35))
		Expect(countImagesByName(imageVector, "kube-proxy")).To(Equal(35))
		Expect(countImagesByName(imageVector, "kube-scheduler")).To(Equal(35))

		Expect(filterByNamePrefix(imageVector, "ingress-default-backend")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "ingress-default-backend",
				Repository: ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/ingress-default-backend"),
				Tag:        ptr.To("0.24.0@sha256:e8aaca280d906391372963d08d07d9f651d8206ef1259415ceaae6542d449ac9"),
				Version:    ptr.To("0.24.0"),
			}))

		// check resolution by resourceID
		Expect(filterByNamePrefix(imageVector, "apiserver-proxy-sidecar")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "apiserver-proxy-sidecar",
				Repository: ptr.To("repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/apiserver-proxy"),
				Tag:        ptr.To("v0.19.0@sha256:d3b9d9af4f420682fda692211a170a52702efb0ee3a3a12e1532f6ff10e2e764"),
				Version:    ptr.To("v0.19.0"),
			}))

		resources := c.GetResources(refGardener)
		Expect(resources).To(HaveLen(52))
		Expect(resources).To(ContainElements(
			components.Resource{
				Name:    "resource-manager",
				Version: "v1.128.3",
				Type:    "helmChart/v1",
				Value:   "repo.example.com/staging/europe-docker-pkg-dev/gardener-project/releases/charts/gardener/resource-manager:v1.128.3@sha256:9ae791300e5890f419e0a8cc203241c5767de20c39a1a57de47797bb6d67168b",
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
				Value:   "repo.example.com/staging/europe-docker-pkg-dev/gardener-project/releases/charts/gardener/gardenlet:v1.128.3@sha256:aeb1046888bf33d323bf8159c8062379e9c5221f9c644c99f9665252b6295bbb",
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
				Value:   "repo.example.com/staging/europe-docker_pkg_dev/gardener-project/releases/gardener/gardenlet:v1.128.3@sha256:2dda1465ef70cde4de2d5f9316326e245438acd38efe6515073ff73b496bfec8",
			},
		))
	})

	It("should produce correct image vector for gardener/gardener with original reference", func() {
		loadWithDep(3, refLSS)
		Expect(c.ComponentsCount()).To(Equal(75))
		roots := c.GetRootComponents()
		Expect(roots).To(ConsistOf(refLSS))
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
				Tag:           ptr.To("v1.32.2"),
				Version:       ptr.To("1.32.2"),
				TargetVersion: ptr.To("1.32.2"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("europe-docker.pkg.dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.32.3"),
				Version:       ptr.To("1.32.3"),
				TargetVersion: ptr.To("1.32.3"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    ptr.To("europe-docker.pkg.dev/gardener-project/releases/hyperkube"),
				Tag:           ptr.To("v1.31.6"),
				Version:       ptr.To("1.31.6"),
				TargetVersion: ptr.To("1.31.6"),
			}))
		Expect(countImagesByName(imageVector, "hyperkube")).To(Equal(35))
		Expect(countImagesByName(imageVector, "kube-apiserver")).To(Equal(35))
		Expect(countImagesByName(imageVector, "kube-controller-manager")).To(Equal(35))
		Expect(countImagesByName(imageVector, "kube-proxy")).To(Equal(35))
		Expect(countImagesByName(imageVector, "kube-scheduler")).To(Equal(35))

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
