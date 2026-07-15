// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux_test

import (
	"strings"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	. "github.com/gardener/gardener-landscape-kit/pkg/components/flux"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/test"
)

var (
	scheme    = runtime.NewScheme()
	fluxCodec = serializer.NewCodecFactory(scheme)
)

func init() {
	utilruntime.Must(kustomizev1.AddToScheme(scheme))
	utilruntime.Must(sourcev1.AddToScheme(scheme))
}

var _ = Describe("Flux Component Generation", func() {
	var (
		targetPath            string
		relativeLandscapePath string
		repoURL               string

		fs     afero.Afero
		config *v1alpha1.LandscapeKitConfiguration
		opts   components.LandscapeOptions
	)

	buildOpts := func() components.LandscapeOptions {
		v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(config)
		o, err := components.NewLandscapeOptions(
			&generateoptions.Options{
				Options: &cmd.Options{
					Log: logr.Discard(),
				},
				TargetDirPath: targetPath,
				Config:        config,
			},
			fs,
		)
		Expect(err).NotTo(HaveOccurred())
		return o
	}

	BeforeEach(func() {
		targetPath = "/"
		relativeLandscapePath = "./landscapeDir"
		repoURL = "https://github.com/gardener/gardener-ref-landscape"

		fs = afero.Afero{Fs: afero.NewMemMapFs()}

		config = &v1alpha1.LandscapeKitConfiguration{
			Repositories: &v1alpha1.RepositoriesConfig{
				Landscape: &v1alpha1.LandscapeRepositoryConfig{
					URL:    repoURL,
					Target: relativeLandscapePath,
				},
			},
		}

		opts = buildOpts()
	})

	Describe("#GenerateLandscape", func() {
		It("should correctly generate the flux landscape directory", func() {
			component := NewComponent()
			Expect(component.GenerateLandscape(opts)).To(Succeed())
		})

		It("should not recreate a deleted gitignore file", func() {
			component := NewComponent()
			Expect(component.GenerateLandscape(opts)).To(Succeed())
			Expect(fs.Exists("/landscapeDir/flux/flux-system/.gitignore")).To(BeTrue())

			Expect(fs.Remove("/landscapeDir/flux/flux-system/.gitignore")).To(Succeed())

			Expect(component.GenerateLandscape(opts)).To(Succeed())

			Expect(fs.Exists("/landscapeDir/flux/flux-system/.gitignore")).To(BeFalse())
		})

		It("should not reformat previously generated manifests (idempotency)", func() {
			component := NewComponent()
			Expect(component.GenerateLandscape(opts)).To(Succeed())

			initialContents, err := fs.ReadFile("/landscapeDir/flux/flux-system/gotk-sync.yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(component.GenerateLandscape(opts)).To(Succeed())

			newContents, err := fs.ReadFile("/landscapeDir/flux/flux-system/gotk-sync.yaml")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(initialContents)).To(Equal(string(newContents)))
		})

		Context("GOTK Sync Manifest", func() {
			test := func(opts components.LandscapeOptions, refMatcher types.GomegaMatcher) {
				component := NewComponent()
				Expect(component.GenerateLandscape(opts)).To(Succeed())

				gotkData, err := fs.ReadFile("/landscapeDir/flux/flux-system/gotk-sync.yaml")
				Expect(err).NotTo(HaveOccurred())

				objects := make([]client.Object, 0, 2)
				for objRaw := range strings.SplitSeq(string(gotkData), "---\n") {
					if objRaw == "" {
						continue
					}

					obj, _, err := fluxCodec.UniversalDeserializer().Decode([]byte(objRaw), nil, nil)
					Expect(err).NotTo(HaveOccurred())

					objects = append(objects, obj.(client.Object))
				}

				Expect(objects).To(ConsistOf(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"TypeMeta": MatchFields(IgnoreExtras, Fields{
							"Kind": Equal("GitRepository"),
						}),
						"Spec": MatchFields(IgnoreExtras, Fields{
							"Reference": PointTo(refMatcher),
							"URL":       Equal(repoURL),
						}),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"TypeMeta": MatchFields(IgnoreExtras, Fields{
							"Kind": Equal("Kustomization"),
						}),
						"Spec": MatchFields(IgnoreExtras, Fields{
							"Path": Equal(relativeLandscapePath + "/flux"),
						}),
					})),
				))
			}

			It("should contain the correct repository URL, path and default branch", func() {
				test(opts, MatchFields(IgnoreExtras, Fields{
					"Branch": Equal("main"),
				}))
			})

			It("should contain the correct repository URL, path and branch", func() {
				config.Repositories.Landscape.Ref = v1alpha1.GitRepositoryRef{
					Branch: new("develop"),
				}
				opts = buildOpts()

				test(opts, MatchFields(IgnoreExtras, Fields{
					"Branch": Equal("develop"),
				}))
			})

			It("should contain the correct repository URL, path and tag", func() {
				config.Repositories.Landscape.Ref = v1alpha1.GitRepositoryRef{
					Tag: new("v1.0.0"),
				}
				opts = buildOpts()

				test(opts, MatchFields(IgnoreExtras, Fields{
					"Tag": Equal("v1.0.0"),
				}))
			})

			It("should contain the correct repository URL, path and commit", func() {
				config.Repositories.Landscape.Ref = v1alpha1.GitRepositoryRef{
					Commit: new("a1b2c3d4"),
				}
				opts = buildOpts()

				test(opts, MatchFields(IgnoreExtras, Fields{
					"Commit": Equal("a1b2c3d4"),
				}))
			})

			It("should contain prefer the branch configuration", func() {
				config.Repositories.Landscape.Ref = v1alpha1.GitRepositoryRef{
					Branch: new("develop"),
					Tag:    new("v1.0.0"),
				}
				opts = buildOpts()

				test(opts, MatchFields(IgnoreExtras, Fields{
					"Tag": Equal("v1.0.0"),
				}))
			})

			It("should contain prefer the commit configuration", func() {
				config.Repositories.Landscape.Ref = v1alpha1.GitRepositoryRef{
					Branch: new("develop"),
					Tag:    new("v1.0.0"),
					Commit: new("a1b2c3d4"),
				}
				opts = buildOpts()

				test(opts, MatchFields(IgnoreExtras, Fields{
					"Commit": Equal("a1b2c3d4"),
				}))
			})
		})

		Context("GOTK Components Manifest", func() {
			writeComponentsVectorFile := func(resourcesYAML string) {
				resources, err := test.UnmarshalToResources(resourcesYAML)
				Expect(err).NotTo(HaveOccurred())
				cv := &utilscomponentvector.ComponentVector{
					Name:      componentvector.NameGardenerGardenerLandscapeKit,
					Version:   "v0.3.0-dev",
					Resources: resources,
				}
				cvs := &utilscomponentvector.Components{Components: []*utilscomponentvector.ComponentVector{cv}}
				content, err := yaml.Marshal(cvs)
				Expect(err).NotTo(HaveOccurred())
				// The flux test uses targetPath="/" → repoRoot="/", baseLink="", base.Target="./"
				// so the component vector override file is read from "/components.yaml".
				Expect(fs.WriteFile("/components.yaml", content, 0o644)).To(Succeed())
			}

			It("should use default images when no component vector resources are set", func() {
				component := NewComponent()
				Expect(component.GenerateLandscape(opts)).To(Succeed())

				data, err := fs.ReadFile("/landscapeDir/flux/flux-system/gotk-components.yaml")
				Expect(err).NotTo(HaveOccurred())
				content := string(data)

				Expect(content).To(ContainSubstring("image: ghcr.io/fluxcd/source-controller:v"))
				Expect(content).To(ContainSubstring("image: ghcr.io/fluxcd/kustomize-controller:v"))
				Expect(content).To(ContainSubstring("image: ghcr.io/fluxcd/helm-controller:v"))
				Expect(content).To(ContainSubstring("image: ghcr.io/fluxcd/notification-controller:v"))
			})

			It("should use overridden images from the component vector resources", func() {
				writeComponentsVectorFile(`
sourceController:
  ociImage:
    ref: my.registry.io/fluxcd/source-controller:v9.9.9
kustomizeController:
  ociImage:
    ref: my.registry.io/fluxcd/kustomize-controller:v9.9.9
helmController:
  ociImage:
    ref: my.registry.io/fluxcd/helm-controller:v9.9.9
notificationController:
  ociImage:
    ref: my.registry.io/fluxcd/notification-controller:v9.9.9
`)
				opts = buildOpts()

				component := NewComponent()
				Expect(component.GenerateLandscape(opts)).To(Succeed())

				data, err := fs.ReadFile("/landscapeDir/flux/flux-system/gotk-components.yaml")
				Expect(err).NotTo(HaveOccurred())
				content := string(data)

				Expect(content).To(ContainSubstring("image: my.registry.io/fluxcd/source-controller:v9.9.9"))
				Expect(content).To(ContainSubstring("image: my.registry.io/fluxcd/kustomize-controller:v9.9.9"))
				Expect(content).To(ContainSubstring("image: my.registry.io/fluxcd/helm-controller:v9.9.9"))
				Expect(content).To(ContainSubstring("image: my.registry.io/fluxcd/notification-controller:v9.9.9"))

				Expect(content).NotTo(ContainSubstring("ghcr.io/fluxcd"))
			})
		})
	})
})
