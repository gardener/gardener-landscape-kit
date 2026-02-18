// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package calico_test

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
	networking_calico "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/networking-calico"
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
			component := networking_calico.NewComponent()
			Expect(component.GenerateBase(opts)).To(Succeed())

			content, err := fs.ReadFile("/repo/baseDir/components/gardener-extensions/networking-calico/extension.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("apiVersion: operator.gardener.cloud/v1alpha1"))
			Expect(string(content)).To(ContainSubstring("kind: Extension"))

			content, err = fs.ReadFile("/repo/baseDir/components/gardener-extensions/networking-calico/kustomization.yaml")
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
			component := networking_calico.NewComponent()
			landscapeOpts, err := components.NewLandscapeOptions(generateOpts, fs)
			Expect(component.GenerateLandscape(landscapeOpts)).To(Succeed())
			Expect(err).ToNot(HaveOccurred())

			exists, err := fs.DirExists("/repo/baseDir")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())

			content, err := fs.ReadFile("/repo/landscapeDir/components/gardener-extensions/networking-calico/flux-kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("path: landscapeDir/components/gardener-extensions/networking-calico"))

			content, err = fs.ReadFile("/repo/landscapeDir/components/gardener-extensions/networking-calico/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- ../../../../baseDir/components/gardener-extensions/networking-calico"))
		})
	})

	DescribeTable("Kustomize",
		func(fcv testutils.ComponentVectorFactory, expectedFile string) {
			component := networking_calico.NewComponent()
			componentsVectorFile, err := testutils.CreateComponentsVectorFile(fs, fcv)
			Expect(err).ToNot(HaveOccurred())
			result, err := testutils.KustomizeComponent(fs, component, "components/gardener-extensions/networking-calico", componentsVectorFile)
			Expect(err).ToNot(HaveOccurred())
			expected, err := os.ReadFile(expectedFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal(string(expected)))
		},
		Entry("plain",
			testutils.ComponentVector("github.com/gardener/gardener-extension-networking-calico", "v1.2.3").Build(),
			"testdata/expected-kustomize-plain.yaml"),
		Entry("ocm",
			testutils.ComponentVector("github.com/gardener/gardener-extension-networking-calico", "v1.2.3").
				WithImageVectorOverwrite(componentvector.ImageVectorOverwrite{
					Images: []imagevector.ImageSource{
						{
							Name: "component1",
							Ref:  ptr.To("test.repo/path/component1:v1.2.3"),
						},
					},
				}).
				WithResourcesYAML(`
admissionCalicoApplication:
  helmChart:
    ref: test-repo/path/charts/gardener/extensions/admission-calico-application:v1.2.3
  helmchartImagemap:
    gardenerExtensionAdmissionCalico:
      image:
        repository: test-repo/path/gardener/extensions/admission-calico
        tag: v1.2.3
admissionCalicoRuntime:
  helmChart:
    ref: test-repo/path/charts/gardener/extensions/admission-calico-runtime:v1.2.3
  helmchartImagemap:
    gardenerExtensionAdmissionCalico:
      image:
        repository: test-repo/path/gardener/extensions/admission-calico
        tag: v1.2.3
cniPlugins:
  ociImage:
    ref: test-repo/path/gardener/extensions/cni-plugins:v1.2.3
gardenerExtensionAdmissionCalico:
  ociImage:
    ref: test-repo/path/gardener/extensions/admission-calico:v1.2.3
gardenerExtensionNetworkingCalico:
  ociImage:
    ref: test-repo/path/gardener/extensions/networking-calico:v1.2.3
networkingCalico:
  helmChart:
    ref: test-repo/path/charts/gardener/extensions/networking-calico:v1.2.3
  helmchartImagemap:
    gardenerExtensionNetworkingCalico:
      image:
        repository: test-repo/path/gardener/extensions/networking-calico
        tag: v1.2.3
`).Build(),
			"testdata/expected-kustomize-ocm.yaml"),
	)
})
