// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	generateoptions "github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/registry"
)

var _ = Describe("Registry", func() {
	var (
		reg registry.Interface

		options          components.Options
		landscapeOptions components.LandscapeOptions
	)

	BeforeEach(func() {
		reg = registry.New()

		options = components.NewOptions(
			&generateoptions.Options{Options: &cmd.Options{Log: logr.Discard()}},
			afero.Afero{Fs: afero.NewMemMapFs()},
		)
		landscapeOptions = components.NewLandscapeOptions(
			&generateoptions.Options{
				Options: &cmd.Options{Log: logr.Discard()},
				Config: &v1alpha1.LandscapeKitConfiguration{
					Git: &v1alpha1.GitRepository{},
				},
			},
			afero.Afero{Fs: afero.NewMemMapFs()},
		)
	})

	Describe("#RegisterComponent", func() {
		It("should register components", func() {
			mockComp1 := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}
			mockComp2 := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			reg.RegisterComponent(mockComp1)
			reg.RegisterComponent(mockComp2)

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockComp1.generateBaseCalled).To(BeTrue())
			Expect(mockComp2.generateBaseCalled).To(BeTrue())
		})
	})

	Describe("#GenerateBase", func() {
		It("should successfully generate with no components", func() {
			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should call GenerateBase on all registered components", func() {
			mockComp1 := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}
			mockComp2 := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			reg.RegisterComponent(mockComp1)
			reg.RegisterComponent(mockComp2)

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockComp1.generateBaseCalled).To(BeTrue())
			Expect(mockComp2.generateBaseCalled).To(BeTrue())
		})

		It("should pass options to components", func() {
			var receivedOpts components.Options
			mockComp := &mockComponent{
				generateBaseFunc: func(opts components.Options) error {
					receivedOpts = opts
					return nil
				},
			}

			reg.RegisterComponent(mockComp)

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(receivedOpts).To(Equal(options))
		})

		It("should return error if a component fails", func() {
			expectedErr := errors.New("component error")
			mockComp := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					return expectedErr
				},
			}

			reg.RegisterComponent(mockComp)

			err := reg.GenerateBase(options)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		})

		It("should stop at first error and not call subsequent components", func() {
			expectedErr := errors.New("first component error")
			mockComp1 := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					return expectedErr
				},
			}
			mockComp2 := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
			}

			reg.RegisterComponent(mockComp1)
			reg.RegisterComponent(mockComp2)

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
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return nil
				},
			}
			mockComp2 := &mockComponent{
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return nil
				},
			}

			reg.RegisterComponent(mockComp1)
			reg.RegisterComponent(mockComp2)

			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockComp1.generateLandscapeCalled).To(BeTrue())
			Expect(mockComp2.generateLandscapeCalled).To(BeTrue())
		})

		It("should pass options to components", func() {
			var receivedOpts components.LandscapeOptions
			mockComp := &mockComponent{
				generateLandscapeFunc: func(opts components.LandscapeOptions) error {
					receivedOpts = opts
					return nil
				},
			}

			reg.RegisterComponent(mockComp)

			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).NotTo(HaveOccurred())
			Expect(receivedOpts).To(Equal(landscapeOptions))
		})

		It("should return error if a component fails", func() {
			expectedErr := errors.New("landscape component error")
			mockComp := &mockComponent{
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return expectedErr
				},
			}

			reg.RegisterComponent(mockComp)

			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		})

		It("should stop at first error and not call subsequent components", func() {
			expectedErr := errors.New("first landscape component error")
			mockComp1 := &mockComponent{
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return expectedErr
				},
			}
			mockComp2 := &mockComponent{
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return nil
				},
			}

			reg.RegisterComponent(mockComp1)
			reg.RegisterComponent(mockComp2)

			err := reg.GenerateLandscape(landscapeOptions)
			Expect(err).To(Equal(expectedErr))
			Expect(mockComp1.generateLandscapeCalled).To(BeTrue())
			Expect(mockComp2.generateLandscapeCalled).To(BeFalse())
		})
	})

	Describe("Integration", func() {
		It("should work with components that implement both GenerateBase and GenerateLandscape", func() {
			mockComp := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					return nil
				},
				generateLandscapeFunc: func(_ components.LandscapeOptions) error {
					return nil
				},
			}

			reg.RegisterComponent(mockComp)

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
				generateBaseFunc: func(_ components.Options) error {
					callOrder = append(callOrder, "comp1-base")
					return nil
				},
			}
			mockComp2 := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					callOrder = append(callOrder, "comp2-base")
					return nil
				},
			}
			mockComp3 := &mockComponent{
				generateBaseFunc: func(_ components.Options) error {
					callOrder = append(callOrder, "comp3-base")
					return nil
				},
			}

			reg.RegisterComponent(mockComp1)
			reg.RegisterComponent(mockComp2)
			reg.RegisterComponent(mockComp3)

			err := reg.GenerateBase(options)
			Expect(err).NotTo(HaveOccurred())
			Expect(callOrder).To(Equal([]string{"comp1-base", "comp2-base", "comp3-base"}))
		})
	})
})

// mockComponent is a test helper that implements components.Interface
type mockComponent struct {
	generateBaseFunc        func(components.Options) error
	generateLandscapeFunc   func(components.LandscapeOptions) error
	generateBaseCalled      bool
	generateLandscapeCalled bool
}

func (m *mockComponent) GenerateBase(opts components.Options) error {
	m.generateBaseCalled = true
	if m.generateBaseFunc != nil {
		return m.generateBaseFunc(opts)
	}
	return nil
}

func (m *mockComponent) GenerateLandscape(opts components.LandscapeOptions) error {
	m.generateLandscapeCalled = true
	if m.generateLandscapeFunc != nil {
		return m.generateLandscapeFunc(opts)
	}
	return nil
}
