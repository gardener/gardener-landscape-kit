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

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	. "github.com/gardener/gardener-landscape-kit/pkg/components/gardener-extensions/networking-calico"
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
			Config:        &v1alpha1.LandscapeKitConfiguration{},
		}
		v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(generateOpts.Config)
	})

	Describe("#GenerateBase", func() {
		var opts components.Options

		BeforeEach(func() {
			var err error
			opts, err = components.NewOptions(generateOpts, fs)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should generate the component base", func() {
			component := NewComponent()
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
			generateOpts.TargetDirPath = "/repo"
			generateOpts.Config = &v1alpha1.LandscapeKitConfiguration{
				Repositories: &v1alpha1.RepositoriesConfig{
					Base: &v1alpha1.BaseRepositoryConfig{Target: "."},
					Landscape: &v1alpha1.LandscapeRepositoryConfig{
						BaseLink: "./baseDir",
						Target:   "./landscapeDir",
					},
				},
			}
			v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(generateOpts.Config)
		})

		It("should generate only the flux kustomization into the landscape dir", func() {
			component := NewComponent()
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

		DescribeTable("should generate correct kustomized build output",
			func(build test.BuildComponentVectorFn, expectedFile string) {
				component := NewComponent()
				optsFn, err := test.CreateComponentsVectorFile(fs, build)
				Expect(err).ToNot(HaveOccurred())
				result, err := test.KustomizeComponent(fs, component, "components/gardener-extensions/networking-calico", optsFn)
				Expect(err).ToNot(HaveOccurred())
				expected, err := os.ReadFile(expectedFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(result)).To(Equal(string(expected)))
			},
			Entry("with plain component vector without OCM resources",
				test.NewComponentVectorFactoryBuilder("github.com/gardener/gardener-extension-networking-calico", "v1.2.3").WithDefaultResources().Build(),
				"testdata/expected-kustomize-plain.yaml"),
			Entry("with OCM component vector including helm charts and OCI images",
				test.NewComponentVectorFactoryBuilder("github.com/gardener/gardener-extension-networking-calico", "v1.2.3").
					WithImageVectorOverwrite(componentvector.ImageVectorOverwrite{
						Images: []imagevector.ImageSource{
							{
								Name: "component1",
								Ref:  new("test.repo/path/component1:v1.2.3"),
							},
						},
					}).
					WithResourcesYAML(`
admissionCalicoApplication:
  helmChart:
    ref: test-repo/path/charts/gardener/extensions/admission-calico-application:v1.2.3
    imageMap:
      gardenerExtensionAdmissionCalico:
        image:
          repository: test-repo/path/gardener/extensions/admission-calico
          tag: v1.2.3
admissionCalicoRuntime:
  helmChart:
    ref: test-repo/path/charts/gardener/extensions/admission-calico-runtime:v1.2.3
    imageMap:
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
    imageMap:
      gardenerExtensionNetworkingCalico:
        image:
          repository: test-repo/path/gardener/extensions/networking-calico
          tag: v1.2.3
`).Build(),
				"testdata/expected-kustomize-ocm.yaml"),
		)
	})
})
