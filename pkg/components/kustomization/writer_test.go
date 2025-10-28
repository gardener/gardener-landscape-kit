// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package kustomization_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-landscape-kit/pkg/components/kustomization"
)

var _ = Describe("Kustomization Handling", func() {
	It("should compute the base path correctly", func() {
		Expect(kustomization.ComputeBasePath("/someBase/path", "/someLandscape/path")).To(Equal("/someBase/path"))
		Expect(kustomization.ComputeBasePath("/sharedPrefix/base", "/sharedPrefix/landscape")).To(Equal("base"))
	})
})
