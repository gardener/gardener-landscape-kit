// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	descriptorruntime "ocm.software/open-component-model/bindings/go/descriptor/runtime"
	accessv1 "ocm.software/open-component-model/bindings/go/oci/spec/access/v1"
	ocmruntime "ocm.software/open-component-model/bindings/go/runtime"

	"github.com/gardener/gardener-landscape-kit/pkg/ocm/ociaccess"
)

var _ = Describe("#resourceToImageSource", func() {
	makeResource := func(access ocmruntime.Typed) descriptorruntime.Resource {
		res := descriptorruntime.Resource{
			Type:   ResourceTypeOCIImage,
			Access: access,
		}
		res.ElementMeta.ObjectMeta.Name = "my-image"
		res.ElementMeta.ObjectMeta.Version = "1.2.3"
		return res
	}

	It("resolves an absolute OCIImage reference unchanged", func() {
		res := makeResource(&accessv1.OCIImage{
			Type:           ocmruntime.Type{Name: accessv1.OCIImageType},
			ImageReference: "registry.example.com/my-image:1.2.3",
		})

		src, err := resourceToImageSource(res, "ignored.example.com/path")
		Expect(err).NotTo(HaveOccurred())
		Expect(src).NotTo(BeNil())
		Expect(*src.Ref).To(Equal("registry.example.com/my-image:1.2.3"))
		Expect(*src.Repository).To(Equal("registry.example.com/my-image"))
		Expect(*src.Tag).To(Equal("1.2.3"))
	})

	DescribeTable("resolves a relativeOciReference by prepending the repository URL",
		func(repoURL, relRef, expectedRef, expectedRepo, expectedTag string) {
			res := makeResource(&ociaccess.RelativeOciReferenceType{
				Type:      ocmruntime.Type{Name: ociaccess.RelativeOciReferenceTypeName},
				Reference: relRef,
			})

			src, err := resourceToImageSource(res, repoURL)
			Expect(err).NotTo(HaveOccurred())
			Expect(src).NotTo(BeNil())
			Expect(*src.Ref).To(Equal(expectedRef))
			Expect(*src.Repository).To(Equal(expectedRepo))
			Expect(*src.Tag).To(Equal(expectedTag))
		},
		Entry("simple repo and tag",
			"registry.example.com/path",
			"my-image:1.2.3",
			"registry.example.com/path/my-image:1.2.3",
			"registry.example.com/path/my-image",
			"1.2.3",
		),
		Entry("digest reference",
			"registry.example.com/path",
			"my-image@sha256:abc",
			"registry.example.com/path/my-image@sha256:abc",
			"registry.example.com/path/my-image",
			"sha256:abc",
		),
	)
})
