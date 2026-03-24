// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package azure_test

import (
	"os"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	provider_azure "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-azure"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/test"
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
			component := provider_azure.NewComponent()
			Expect(component.GenerateBase(opts)).To(Succeed())

			content, err := fs.ReadFile("/repo/baseDir/components/gardener-extensions/provider-azure/extension.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("apiVersion: operator.gardener.cloud/v1alpha1"))
			Expect(string(content)).To(ContainSubstring("kind: Extension"))

			content, err = fs.ReadFile("/repo/baseDir/components/gardener-extensions/provider-azure/kustomization.yaml")
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
			component := provider_azure.NewComponent()
			landscapeOpts, err := components.NewLandscapeOptions(generateOpts, fs)
			Expect(component.GenerateLandscape(landscapeOpts)).To(Succeed())
			Expect(err).ToNot(HaveOccurred())

			exists, err := fs.DirExists("/repo/baseDir")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())

			content, err := fs.ReadFile("/repo/landscapeDir/components/gardener-extensions/provider-azure/flux-kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("path: landscapeDir/components/gardener-extensions/provider-azure"))

			content, err = fs.ReadFile("/repo/landscapeDir/components/gardener-extensions/provider-azure/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- ../../../../baseDir/components/gardener-extensions/provider-azure"))
		})

		DescribeTable("should generate correct kustomized build output",
			func(build test.BuildComponentVectorFn, expectedFile string) {
				component := provider_azure.NewComponent()
				Expect(test.CreateComponentsVectorFile(fs, build)).To(Succeed())
				result, err := test.KustomizeComponent(fs, component, "components/gardener-extensions/provider-azure")
				Expect(err).ToNot(HaveOccurred())
				expected, err := os.ReadFile(expectedFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(result)).To(Equal(string(expected)))
			},
			Entry("with plain component vector without OCM resources",
				test.NewComponentVectorFactoryBuilder("github.com/gardener/gardener-extension-provider-azure", "v1.2.3").WithDefaultResources().Build(),
				"testdata/expected-kustomize-plain.yaml"),
			Entry("with OCM component vector including helm charts and OCI images",
				test.NewComponentVectorFactoryBuilder("github.com/gardener/gardener-extension-provider-azure", "v1.2.3").
					WithImageVectorOverwrite(componentvector.ImageVectorOverwrite{
						Images: []imagevector.ImageSource{
							{
								Name: "component1",
								Ref:  new("test.repo/path/component1:v1.2.3"),
							},
						},
					}).
					WithResourcesYAML(`
admissionAzureApplication:
  helmChart:
    ref: test-repo/path/charts/gardener/extensions/admission-azure-application:v1.2.3
    imageMap:
      gardenerExtensionAdmissionAzure:
        image:
          repository: test-repo/path/gardener/extensions/admission-azure
          tag: v1.2.3
admissionAzureRuntime:
  helmChart:
    ref: test-repo/path/charts/gardener/extensions/admission-azure-runtime:v1.2.3
    imageMap:
      gardenerExtensionAdmissionAzure:
        image:
          repository: test-repo/path/gardener/extensions/admission-azure
          tag: v1.2.3
gardenerExtensionAdmissionAzure:
  ociImage:
    ref: test-repo/path/gardener/extensions/admission-azure:v1.2.3
gardenerExtensionProviderAzure:
  ociImage:
    ref: test-repo/path/gardener/extensions/provider-azure:v1.2.3
providerAzure:
  helmChart:
    ref: test-repo/path/charts/gardener/extensions/provider-azure:v1.2.3
    imageMap:
      gardenerExtensionProviderAzure:
        image:
          repository: test-repo/path/gardener/extensions/provider-azure
          tag: v1.2.3
`).Build(),
				"testdata/expected-kustomize-ocm.yaml"),
		)
	})
})
