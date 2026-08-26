// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package components_test

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

// fakeVector is a minimal componentvector.Interface for testing.
type fakeVector struct {
	versions map[string]string
}

func (f *fakeVector) FindComponentVersion(name string) (string, bool) {
	v, ok := f.versions[name]
	return v, ok
}

func (f *fakeVector) FindComponentVector(_ string) *componentvector.ComponentVector {
	return nil
}

func (f *fakeVector) ComponentNames() []string {
	names := make([]string, 0, len(f.versions))
	for k := range f.versions {
		names = append(names, k)
	}
	return names
}

// fakeMetadata wraps a Metadata value so it satisfies components.MetadataInterface.
type fakeMetadata struct {
	meta *components.Metadata
}

func (f *fakeMetadata) GetComponentMetadata() *components.Metadata { return f.meta }

// adder exposes AddComponentContext which is not part of the Context interface.
type adder interface {
	AddComponentContext(logr.Logger, components.MetadataInterface, componentvector.Interface, componentvector.Interface) error
}

var _ = Describe("Context", func() {
	var (
		log           logr.Logger
		ctx           components.Context
		componentRef  string
		currentVector *fakeVector
		nextVector    *fakeVector
	)

	BeforeEach(func() {
		log = logr.Discard()
		ctx = components.NewContext()
		componentRef = "github.com/gardener/gardener"
		currentVector = &fakeVector{versions: map[string]string{componentRef: "v1.0.0"}}
		nextVector = &fakeVector{versions: map[string]string{componentRef: "v2.0.0"}}
	})

	Describe("#NewComponentsContext", func() {
		It("should return a non-nil Context", func() {
			Expect(ctx).NotTo(BeNil())
		})
	})

	Describe("#AddComponentContext", func() {
		var (
			a    adder
			meta *fakeMetadata
		)

		BeforeEach(func() {
			var ok bool
			a, ok = ctx.(adder)
			Expect(ok).To(BeTrue(), "NewComponentsContext result must implement AddComponentContext")
			meta = &fakeMetadata{meta: &components.Metadata{Name: "gardener", ComponentRef: &componentRef}}
		})

		It("should add a component context with versions from both vectors", func() {
			Expect(a.AddComponentContext(log, meta, currentVector, nextVector)).To(Succeed())

			compCtx, found := ctx.FindComponentContext("gardener")
			Expect(found).To(BeTrue())
			Expect(compCtx.GetUpgradePath().CurrentVersion).To(Equal("v1.0.0"))
			Expect(compCtx.GetUpgradePath().NextVersion).To(Equal("v2.0.0"))
		})

		It("should return an error when a context for the component already exists", func() {
			Expect(a.AddComponentContext(log, meta, currentVector, nextVector)).To(Succeed())
			err := a.AddComponentContext(log, meta, currentVector, nextVector)
			Expect(err).To(MatchError(ContainSubstring("component context for component gardener already exists")))
		})

		It("should set empty version strings when componentRef is nil", func() {
			noRefMeta := &fakeMetadata{meta: &components.Metadata{Name: "no-ref-component"}}
			Expect(a.AddComponentContext(log, noRefMeta, currentVector, nextVector)).To(Succeed())

			compCtx, found := ctx.FindComponentContext("no-ref-component")
			Expect(found).To(BeTrue())
			Expect(compCtx.GetUpgradePath().CurrentVersion).To(BeEmpty())
			Expect(compCtx.GetUpgradePath().NextVersion).To(BeEmpty())
		})

		It("should set empty version strings when the component is absent from the vectors", func() {
			unknownRef := "github.com/gardener/unknown"
			unknownMeta := &fakeMetadata{meta: &components.Metadata{Name: "unknown", ComponentRef: &unknownRef}}
			Expect(a.AddComponentContext(log, unknownMeta, currentVector, nextVector)).To(Succeed())

			compCtx, found := ctx.FindComponentContext("unknown")
			Expect(found).To(BeTrue())
			Expect(compCtx.GetUpgradePath().CurrentVersion).To(BeEmpty())
			Expect(compCtx.GetUpgradePath().NextVersion).To(BeEmpty())
		})

		It("should set empty CurrentVersion when currentVector is nil", func() {
			Expect(a.AddComponentContext(log, meta, nil, nextVector)).To(Succeed())

			compCtx, found := ctx.FindComponentContext("gardener")
			Expect(found).To(BeTrue())
			Expect(compCtx.GetUpgradePath().CurrentVersion).To(BeEmpty())
			Expect(compCtx.GetUpgradePath().NextVersion).To(Equal("v2.0.0"))
		})
	})

	Describe("#FindComponentContext", func() {
		It("should return false when no context has been added", func() {
			_, found := ctx.FindComponentContext("github.com/gardener/gardener")
			Expect(found).To(BeFalse())
		})
	})

	Describe("#Own", func() {
		var (
			a    adder
			meta *fakeMetadata
		)

		BeforeEach(func() {
			var ok bool
			a, ok = ctx.(adder)
			Expect(ok).To(BeTrue())
			meta = &fakeMetadata{meta: &components.Metadata{Name: "gardener", ComponentRef: &componentRef}}
		})

		It("should return the component context for the component", func() {
			Expect(a.AddComponentContext(log, meta, currentVector, nextVector)).To(Succeed())

			compCtx, err := ctx.Own(meta)
			Expect(err).NotTo(HaveOccurred())
			Expect(compCtx.GetUpgradePath().CurrentVersion).To(Equal("v1.0.0"))
			Expect(compCtx.GetUpgradePath().NextVersion).To(Equal("v2.0.0"))
		})

		It("should return an error when no context exists for the component", func() {
			_, err := ctx.Own(meta)
			Expect(err).To(MatchError(ContainSubstring("no component context found for 'gardener'")))
		})
	})
})
