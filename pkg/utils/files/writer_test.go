// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package files_test

import (
	_ "embed"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

var _ = Describe("Writer", func() {
	var (
		fs afero.Afero

		obj     *corev1.ConfigMap
		objYaml []byte
	)

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}

		obj = &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			Data: map[string]string{
				"key": "value",
			},
		}

		var err error
		objYaml, err = yaml.Marshal(obj)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("#WriteObjectsToFilesystem", func() {
		It("should ensure the directories within the path and write the objects", func() {
			objects := map[string][]byte{
				"file.txt":    []byte("This is the file's content"),
				"another.txt": []byte("Some other content"),
			}
			baseDir := "/path/to"
			path := "my/files"

			Expect(files.WriteObjectsToFilesystem(objects, baseDir, path, fs)).To(Succeed())

			contents, err := fs.ReadFile("/path/to/my/files/file.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("This is the file's content\n"))

			contents, err = fs.ReadFile("/path/to/my/files/another.txt")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("Some other content\n"))
		})

		It("should overwrite the manifest file if no meta file is present yet", func() {
			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"config.yaml": objYaml}, "/landscape", "manifest", fs)).To(Succeed())

			content, err := fs.ReadFile("/landscape/.glk/defaults/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(objYaml))

			content, err = fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(objYaml))
		})

		It("should patch only changed default values on subsequent generates and retain custom modifications", func() {
			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"config.yaml": objYaml}, "/landscape", "manifest", fs)).To(Succeed())

			content, err := fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(MatchYAML(objYaml))

			modifiedContent := []byte(strings.ReplaceAll(string(content), "value", "changedValue"))
			Expect(fs.WriteFile("/landscape/manifest/config.yaml", modifiedContent, 0600)).To(Succeed())

			// Patch the default object and generate again
			obj := obj.DeepCopy()
			obj.Data = map[string]string{
				"key":    "value",
				"newKey": "anotherValue",
			}

			objYaml, err = yaml.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())

			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"config.yaml": objYaml}, "/landscape", "manifest", fs)).To(Succeed())

			content, err = fs.ReadFile("/landscape/.glk/defaults/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(objYaml))

			content, err = fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(strings.ReplaceAll(string(objYaml), "key: value", "key: changedValue")))
		})

		It("should add a disclaimer to files containing manifests of kind secret", func() {
			obj.Kind = "Secret"
			objYaml, err := yaml.Marshal(obj)
			Expect(err).NotTo(HaveOccurred())

			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"secret.yaml": objYaml}, "/landscape", "manifest", fs)).To(Succeed())

			content, err := fs.ReadFile("/landscape/manifest/secret.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(And(
				ContainSubstring(`kind: Secret`),
				ContainSubstring(`# SECURITY ADVISORY`),
			))
		})
	})

	Describe("#RelativePathFromDirDepth", func() {
		It("should not go up if the passed relativeDir has no real depth", func() {
			Expect(files.RelativePathFromDirDepth("")).To(Equal("."))
			Expect(files.RelativePathFromDirDepth(".")).To(Equal("."))
			Expect(files.RelativePathFromDirDepth("./")).To(Equal("."))
		})

		It("should go up the depth of the passed relativeDir", func() {
			Expect(files.RelativePathFromDirDepth("examples")).To(Equal("../."))
			Expect(files.RelativePathFromDirDepth("some/directory")).To(Equal("../../."))
			Expect(files.RelativePathFromDirDepth("some/path/to/dir")).To(Equal("../../../../."))
		})

		It("should resolve path switches within relativeDir correctly", func() {
			Expect(files.RelativePathFromDirDepth("enter/../not/../only/destination")).To(Equal("../../."))
			Expect(files.RelativePathFromDirDepth("./././")).To(Equal("."))
			Expect(files.RelativePathFromDirDepth("/")).To(Equal("."))
		})
	})
})
