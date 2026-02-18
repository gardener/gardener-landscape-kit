// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package operator_test

import (
	"os"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/spf13/afero"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/components/gardener/operator"
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
		format.CharactersAroundMismatchToInclude = 100
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
			component := operator.NewComponent()
			Expect(component.GenerateBase(opts)).To(Succeed())

			for _, file := range []string{
				"/repo/baseDir/.glk/defaults/components/gardener/operator/oci-repository.yaml",
				"/repo/baseDir/components/gardener/operator/oci-repository.yaml",
			} {
				content, err := fs.ReadFile(file)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("OCIRepository"))
			}

			for _, file := range []string{
				"/repo/baseDir/.glk/defaults/components/gardener/operator/helm-release.yaml",
				"/repo/baseDir/components/gardener/operator/helm-release.yaml",
			} {
				content, err := fs.ReadFile(file)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("HelmRelease"))
			}

			content, err := fs.ReadFile("/repo/baseDir/components/gardener/operator/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(And(
				ContainSubstring("- oci-repository.yaml"),
				ContainSubstring("- helm-release.yaml"),
			))
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
			component := operator.NewComponent()
			landscapeOpts, err := components.NewLandscapeOptions(generateOpts, fs)
			Expect(err).ToNot(HaveOccurred())
			Expect(component.GenerateLandscape(landscapeOpts)).To(Succeed())

			exists, err := fs.DirExists("/repo/baseDir")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())

			content, err := fs.ReadFile("/repo/landscapeDir/components/gardener/operator/flux-kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("path: landscapeDir/components/gardener/operator"))

			content, err = fs.ReadFile("/repo/landscapeDir/components/gardener/operator/kustomization.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(ContainSubstring("- ../../../../baseDir/components/gardener/operator"))
		})
	})

	DescribeTable("Kustomize",
		func(fcv testutils.ComponentVectorFactory, expectedFile string) {
			component := operator.NewComponent()
			componentsVectorFile, err := testutils.CreateComponentsVectorFile(fs, fcv)
			Expect(err).ToNot(HaveOccurred())
			result, err := testutils.KustomizeComponent(fs, component, "components/gardener/operator", componentsVectorFile)
			Expect(err).ToNot(HaveOccurred())
			expected, err := os.ReadFile(expectedFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal(string(expected)))
		},
		Entry("plain",
			testutils.ComponentVector("github.com/gardener/gardener", "v1.2.3").Build(),
			"testdata/expected-kustomize-plain.yaml"),
		Entry("ocm",
			testutils.ComponentVector("github.com/gardener/gardener", "v1.2.3").
				WithImageVectorOverwrite(componentvector.ImageVectorOverwrite{
					Images: []imagevector.ImageSource{
						{
							Name: "component1",
							Ref:  ptr.To("test.repo/path/component1:v1.2.3"),
						},
					},
				}).
				WithComponentImageVectorOverwrites(componentvector.ComponentImageVectorOverwrites{
					Components: []componentvector.ComponentImageVectorOverwrite{
						{
							Name: "etcd-druid",
							ImageVectorOverwrite: componentvector.ImageVectorOverwrite{
								Images: []imagevector.ImageSource{
									{
										Name: "component2",
										Ref:  ptr.To("test.repo/path/component2:v1.2.3"),
									},
								},
							},
						},
					},
				}).
				WithResourcesYAML(`
operator:
  helmChart:
    ref: test.repo/path/charts/gardener/operator:v1.2.3
  helmchartImagemap:
    operator:
      image:
        repository: test.repo/path/gardener/operator
        tag: v1.2.3
  ociImage:
     ref: test.repo/path/gardener/operator:v1.2.3
`).Build(),
			"testdata/expected-kustomize-ocm.yaml"),
	)
})
