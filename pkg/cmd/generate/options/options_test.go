// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package options_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd"
	"github.com/gardener/gardener-landscape-kit/pkg/cmd/generate/options"
)

var _ = Describe("Options", func() {
	Describe("#ShouldWriteEffectiveComponentsFile", func() {
		modePtr := func(m configv1alpha1.EffectiveComponentsVectorFileMode) *configv1alpha1.EffectiveComponentsVectorFileMode {
			return &m
		}

		DescribeTable("mode resolution",
			func(configured *configv1alpha1.EffectiveComponentsVectorFileMode, queried configv1alpha1.EffectiveComponentsVectorFileMode, expected bool) {
				opts := &options.Options{
					Options: &cmd.Options{},
					Config: &configv1alpha1.LandscapeKitConfiguration{
						VersionConfig: &configv1alpha1.VersionConfiguration{
							WriteEffectiveComponentsVectorFile: configured,
						},
					},
				}
				Expect(opts.ShouldWriteEffectiveComponentsFile(queried)).To(Equal(expected))
			},
			// nil → defaults to Landscape
			Entry("nil config defaults to Landscape: queried Landscape → true",
				nil, configv1alpha1.EffectiveComponentsVectorFileModeLandscape, true),
			Entry("nil config defaults to Landscape: queried Base → false",
				nil, configv1alpha1.EffectiveComponentsVectorFileModeBase, false),

			// None → never writes
			Entry("None: queried Landscape → false",
				modePtr(configv1alpha1.EffectiveComponentsVectorFileModeNone), configv1alpha1.EffectiveComponentsVectorFileModeLandscape, false),
			Entry("None: queried Base → false",
				modePtr(configv1alpha1.EffectiveComponentsVectorFileModeNone), configv1alpha1.EffectiveComponentsVectorFileModeBase, false),

			// Base → only Base
			Entry("Base: queried Base → true",
				modePtr(configv1alpha1.EffectiveComponentsVectorFileModeBase), configv1alpha1.EffectiveComponentsVectorFileModeBase, true),
			Entry("Base: queried Landscape → false",
				modePtr(configv1alpha1.EffectiveComponentsVectorFileModeBase), configv1alpha1.EffectiveComponentsVectorFileModeLandscape, false),

			// Landscape → only Landscape
			Entry("Landscape: queried Landscape → true",
				modePtr(configv1alpha1.EffectiveComponentsVectorFileModeLandscape), configv1alpha1.EffectiveComponentsVectorFileModeLandscape, true),
			Entry("Landscape: queried Base → false",
				modePtr(configv1alpha1.EffectiveComponentsVectorFileModeLandscape), configv1alpha1.EffectiveComponentsVectorFileModeBase, false),

			// Both → always writes
			Entry("Both: queried Landscape → true",
				modePtr(configv1alpha1.EffectiveComponentsVectorFileModeBoth), configv1alpha1.EffectiveComponentsVectorFileModeLandscape, true),
			Entry("Both: queried Base → true",
				modePtr(configv1alpha1.EffectiveComponentsVectorFileModeBoth), configv1alpha1.EffectiveComponentsVectorFileModeBase, true),
		)

		It("should default to Landscape when Config is nil", func() {
			opts := &options.Options{Options: &cmd.Options{}}
			Expect(opts.ShouldWriteEffectiveComponentsFile(configv1alpha1.EffectiveComponentsVectorFileModeLandscape)).To(BeTrue())
			Expect(opts.ShouldWriteEffectiveComponentsFile(configv1alpha1.EffectiveComponentsVectorFileModeBase)).To(BeFalse())
		})

		It("should default to Landscape when VersionConfig is nil", func() {
			opts := &options.Options{
				Options: &cmd.Options{},
				Config:  &configv1alpha1.LandscapeKitConfiguration{},
			}
			Expect(opts.ShouldWriteEffectiveComponentsFile(configv1alpha1.EffectiveComponentsVectorFileModeLandscape)).To(BeTrue())
			Expect(opts.ShouldWriteEffectiveComponentsFile(configv1alpha1.EffectiveComponentsVectorFileModeBase)).To(BeFalse())
		})
	})
})
