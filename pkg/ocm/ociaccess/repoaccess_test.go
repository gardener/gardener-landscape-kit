// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package ociaccess

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("repoaccess", func() {
	DescribeTable("#normalizeRepositoryURL",
		func(input, expected string) {
			Expect(normalizeRepositoryURL(input)).To(Equal(expected))
		},
		Entry("empty", "", ""),
		Entry("no scheme no slash", "registry.example.com/path", "registry.example.com/path"),
		Entry("https scheme", "https://registry.example.com/path", "registry.example.com/path"),
		Entry("http scheme", "http://registry.example.com/path", "registry.example.com/path"),
		Entry("trailing slash", "registry.example.com/path/", "registry.example.com/path"),
		Entry("scheme and trailing slash", "https://registry.example.com/path/", "registry.example.com/path"),
		Entry("scheme only no path", "https://registry.example.com", "registry.example.com"),
		Entry("with port", "https://registry.example.com:5000/path", "registry.example.com:5000/path"),
		Entry("leading scheme delimiter", "://nothing", "://nothing"),
	)
})
