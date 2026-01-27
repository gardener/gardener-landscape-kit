// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	descriptorruntime "ocm.software/open-component-model/bindings/go/descriptor/runtime"
	descriptorv2 "ocm.software/open-component-model/bindings/go/descriptor/v2"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/ocm/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/ociaccess"
)

const resourcesDir = "testdata"

const (
	refShootCertService               = components.ComponentReference("github.com/gardener/gardener-extension_extension-test:v1.53.0")
	refGardener                       = components.ComponentReference("github.com/gardener/gardener:v1.128.3")
	refRoot                           = components.ComponentReference("example.com/kubernetes-root-example:0.1499.0")
	gardenletHelmChartImageMapContent = `{"helmchartResource": {"name": "gardenlet"}, "imageMapping": [{"resource": {"name": "gardenlet"}, "repository": "image.repository", "tag": "image.tag"}]}`
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

	It("should produce correct image vector for extension-test", func() {
		loadWithDep(1, refShootCertService)
		Expect(c.ComponentsCount()).To(Equal(1))
		roots := c.GetRootComponents()
		Expect(roots).To(ConsistOf(refShootCertService))

		By("resolved references")
		imageVector, err := c.GetImageVector(refShootCertService, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(imageVector).To(HaveLen(1))
		Expect(imageVector).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "cert-management",
				Repository: new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/cert-controller-manager"),
				Tag:        new("v0.17.7@sha256:6f55f7bf5a6498dc0d138e5cde33eb39a090ceeee1fe80647008cb8e04676d8c"),
				Version:    new("v0.17.7"),
			}))

		By("original references")
		imageVector, err = c.GetImageVector(refShootCertService, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(imageVector).To(HaveLen(1))
		Expect(imageVector).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "cert-management",
				Repository: new("europe-docker.pkg.dev/gardener-project/releases/cert-controller-manager"),
				Tag:        new("v0.17.7"),
				Version:    new("v0.17.7"),
			}))

		By("check resources for ociImage, helmChart and helmchart-imagemap types")
		resources := c.GetResources(refShootCertService)
		Expect(resources).To(HaveLen(4))
		Expect(resources).To(ContainElements(
			components.Resource{
				Name:    "gardener-extension_extension-test",
				Version: "v1.53.0",
				Type:    "ociImage",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/extensions/extension-test:v1.53.0@sha256:73d1016d52140655c444d1189ad90826a81eb2418126fbbae365b9c9ee0ddcfd",
				Local:   true,
			},
			components.Resource{
				Name:    "extension-test",
				Version: "v1.53.0",
				Type:    "helmChart/v1",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/charts/gardener/extensions/extension-test:v1.53.0@sha256:1236fb136e6951d2c438d6ae315721425f866fc494e2d811582b43c0a579e90e",
			},
			components.Resource{
				Name:    "extension-test",
				Version: "v1.53.0",
				Type:    "helmchart-imagemap",
				Value:   "{\"helmchartResource\": {\"name\": \"extension-test\"}, \"imageMapping\": []}",
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
				TargetVersion: new("1.31.x"),
				Repository:    new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           new("v1.31.0@sha256:501e13c151f6d655bc38bc6f9c08e4f327f074c4a958e9e0a36f1d0819e091ff"),
				Version:       new("v1.31.0"),
			},
			imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: new("1.32.x"),
				Repository:    new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           new("v1.32.1@sha256:8c7080fa391ba569c67d4f2470dad4de0f6f1a83aa1a539f7df6fadad1fb6240"),
				Version:       new("v1.32.1"),
			}, imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: new(">= 1.33"),
				Repository:    new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           new("v1.33.0@sha256:4932021e763b81c2679dda43af369619976a1a177726c8a507aa9003e84c18e3"),
				Version:       new("v1.33.0"),
			}))

		By("check gardener images")
		Expect(filterByNamePrefix(imageVector, "gardener-")).To(ContainElements(
			imagevector.ImageSource{
				Name:       "gardener-admission-controller",
				Repository: new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/admission-controller"),
				Tag:        new("v1.128.3@sha256:1b4d5332ebe78b9e4970361230ec043aa967ea70ea6e53b2c3a8538e2e4a496d"),
				Version:    new("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-apiserver",
				Repository: new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/apiserver"),
				Tag:        new("v1.128.3@sha256:d8679b8760f77e540c28d1e32e938b082d3dfdd3b7603666d474726940bb8942"),
				Version:    new("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-controller-manager",
				Repository: new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/controller-manager"),
				Tag:        new("v1.128.3@sha256:f1509f9f7d43902d319a87757612bd369439739fc4381ef77698d3e5447896f7"),
				Version:    new("v1.128.3"),
			}))

		By("check images from a referenced components")
		Expect(filterByNamePrefix(imageVector, "vpn-")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "vpn-client",
				Repository: new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/vpn-client"),
				Tag:        new("0.41.1@sha256:1871708ac9d09183b11d4f6d55548052db89e075faa0eddb1eb8bd5bb8ee956f"),
				Version:    new("0.41.1"),
			},
			imagevector.ImageSource{
				Name:       "vpn-server",
				Repository: new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/vpn-server"),
				Tag:        new("0.41.1@sha256:25b166baba9426d77929dc6cd38ab1e3d6dd846e3e9f847365dd216a1d9dd1ab"),
				Version:    new("0.41.1"),
			}))

		By("check images referenced from Kubernetes component by label `imagevector.gardener.cloud/images`")
		Expect(filterByNamePrefix(imageVector, "hyperkube")).To(ContainElements(
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           new("v1.32.6@sha256:7ea3d70fa985db8b1577cbae3af775af304eec06432518cc4e69d1fb5e48459f"),
				Version:       new("1.32.6"),
				TargetVersion: new("1.32.6"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           new("v1.32.8@sha256:6cd7f4cf3eaabfd34eb0c500b92c1bca0c88e23b851cca6e076f0e2b6f3a18e5"),
				Version:       new("1.32.8"),
				TargetVersion: new("1.32.8"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/hyperkube"),
				Tag:           new("v1.31.10@sha256:54ba6d6336e7ce9585db3ec3bba789bfc0980e1e514b03eec6d69d193bd15c55"),
				Version:       new("1.31.10"),
				TargetVersion: new("1.31.10"),
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
				Repository: new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/ingress-default-backend"),
				Tag:        new("0.24.0@sha256:4c9b83da61f44d255d128c56421450de95fbdd2c9f77d8ff8e9a15b8ca41a924"),
				Version:    new("0.24.0"),
			}))

		By("check resolution if resourceID != component name")
		Expect(filterByNamePrefix(imageVector, "apiserver-proxy-sidecar")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "apiserver-proxy-sidecar",
				Repository: new("registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/apiserver-proxy"),
				Tag:        new("v0.19.0@sha256:18fd91eb57eef02cfb2c0f9943deeefc4e44c4a1577863d808250538af8a6e03"),
				Version:    new("v0.19.0"),
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
				Value:   "{\"helmchartResource\": {\"name\": \"resource-manager\"}, \"imageMapping\": []}",
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
				Value:   gardenletHelmChartImageMapContent,
			},
			components.Resource{
				Name:    "gardenlet",
				Version: "v1.128.3",
				Type:    "ociImage",
				Value:   "registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/gardenlet:v1.128.3@sha256:a5880e6933465e58536fdfb381acee013905ecd6888d94f0d484dff081ab0b11",
				Local:   true,
			},
		))

		By("GLK componentvector")
		componentVersions, err := c.GetGLKComponents(true)
		Expect(err).NotTo(HaveOccurred())
		Expect(componentVersions.Components).To(HaveLen(1))
		Expect(componentVersions.Components[0].Version).To(Equal("v1.128.3"))

		data, err := yaml.Marshal(componentVersions.Components[0].Resources)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal(`admissionController:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/admission-controller:v1.128.3@sha256:1b4d5332ebe78b9e4970361230ec043aa967ea70ea6e53b2c3a8538e2e4a496d
apiserver:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/apiserver:v1.128.3@sha256:d8679b8760f77e540c28d1e32e938b082d3dfdd3b7603666d474726940bb8942
controllerManager:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/controller-manager:v1.128.3@sha256:f1509f9f7d43902d319a87757612bd369439739fc4381ef77698d3e5447896f7
controlplane:
  helmChart:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/charts/gardener/controlplane:v1.128.3@sha256:9916e1dbe8de5bcc354c41adc52b51fe7ef67cc2e1f34d44d1ec51ff06d962e5
  helmchartImagemap: {}
gardenerAdmissionController:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/admission-controller:v1.128.3@sha256:1b4d5332ebe78b9e4970361230ec043aa967ea70ea6e53b2c3a8538e2e4a496d
gardenerApiserver:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/apiserver:v1.128.3@sha256:d8679b8760f77e540c28d1e32e938b082d3dfdd3b7603666d474726940bb8942
gardenerControllerManager:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/controller-manager:v1.128.3@sha256:f1509f9f7d43902d319a87757612bd369439739fc4381ef77698d3e5447896f7
gardenerNodeAgent:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/node-agent:v1.128.3@sha256:3c89691e4b2edc9fc8a824a881b819b4fb1de2af1c980d08262afbbea4896227
gardenerResourceManager:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/resource-manager:v1.128.3@sha256:ce1e87bde456347a364035314092ff699a7522bb3f90c65a2f21a88915ad4e7e
gardenerScheduler:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/scheduler:v1.128.3@sha256:c612f6a97c5c688cd71b4a11bae43a3cbcb4602d0f5fdb9cd8dc60a49224aa71
gardenlet:
  helmChart:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/charts/gardener/gardenlet:v1.128.3@sha256:a5880e6933465e58536fdfb381acee013905ecd6888d94f0d484dff081ab0b11
  helmchartImagemap:
    gardenlet:
      image:
        repository: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/gardenlet
        tag: v1.128.3@sha256:a5880e6933465e58536fdfb381acee013905ecd6888d94f0d484dff081ab0b11
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/gardenlet:v1.128.3@sha256:a5880e6933465e58536fdfb381acee013905ecd6888d94f0d484dff081ab0b11
nodeAgent:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/node-agent:v1.128.3@sha256:3c89691e4b2edc9fc8a824a881b819b4fb1de2af1c980d08262afbbea4896227
operator:
  helmChart:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/charts/gardener/operator:v1.128.3@sha256:1e544a671edc7e5b72c45f1fce3817a67c4eefa3d709855c119708fb04737f5b
  helmchartImagemap: {}
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/operator:v1.128.3@sha256:1e544a671edc7e5b72c45f1fce3817a67c4eefa3d709855c119708fb04737f5b
resourceManager:
  helmChart:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/charts/gardener/resource-manager:v1.128.3@sha256:ce1e87bde456347a364035314092ff699a7522bb3f90c65a2f21a88915ad4e7e
  helmchartImagemap: {}
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/resource-manager:v1.128.3@sha256:ce1e87bde456347a364035314092ff699a7522bb3f90c65a2f21a88915ad4e7e
scheduler:
  ociImage:
    ref: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/scheduler:v1.128.3@sha256:c612f6a97c5c688cd71b4a11bae43a3cbcb4602d0f5fdb9cd8dc60a49224aa71
`))
		Expect(componentVersions.Components[0].ImageVectorOverwrite).To(ContainSubstring(`images:
- name: alertmanager
  repository: registry.example.com/path/to/repo/quay_io/prometheus/alertmanager
  tag: v0.28.1@sha256:ec1de8cc83bac6ec73b7c8bd0530341d70d8407e5749d561681654731881f351
  version: v0.28.1
- name: alpine-conntrack
  repository: registry.example.com/path/to/repo/europe-docker_pkg_dev/gardener-project/releases/gardener/alpine-conntrack
  tag: 3.21.3@sha256:d776104e96516887cd33abb4fc4786fb6c1872cf3e03bd2d53b93c1652b947fa
  version: 3.21.3`))
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
				TargetVersion: new("1.31.x"),
				Repository:    new("europe-docker.pkg.dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           new("v1.31.0"),
				Version:       new("v1.31.0"),
			},
			imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: new("1.32.x"),
				Repository:    new("europe-docker.pkg.dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           new("v1.32.1"),
				Version:       new("v1.32.1"),
			}, imagevector.ImageSource{
				Name:          "cluster-autoscaler",
				TargetVersion: new(">= 1.33"),
				Repository:    new("europe-docker.pkg.dev/gardener-project/releases/gardener/autoscaler/cluster-autoscaler"),
				Tag:           new("v1.33.0"),
				Version:       new("v1.33.0"),
			}))
		Expect(filterByNamePrefix(imageVector, "gardener-")).To(ContainElements(
			imagevector.ImageSource{
				Name:       "gardener-admission-controller",
				Repository: new("europe-docker.pkg.dev/gardener-project/releases/gardener/admission-controller"),
				Tag:        new("v1.128.3"),
				Version:    new("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-apiserver",
				Repository: new("europe-docker.pkg.dev/gardener-project/releases/gardener/apiserver"),
				Tag:        new("v1.128.3"),
				Version:    new("v1.128.3"),
			},
			imagevector.ImageSource{
				Name:       "gardener-controller-manager",
				Repository: new("europe-docker.pkg.dev/gardener-project/releases/gardener/controller-manager"),
				Tag:        new("v1.128.3"),
				Version:    new("v1.128.3"),
			}))

		Expect(filterByNamePrefix(imageVector, "vpn-")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "vpn-client",
				Repository: new("europe-docker.pkg.dev/gardener-project/releases/gardener/vpn-client"),
				Tag:        new("0.41.1"),
				Version:    new("0.41.1"),
			},
			imagevector.ImageSource{
				Name:       "vpn-server",
				Repository: new("europe-docker.pkg.dev/gardener-project/releases/gardener/vpn-server"),
				Tag:        new("0.41.1"),
				Version:    new("0.41.1"),
			}))

		Expect(filterByNamePrefix(imageVector, "hyperkube")).To(ContainElements(
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    new("europe-docker.pkg.dev/gardener-project/releases/hyperkube"),
				Tag:           new("v1.32.6"),
				Version:       new("1.32.6"),
				TargetVersion: new("1.32.6"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    new("europe-docker.pkg.dev/gardener-project/releases/hyperkube"),
				Tag:           new("v1.32.8"),
				Version:       new("1.32.8"),
				TargetVersion: new("1.32.8"),
			},
			imagevector.ImageSource{
				Name:          "hyperkube",
				Repository:    new("europe-docker.pkg.dev/gardener-project/releases/hyperkube"),
				Tag:           new("v1.31.10"),
				Version:       new("1.31.10"),
				TargetVersion: new("1.31.10"),
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
				Repository: new("europe-docker.pkg.dev/gardener-project/releases/gardener/ingress-default-backend"),
				Tag:        new("0.24.0"),
				Version:    new("0.24.0"),
			}))

		// check resolution by resourceID
		Expect(filterByNamePrefix(imageVector, "apiserver-proxy-sidecar")).To(ConsistOf(
			imagevector.ImageSource{
				Name:       "apiserver-proxy-sidecar",
				Repository: new("europe-docker.pkg.dev/gardener-project/releases/gardener/apiserver-proxy"),
				Tag:        new("v0.19.0"),
				Version:    new("v0.19.0"),
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
			json := fmt.Sprintf(`{"helmchartResource": {"name": %q}, "imageMapping": []}`, res.Name)
			if res.Name == "gardenlet" {
				json = gardenletHelmChartImageMapContent
			}
			blobs[ociaccess.ResourceToBlobKey(res)] = []byte(json)
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
