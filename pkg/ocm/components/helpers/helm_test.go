// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helpers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gardener/gardener-landscape-kit/pkg/ocm/components/helpers"
)

var _ = Describe("helm", func() {
	Describe("#ParseHelmChartImageMap", func() {
		It("should successfully parse valid JSON data", func() {
			jsonData := []byte(`{
				"helmchartResource": {"name": "admission-cilium-application"},
				"imageMapping": [
					{
						"resource": {"name": "gardener-extension-admission-cilium"},
						"repository": "image.repository",
						"tag": "image.tag"
					}
				]
			}`)

			result, err := ParseHelmChartImageMap(jsonData)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.HelmChartResource.Name).To(Equal("admission-cilium-application"))
			Expect(result.ImageMapping).To(HaveLen(1))
			Expect(result.ImageMapping[0].Resource.Name).To(Equal("gardener-extension-admission-cilium"))
			Expect(result.ImageMapping[0].Repository).To(Equal("image.repository"))
			Expect(result.ImageMapping[0].Tag).To(Equal("image.tag"))
		})

		It("should successfully parse JSON with multiple image mappings", func() {
			jsonData := []byte(`{
				"helmchartResource": {"name": "test-chart"},
				"imageMapping": [
					{
						"resource": {"name": "image1"},
						"repository": "repo.path1",
						"tag": "tag.path1"
					},
					{
						"resource": {"name": "image2"},
						"repository": "repo.path2",
						"tag": "tag.path2"
					}
				]
			}`)

			result, err := ParseHelmChartImageMap(jsonData)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.ImageMapping).To(HaveLen(2))
			Expect(result.ImageMapping[0].Resource.Name).To(Equal("image1"))
			Expect(result.ImageMapping[1].Resource.Name).To(Equal("image2"))
		})

		It("should successfully parse JSON with empty image mapping", func() {
			jsonData := []byte(`{
				"helmchartResource": {"name": "test-chart"},
				"imageMapping": []
			}`)

			result, err := ParseHelmChartImageMap(jsonData)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.HelmChartResource.Name).To(Equal("test-chart"))
			Expect(result.ImageMapping).To(BeEmpty())
		})

		It("should return error for invalid JSON", func() {
			jsonData := []byte(`{invalid json}`)

			result, err := ParseHelmChartImageMap(jsonData)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error for empty data", func() {
			jsonData := []byte(``)

			result, err := ParseHelmChartImageMap(jsonData)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error for malformed JSON structure", func() {
			jsonData := []byte(`{"helmchartResource": "not an object"}`)

			result, err := ParseHelmChartImageMap(jsonData)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Describe("#SplitOCIImageReference", func() {
		It("should successfully split a valid OCI image reference", func() {
			ref := "europe-docker.pkg.dev/gardener-project/releases/gardener/admission-cilium:v1.2.3"

			repo, tag, err := SplitOCIImageReference(ref)

			Expect(err).ToNot(HaveOccurred())
			Expect(repo).To(Equal("europe-docker.pkg.dev/gardener-project/releases/gardener/admission-cilium"))
			Expect(tag).To(Equal("v1.2.3"))
		})

		It("should successfully split image reference with port number", func() {
			ref := "localhost:5000/myimage:v1.0.0"

			repo, tag, err := SplitOCIImageReference(ref)

			Expect(err).ToNot(HaveOccurred())
			Expect(repo).To(Equal("localhost:5000/myimage"))
			Expect(tag).To(Equal("v1.0.0"))
		})

		It("should handle reference with colon only at the end", func() {
			ref := "myrepo/myimage:"

			repo, tag, err := SplitOCIImageReference(ref)

			Expect(err).ToNot(HaveOccurred())
			Expect(repo).To(Equal("myrepo/myimage"))
			Expect(tag).To(BeEmpty())
		})

		It("should handle reference with multiple colons by using first colon as separator", func() {
			ref := "localhost:5000/myimage:v1:2:3"

			repo, tag, err := SplitOCIImageReference(ref)

			Expect(err).ToNot(HaveOccurred())
			Expect(repo).To(Equal("localhost:5000/myimage"))
			Expect(tag).To(Equal("v1:2:3"))
		})
	})
})
