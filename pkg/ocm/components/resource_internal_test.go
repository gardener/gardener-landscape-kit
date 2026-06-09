// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	descriptorruntime "ocm.software/open-component-model/bindings/go/descriptor/runtime"
	accessv1 "ocm.software/open-component-model/bindings/go/oci/spec/access/v1"
	ocmruntime "ocm.software/open-component-model/bindings/go/runtime"

	"github.com/gardener/gardener-landscape-kit/pkg/ocm/ociaccess"
)

var _ = Describe("#resourceToImageSource", func() {
	newAccessObj := func(typeName, ref string) ocmruntime.Typed {
		switch typeName {
		case accessv1.OCIImageType:
			return &accessv1.OCIImage{
				Type:           ocmruntime.Type{Name: typeName},
				ImageReference: ref,
			}
		case ociaccess.RelativeOciReferenceTypeName:
			return &ociaccess.RelativeOciReference{
				Type:      ocmruntime.Type{Name: typeName},
				Reference: ref,
			}
		}
		Fail("unsupported access type: " + typeName)
		return nil
	}
	newAccess := func(typeName, ref string, raw bool) ocmruntime.Typed {
		obj := newAccessObj(typeName, ref)
		if !raw {
			return obj
		}
		data, err := json.Marshal(obj)
		Expect(err).NotTo(HaveOccurred())
		return &ocmruntime.Raw{
			Type: ocmruntime.Type{Name: typeName},
			Data: data,
		}
	}

	DescribeTable("resolves the image reference",
		func(accessType, accessRef, repoURL, expectedRef, expectedRepo, expectedTag string) {
			for _, raw := range []bool{true, false} {
				res := descriptorruntime.Resource{
					Type:   ResourceTypeOCIImage,
					Access: newAccess(accessType, accessRef, raw),
				}
				res.Name = "my-image"
				res.Version = "1.2.3"

				src, err := resourceToImageSource(res, repoURL)
				Expect(err).NotTo(HaveOccurred())
				Expect(src).NotTo(BeNil())
				Expect(*src.Ref).To(Equal(expectedRef))
				Expect(*src.Repository).To(Equal(expectedRepo))
				Expect(*src.Tag).To(Equal(expectedTag))
			}
		},
		Entry("absolute OCIImage with tag, repoURL ignored",
			accessv1.OCIImageType, "registry.example.com/my-image:1.2.3",
			"ignored.example.com/path",
			"registry.example.com/my-image:1.2.3",
			"registry.example.com/my-image",
			"1.2.3",
		),
		Entry("relative reference with tag, repoURL prepended",
			ociaccess.RelativeOciReferenceTypeName, "my-image:1.2.3",
			"registry.example.com/path",
			"registry.example.com/path/my-image:1.2.3",
			"registry.example.com/path/my-image",
			"1.2.3",
		),
		Entry("relative reference with digest, repoURL prepended",
			ociaccess.RelativeOciReferenceTypeName, "my-image@sha256:abc",
			"registry.example.com/path",
			"registry.example.com/path/my-image@sha256:abc",
			"registry.example.com/path/my-image",
			"sha256:abc",
		),
		Entry("relative reference with leading slash and repoURL with trailing slash, no double slash",
			ociaccess.RelativeOciReferenceTypeName, "/my-image:1.2.3",
			"registry.example.com/path/",
			"registry.example.com/path/my-image:1.2.3",
			"registry.example.com/path/my-image",
			"1.2.3",
		),
	)
})
