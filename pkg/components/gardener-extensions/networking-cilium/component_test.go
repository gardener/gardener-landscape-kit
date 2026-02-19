// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cilium_test

import (
	"os"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	networking_cilium "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/networking-cilium"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
	testutils "github.com/gardener/gardener-landscape-kit/test/utils"
)

var _ = Describe("Component Generation", func() {
	var (
		fs           afero.Afero
		cmdOpts      *cmd.Options
		generateOpts *generateoptions.Options
	)

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
		cmdOpts = &cmd.Options{Log: logr.Discard()}
		generateOpts = &generateoptions.Options{
			TargetDirPath: "/repo/baseDir",
			Options:       cmdOpts,
		}
	})

	Describe("#GenerateBase", func() {
		var opts components.Options

		BeforeEach(func() {
			var err error
			opts, err = components.NewOptions(generateOpts, fs)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should generate the component base", func() {
			component := networking_cilium.NewComponent()
			Expect(component.GenerateBase(opts)).To(Succeed())

			content, err := fs.ReadFile("/repo/baseDir/components/gardener-extensions/networking-cilium/extension.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("apiVersion: operator.gardener.cloud/v1alpha1"))
			Expect(string(content)).To(ContainSubstring("kind: Extension"))

			content, err = fs.ReadFile("/repo/baseDir/components/gardener-extensions/networking-cilium/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- extension.yaml"))
		})
	})

	Describe("#GenerateLandscape", func() {
		BeforeEach(func() {
			generateOpts.TargetDirPath = "/repo/landscapeDir"
			generateOpts.Config = &v1alpha1.LandscapeKitConfiguration{
				Git: &v1alpha1.GitRepository{Paths: v1alpha1.PathConfiguration{Landscape: "./landscapeDir", Base: "./baseDir"}},
			}
		})

		It("should generate only the flux kustomization into the landscape dir", func() {
			component := networking_cilium.NewComponent()
			landscapeOpts, err := components.NewLandscapeOptions(generateOpts, fs)
			Expect(component.GenerateLandscape(landscapeOpts)).To(Succeed())
			Expect(err).ToNot(HaveOccurred())

			exists, err := fs.DirExists("/repo/baseDir")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())

			content, err := fs.ReadFile("/repo/landscapeDir/components/gardener-extensions/networking-cilium/flux-kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("path: landscapeDir/components/gardener-extensions/networking-cilium"))

			content, err = fs.ReadFile("/repo/landscapeDir/components/gardener-extensions/networking-cilium/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- ../../../../baseDir/components/gardener-extensions/networking-cilium"))
		})
	})

	DescribeTable("Kustomize",
		func(fcv testutils.ComponentVectorFactory, expectedFile string) {
			component := networking_cilium.NewComponent()
			componentsVectorFile, err := testutils.CreateComponentsVectorFile(fs, fcv)
			Expect(err).ToNot(HaveOccurred())
			result, err := testutils.KustomizeComponent(fs, component, "components/gardener-extensions/networking-cilium", componentsVectorFile)
			Expect(err).ToNot(HaveOccurred())
			expected, err := os.ReadFile(expectedFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal(string(expected)))
		},
		Entry("plain",
			testutils.ComponentVector("github.com/gardener/gardener-extension-networking-cilium", "v1.2.3").Build(),
			"testdata/expected-kustomize-plain.yaml"),
		Entry("ocm",
			testutils.ComponentVector("github.com/gardener/gardener-extension-networking-cilium", "v1.2.3").
				WithImageVectorOverwrite(componentvector.ImageVectorOverwrite{
					Images: []imagevector.ImageSource{
						{
							Name: "component1",
							Ref:  ptr.To("test.repo/path/component1:v1.2.3"),
						},
					},
				}).
				WithResourcesYAML(`
admissionCiliumApplication:
  helmChartRef: test-repo/path/charts/gardener/extensions/admission-cilium-application:v1.2.3
  helmChartImageMap:
    gardenerExtensionAdmissionCilium:
      image:
        repository: test-repo/path/gardener/extensions/admission-cilium
        tag: v1.2.3
admissionCiliumRuntime:
  helmChartRef: test-repo/path/charts/gardener/extensions/admission-cilium-runtime:v1.2.3
  helmChartImageMap:
    gardenerExtensionAdmissionCilium:
      image:
        repository: test-repo/path/gardener/extensions/admission-cilium
        tag: v1.2.3
gardenerExtensionAdmissionCilium:
  ociImageRef: test-repo/path/gardener/extensions/admission-cilium:v1.2.3
gardenerExtensionNetworkingCilium:
  ociImageRef: test-repo/path/gardener/extensions/networking-cilium:v1.2.3
networkingCilium:
  helmChartRef: test-repo/path/charts/gardener/extensions/networking-cilium:v1.2.3
  helmChartImageMap:
    gardenerExtensionNetworkingCilium:
      image:
        repository: test-repo/path/gardener/extensions/networking-cilium
        tag: v1.2.3
`).Build(),
			"testdata/expected-kustomize-ocm.yaml"),
	)
})
