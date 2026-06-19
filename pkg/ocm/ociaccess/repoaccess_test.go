// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package ociaccess

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RepoAccess", func() {
	DescribeTable("#hostFromURL",
		func(input, expectedHost string, expectErr bool) {
			host, err := hostFromURL(input)
			if expectErr {
				Expect(err).To(HaveOccurred())
				Expect(host).To(BeEmpty())
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(host).To(Equal(expectedHost))
		},
		Entry("https scheme", "https://registry.example.com/path", "registry.example.com", false),
		Entry("http scheme", "http://registry.example.com/path", "registry.example.com", false),
		Entry("https with port", "https://registry.example.com:5000/path", "registry.example.com:5000", false),
		Entry("https scheme only no path", "https://registry.example.com", "registry.example.com", false),
		Entry("trailing slash", "https://registry.example.com/path/", "registry.example.com", false),
		Entry("no scheme", "registry.example.com/path", "", true),
		Entry("empty", "", "", true),
		Entry("malformed", "://nothing", "", true),
	)
})
