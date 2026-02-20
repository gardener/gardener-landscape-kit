// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package aws_test

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
	provider_aws "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/provider-aws"
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
			component := provider_aws.NewComponent()
			Expect(component.GenerateBase(opts)).To(Succeed())

			content, err := fs.ReadFile("/repo/baseDir/components/gardener-extensions/provider-aws/extension.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("apiVersion: operator.gardener.cloud/v1alpha1"))
			Expect(string(content)).To(ContainSubstring("kind: Extension"))

			content, err = fs.ReadFile("/repo/baseDir/components/gardener-extensions/provider-aws/kustomization.yaml")
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
			component := provider_aws.NewComponent()
			landscapeOpts, err := components.NewLandscapeOptions(generateOpts, fs)
			Expect(component.GenerateLandscape(landscapeOpts)).To(Succeed())
			Expect(err).ToNot(HaveOccurred())

			exists, err := fs.DirExists("/repo/baseDir")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())

			content, err := fs.ReadFile("/repo/landscapeDir/components/gardener-extensions/provider-aws/flux-kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("path: landscapeDir/components/gardener-extensions/provider-aws"))

			content, err = fs.ReadFile("/repo/landscapeDir/components/gardener-extensions/provider-aws/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- ../../../../baseDir/components/gardener-extensions/provider-aws"))
		})

		DescribeTable("should generate correct kustomize build output",
			func(build test.BuildComponentVectorFn, expectedFile string) {
				component := provider_aws.NewComponent()
				componentsVectorFile, err := test.CreateComponentsVectorFile(fs, build)
				Expect(err).ToNot(HaveOccurred())
				result, err := test.KustomizeComponent(fs, component, "components/gardener-extensions/provider-aws", componentsVectorFile)
				Expect(err).ToNot(HaveOccurred())
				expected, err := os.ReadFile(expectedFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(result)).To(Equal(string(expected)))
			},
			Entry("with plain component vector without OCM resources",
				test.NewComponentVectorFactoryBuilder("github.com/gardener/gardener-extension-provider-aws", "v1.2.3").Build(),
				"testdata/expected-kustomize-plain.yaml"),
			Entry("with OCM component vector including helm charts and OCI images",
				test.NewComponentVectorFactoryBuilder("github.com/gardener/gardener-extension-provider-aws", "v1.2.3").
					WithImageVectorOverwrite(componentvector.ImageVectorOverwrite{
						Images: []imagevector.ImageSource{
							{
								Name: "component1",
								Ref:  ptr.To("test.repo/path/component1:v1.2.3"),
							},
						},
					}).
					WithResourcesYAML(`
admissionAwsApplication:
  helmChartRef: test-repo/path/charts/gardener/extensions/admission-aws-application:v1.2.3
  helmChartImageMap:
    gardenerExtensionAdmissionAws:
      image:
        repository: test-repo/path/gardener/extensions/admission-aws
        tag: v1.2.3
admissionAwsRuntime:
  helmChartRef: test-repo/path/charts/gardener/extensions/admission-aws-runtime:v1.2.3
  helmChartImageMap:
    gardenerExtensionAdmissionAws:
      image:
        repository: test-repo/path/gardener/extensions/admission-aws
        tag: v1.2.3
gardenerExtensionAdmissionAws:
  ociImageRef: test-repo/path/gardener/extensions/admission-aws:v1.2.3
gardenerExtensionProviderAws:
  ociImageRef: test-repo/path/gardener/extensions/provider-aws:v1.2.3
providerAws:
  helmChartRef: test-repo/path/charts/gardener/extensions/provider-aws:v1.2.3
  helmChartImageMap:
    gardenerExtensionProviderAws:
      image:
        repository: test-repo/path/gardener/extensions/provider-aws
        tag: v1.2.3
`).Build(),
				"testdata/expected-kustomize-ocm.yaml"),
		)
	})
})
