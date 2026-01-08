// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package files_test

import (
	"embed"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

var (
	//go:embed testdata
	testdata embed.FS
)

var _ = Describe("Templates", func() {
	Describe("#RenderTemplateFiles", func() {
		It("should render all template files, including subdirectories, correctly", func() {
			objects, err := files.RenderTemplateFiles(testdata, "testdata", map[string]any{
				"key":        "RenderedValue",
				"anotherKey": "anotherValue",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(objects).To(HaveKey("file.txt"))
			Expect(string(objects["file.txt"])).To(ContainSubstring("key: RenderedValue\n"))

			Expect(objects).To(HaveKey("subdirectory/file.txt"))
			Expect(string(objects["subdirectory/file.txt"])).To(ContainSubstring("anotherKey: anotherValue\n"))
		})
	})
})
