// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	"errors"

	"github.com/gardener/gardener/pkg/utils/test"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	. "github.com/gardener/gardener-landscape-kit/pkg/registry"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

var _ = Describe("Registry", func() {
	var (
		reg Interface
		log logr.Logger

		config           *v1alpha1.LandscapeKitConfiguration
		options          components.Options
		landscapeOptions components.LandscapeOptions
	)

	BeforeEach(func() {
		reg = New(nil, nil)
		log = logr.Discard()

		config = &v1alpha1.LandscapeKitConfiguration{
			Repositories: &v1alpha1.RepositoriesConfig{
				Landscape: &v1alpha1.LandscapeRepositoryConfig{},
			},
		}
		v1alpha1.SetObjectDefaults_LandscapeKitConfiguration(config)

		var err error
		options, err = components.NewOptions(
			&generateoptions.Options{
				Options: &cmd.Options{Log: log},
				Config:  config,
			},
			afero.Afero{Fs: afero.NewMemMapFs()},
		)
		Expect(err).NotTo(HaveOccurred())

		landscapeOptions, err = components.NewLandscapeOptions(
			&generateoptions.Options{
				Options: &cmd.Options{Log: log},
				Config:  config,
			},
			afero.Afero{Fs: afero.NewMemMapFs()},
		)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("#RegisterComponent", func() {
		It("should register components", func() {
			mockComp1 := &mockComponent{
				name: "mockComp1",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}
			mockComp2 := &mockComponent{
				name: "mockComp2",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			Expect(reg.RegisterComponent(log, mockComp1)).To(Succeed())
			Expect(reg.RegisterComponent(log, mockComp2)).To(Succeed())

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockComp1.generateBaseCalled).To(BeTrue())
			Expect(mockComp2.generateBaseCalled).To(BeTrue())
		})

		It("should create the correct component context from the provided vectors", func() {
			const testComponentRef = "github.com/gardener/test-extension"

			currentYAML := []byte(`components:
- name: github.com/gardener/test-extension
  sourceRepository: https://github.com/gardener/test-extension
  version: v1.0.0
`)
			nextYAML := []byte(`components:
- name: github.com/gardener/test-extension
  sourceRepository: https://github.com/gardener/test-extension
  version: v2.0.0
`)
			currentCV, err := componentvector.NewWithOverride(currentYAML)
			Expect(err).NotTo(HaveOccurred())
			nextCV, err := componentvector.NewWithOverride(nextYAML)
			Expect(err).NotTo(HaveOccurred())

			regWithVectors := New(currentCV, nextCV)

			var receivedCtx components.Context
			comp := &mockComponent{
				name:         "test-extension",
				componentRef: testComponentRef,
				captureCtx:   func(ctx components.Context) { receivedCtx = ctx },
			}

			Expect(regWithVectors.RegisterComponent(log, comp)).To(Succeed())
			Expect(regWithVectors.GenerateBase(options)).To(Succeed())

			Expect(receivedCtx).NotTo(BeNil())
			compCtx, err := receivedCtx.Own(comp)
			Expect(err).NotTo(HaveOccurred())
			Expect(compCtx.GetUpgradePath().CurrentVersion).To(Equal("v1.0.0"))
			Expect(compCtx.GetUpgradePath().NextVersion).To(Equal("v2.0.0"))
		})
	})

	Describe("#GenerateBase", func() {
		It("should successfully generate with no components", func() {
			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should call GenerateBase on all registered components", func() {
			mockComp1 := &mockComponent{
				name: "mockComp1",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}
			mockComp2 := &mockComponent{
				name: "mockComp2",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			Expect(reg.RegisterComponent(log, mockComp1)).To(Succeed())
			Expect(reg.RegisterComponent(log, mockComp2)).To(Succeed())

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockComp1.generateBaseCalled).To(BeTrue())
			Expect(mockComp2.generateBaseCalled).To(BeTrue())
		})

		It("should pass options to components", func() {
			var receivedOpts components.Options
			mockComp := &mockComponent{
				name: "mockComp",
				generateBaseFunc: func(opts components.Options) error {
					receivedOpts = opts
					return nil
				},
			}

			Expect(reg.RegisterComponent(log, mockComp)).To(Succeed())

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(receivedOpts).To(Equal(options))
		})

		It("should return error if a component fails", func() {
			expectedErr := errors.New("component error")
			mockComp := &mockComponent{
				name: "mockComp",
				generateBaseFunc: func(_ components.Options) error {
					return expectedErr
				},
			}

			Expect(reg.RegisterComponent(log, mockComp)).To(Succeed())

			err := reg.GenerateBase(options)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		})

		It("should stop at first error and not call subsequent components", func() {
			expectedErr := errors.New("first component error")
			mockComp1 := &mockComponent{
				name: "mockComp1",
				generateBaseFunc: func(_ components.Options) error {
					return expectedErr
				},
			}
			mockComp2 := &mockComponent{
				name: "mockComp2",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			Expect(reg.RegisterComponent(log, mockComp1)).To(Succeed())
			Expect(reg.RegisterComponent(log, mockComp2)).To(Succeed())

			err := reg.GenerateBase(options)
			Expect(err).To(Equal(expectedErr))
			Expect(mockComp1.generateBaseCalled).To(BeTrue())
			Expect(mockComp2.generateBaseCalled).To(BeFalse())
		})
	})

	Describe("GenerateLandscape", func() {
		It("should successfully generate with no components", func() {
			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should call GenerateLandscape on all registered components", func() {
			mockComp1 := &mockComponent{
				name: "mockComp1",
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return nil
				},
			}
			mockComp2 := &mockComponent{
				name: "mockComp2",
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return nil
				},
			}

			Expect(reg.RegisterComponent(log, mockComp1)).To(Succeed())
			Expect(reg.RegisterComponent(log, mockComp2)).To(Succeed())

			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockComp1.generateLandscapeCalled).To(BeTrue())
			Expect(mockComp2.generateLandscapeCalled).To(BeTrue())
		})

		It("should pass options to components", func() {
			var receivedOpts components.LandscapeOptions
			mockComp := &mockComponent{
				name: "mockComp",
				generateLandscapeFunc: func(opts components.LandscapeOptions) error {
					receivedOpts = opts
					return nil
				},
			}

			Expect(reg.RegisterComponent(log, mockComp)).To(Succeed())

			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(receivedOpts).To(Equal(landscapeOptions))
		})

		It("should return error if a component fails", func() {
			expectedErr := errors.New("landscape component error")
			mockComp := &mockComponent{
				name: "mockComp",
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return expectedErr
				},
			}

			Expect(reg.RegisterComponent(log, mockComp)).To(Succeed())

			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		})

		It("should stop at first error and not call subsequent components", func() {
			expectedErr := errors.New("first landscape component error")
			mockComp1 := &mockComponent{
				name: "mockComp1",
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return expectedErr
				},
			}
			mockComp2 := &mockComponent{
				name: "mockComp2",
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return nil
				},
			}

			Expect(reg.RegisterComponent(logr.Discard(), mockComp1)).To(Succeed())
			Expect(reg.RegisterComponent(logr.Discard(), mockComp2)).To(Succeed())

			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).To(Equal(expectedErr))
			Expect(mockComp1.generateLandscapeCalled).To(BeTrue())
			Expect(mockComp2.generateLandscapeCalled).To(BeFalse())
		})
	})

	Describe("Integration", func() {
		It("should work with components that implement both GenerateBase and GenerateLandscape", func() {
			mockComp := &mockComponent{
				name: "mockComp",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return nil
				},
			}

			Expect(reg.RegisterComponent(logr.Discard(), mockComp)).To(Succeed())

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockComp.generateBaseCalled).To(BeTrue())

			err = reg.GenerateLandscape(landscapeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockComp.generateLandscapeCalled).To(BeTrue())
		})

		It("should maintain component order during generation", func() {
			callOrder := []string{}

			mockComp1 := &mockComponent{
				name: "mockComp1",
				generateBaseFunc: func(_ components.Options) error {
					callOrder = append(callOrder, "comp1-base")
					return nil
				},
			}
			mockComp2 := &mockComponent{
				name: "mockComp2",
				generateBaseFunc: func(_ components.Options) error {
					callOrder = append(callOrder, "comp2-base")
					return nil
				},
			}
			mockComp3 := &mockComponent{
				name: "mockComp3",
				generateBaseFunc: func(_ components.Options) error {
					callOrder = append(callOrder, "comp3-base")
					return nil
				},
			}

			Expect(reg.RegisterComponent(logr.Discard(), mockComp1)).To(Succeed())
			Expect(reg.RegisterComponent(logr.Discard(), mockComp2)).To(Succeed())
			Expect(reg.RegisterComponent(logr.Discard(), mockComp3)).To(Succeed())

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(callOrder).To(Equal([]string{"comp1-base", "comp2-base", "comp3-base"}))
		})
	})

	Describe("#RegisterAllComponents", func() {
		var (
			mockComp1, mockComp2, mockComp3 *mockComponent

			mockComponents []func() (components.Interface, error)
		)

		BeforeEach(func() {
			mockComp1 = &mockComponent{
				name: "mockComp1",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			mockComp2 = &mockComponent{
				name: "mockComp2",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			mockComp3 = &mockComponent{
				name: "mockComp3",
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			mockComponents = []func() (components.Interface, error){
				func() (components.Interface, error) {
					return mockComp1, nil
				},
				func() (components.Interface, error) {
					return mockComp2, nil
				},
				func() (components.Interface, error) {
					return mockComp3, nil
				},
			}

			DeferCleanup(test.WithVars(&ComponentList, mockComponents))
		})

		It("should register all components except excluded ones", func() {
			config.Components = &v1alpha1.ComponentsConfiguration{
				Exclude: []string{"mockComp2"},
			}

			Expect(RegisterAllComponents(logr.Discard(), reg, config)).To(Succeed())
			Expect(reg.GenerateBase(options)).To(Succeed())

			Expect(mockComp1.generateBaseCalled).To(BeTrue())
			Expect(mockComp2.generateBaseCalled).To(BeFalse())
			Expect(mockComp3.generateBaseCalled).To(BeTrue())
		})

		It("should return an error if an unknown component is excluded", func() {
			config.Components = &v1alpha1.ComponentsConfiguration{
				Exclude: []string{"unknown", "mockComp2", "unknown2"},
			}

			Expect(RegisterAllComponents(logr.Discard(), reg, config)).To(MatchError(And(
				ContainSubstring(`configuration contains invalid component excludes`),
				ContainSubstring(`unknown`),
				ContainSubstring(`unknown2`),
				ContainSubstring(`available component names are: mockComp1, mockComp2, mockComp3`),
			)))
		})

		It("should register only included components", func() {
			config.Components = &v1alpha1.ComponentsConfiguration{
				Include: []string{"mockComp2", "mockComp3"},
			}

			Expect(RegisterAllComponents(logr.Discard(), reg, config)).To(Succeed())
			Expect(reg.GenerateBase(options)).To(Succeed())

			Expect(mockComp1.generateBaseCalled).To(BeFalse())
			Expect(mockComp2.generateBaseCalled).To(BeTrue())
			Expect(mockComp3.generateBaseCalled).To(BeTrue())
		})

		It("should return an error if an unknown component is included", func() {
			config.Components = &v1alpha1.ComponentsConfiguration{
				Include: []string{"unknown", "mockComp1", "unknown2"},
			}

			Expect(RegisterAllComponents(logr.Discard(), reg, config)).To(MatchError(And(
				ContainSubstring(`configuration contains invalid component includes`),
				ContainSubstring(`unknown`),
				ContainSubstring(`unknown2`),
				ContainSubstring(`available component names are: mockComp1, mockComp2, mockComp3`),
			)))
		})

		It("should succeed when config is nil", func() {
			Expect(RegisterAllComponents(logr.Discard(), reg, nil)).To((Succeed()))
		})
	})
})

// mockComponent is a test helper that implements components.Interface
type mockComponent struct {
	name                    string
	componentRef            string
	captureCtx              func(components.Context)
	generateBaseCalled      bool
	generateLandscapeCalled bool

	generateBaseFunc      func(components.Options) error
	generateLandscapeFunc func(components.LandscapeOptions) error
}

func (m *mockComponent) GetComponentMetadata() *components.Metadata {
	meta := &components.Metadata{Name: m.name}
	if m.componentRef != "" {
		meta.ComponentRef = &m.componentRef
	}
	return meta
}

func (m *mockComponent) GenerateBase(ctx components.Context, opts components.Options) error {
	m.generateBaseCalled = true
	if m.captureCtx != nil {
		m.captureCtx(ctx)
	}
	if m.generateBaseFunc != nil {
		return m.generateBaseFunc(opts)
	}
	return nil
}

func (m *mockComponent) GenerateLandscape(ctx components.Context, opts components.LandscapeOptions) error {
	m.generateLandscapeCalled = true
	if m.captureCtx != nil {
		m.captureCtx(ctx)
	}
	if m.generateLandscapeFunc != nil {
		return m.generateLandscapeFunc(opts)
	}
	return nil
}
