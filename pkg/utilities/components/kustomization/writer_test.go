// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package kustomization_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/components/kustomization"
)

var _ = Describe("Writer", func() {
	var fs afero.Afero

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
	})

	Describe("#ComputeBasePath", func() {
		It("should compute the base path correctly", func() {
			Expect(kustomization.ComputeBasePath("/someBase/path", "/someLandscape/path")).To(Equal("/someBase/path"))
			Expect(kustomization.ComputeBasePath("/sharedPrefix/base", "/sharedPrefix/landscape")).To(Equal("base"))
		})
	})

	Describe("#WriteObjectsToFilesystem", func() {
		It("should ensure the directories within the path and write the objects", func() {
			objects := map[string][]byte{
				"file.txt":    []byte("This is the file's content"),
				"another.txt": []byte("Some other content"),
			}
			baseDir := "/path/to"
			path := "my/files"

			Expect(kustomization.WriteObjectsToFilesystem(objects, baseDir, path, fs)).To(Succeed())

			contents, err := fs.ReadFile("/path/to/my/files/file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("This is the file's content\n"))

			contents, err = fs.ReadFile("/path/to/my/files/another.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("Some other content\n"))
		})
	})
})
