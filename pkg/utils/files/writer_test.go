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

	configv1alpha1 "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/meta"
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
				"file.yaml":    []byte("content: This is the file's content"),
				"another.yaml": []byte("content: Some other content"),
			}
			baseDir := "/path/to"
			path := "my/files"

			Expect(files.WriteObjectsToFilesystem(objects, baseDir, path, fs, configv1alpha1.MergeModeSilent)).To(Succeed())

			contents, err := fs.ReadFile("/path/to/my/files/file.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("content: This is the file's content\n"))

			contents, err = fs.ReadFile("/path/to/my/files/another.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("content: Some other content\n"))
		})

		It("should overwrite the manifest file if no meta file is present yet", func() {
			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"config.yaml": objYaml}, "/landscape", "manifest", fs, configv1alpha1.MergeModeSilent)).To(Succeed())

			content, err := fs.ReadFile("/landscape/.glk/defaults/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(objYaml))

			content, err = fs.ReadFile("/landscape/manifest/config.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(string(content)).To(MatchYAML(objYaml))
		})

		It("should patch only changed default values on subsequent generates and retain custom modifications", func() {
			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"config.yaml": objYaml}, "/landscape", "manifest", fs, configv1alpha1.MergeModeSilent)).To(Succeed())

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

			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"config.yaml": objYaml}, "/landscape", "manifest", fs, configv1alpha1.MergeModeSilent)).To(Succeed())

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

			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"secret.yaml": objYaml}, "/landscape", "manifest", fs, configv1alpha1.MergeModeSilent)).To(Succeed())

			content, err := fs.ReadFile("/landscape/manifest/secret.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(And(
				ContainSubstring(`kind: Secret`),
				ContainSubstring(`# SECURITY ADVISORY`),
			))
		})

		It("should not add an encryption advisory disclaimer for secret references only", func() {
			objYaml := []byte(`kind: CustomObject
spec:
  secretRef:
    kind: Secret
    name: my-secret`)

			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"secret.yaml": objYaml}, "/landscape", "manifest", fs, configv1alpha1.MergeModeSilent)).To(Succeed())

			content, err := fs.ReadFile("/landscape/manifest/secret.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(And(
				ContainSubstring(`    kind: Secret`),
				Not(ContainSubstring(`# SECURITY ADVISORY`)),
			))
		})

		DescribeTable("should annotate operator-overwritten values only in Informative mode",
			func(mode configv1alpha1.MergeMode, expectAnnotation bool) {
				initial := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": initial}, "/landscape", "manifest", fs, mode)).To(Succeed())

				// Operator pins to a custom version with a comment explaining why
				Expect(fs.WriteFile("/landscape/manifest/test.yaml", []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.5 # pinned for production
`), 0600)).To(Succeed())

				// GLK ships a new default with a newer version
				updated := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.1.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": updated}, "/landscape", "manifest", fs, mode)).To(Succeed())

				content, err := fs.ReadFile("/landscape/manifest/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("version: v1.0.5"))
				Expect(string(content)).To(ContainSubstring("pinned for production"))

				if expectAnnotation {
					Expect(string(content)).To(ContainSubstring(meta.GLKDefaultPrefix + "v1.1.0"))

					// Re-run with the same default — annotation persists because the user did not remove it.
					Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": updated}, "/landscape", "manifest", fs, mode)).To(Succeed())
					content2, err := fs.ReadFile("/landscape/manifest/test.yaml")
					Expect(err).NotTo(HaveOccurred())
					Expect(string(content2)).To(Equal(string(content)))
				} else {
					Expect(string(content)).NotTo(ContainSubstring(meta.GLKDefaultPrefix))
				}
			},
			Entry("Silent", configv1alpha1.MergeModeSilent, false),
			Entry("Informative", configv1alpha1.MergeModeInformative, true),
		)

		Context("MergeMode Informative", func() {
			It("should not re-add the annotation after the user removed it, until the default changes again", func() {
				initial := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": initial}, "/landscape", "manifest", fs, configv1alpha1.MergeModeInformative)).To(Succeed())

				// Operator pins to v1.0.5
				Expect(fs.WriteFile("/landscape/manifest/test.yaml", []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.5
`), 0600)).To(Succeed())

				// GLK ships v1.1.0 — annotation appears
				v110 := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.1.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": v110}, "/landscape", "manifest", fs, configv1alpha1.MergeModeInformative)).To(Succeed())
				content, err := fs.ReadFile("/landscape/manifest/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring(meta.GLKDefaultPrefix))

				// User acknowledges the annotation and removes it manually
				Expect(fs.WriteFile("/landscape/manifest/test.yaml", []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.5
`), 0600)).To(Succeed())

				// Re-run with the same default — annotation stays removed
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": v110}, "/landscape", "manifest", fs, configv1alpha1.MergeModeInformative)).To(Succeed())
				content, err = fs.ReadFile("/landscape/manifest/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("version: v1.0.5"))
				Expect(string(content)).NotTo(ContainSubstring(meta.GLKDefaultPrefix))

				// GLK ships v1.2.0 — annotation re-appears because the default changed
				v120 := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.2.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": v120}, "/landscape", "manifest", fs, configv1alpha1.MergeModeInformative)).To(Succeed())
				content, err = fs.ReadFile("/landscape/manifest/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("version: v1.0.5"))
				Expect(string(content)).To(ContainSubstring(meta.GLKDefaultPrefix + "v1.2.0"))
			})

			It("should replace the annotation when the GLK default changes again instead of accumulating", func() {
				initial := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": initial}, "/landscape", "accum", fs, configv1alpha1.MergeModeInformative)).To(Succeed())

				// Operator pins to v1.0.5
				Expect(fs.WriteFile("/landscape/accum/test.yaml", []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.5 # pinned for production
`), 0600)).To(Succeed())

				// GLK ships v1.1.0 — annotation appears
				v110 := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.1.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": v110}, "/landscape", "accum", fs, configv1alpha1.MergeModeInformative)).To(Succeed())
				content, err := fs.ReadFile("/landscape/accum/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring(meta.GLKDefaultPrefix + "v1.1.0"))

				// GLK ships v1.2.0 — annotation is replaced, not accumulated
				v120 := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.2.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": v120}, "/landscape", "accum", fs, configv1alpha1.MergeModeInformative)).To(Succeed())
				content, err = fs.ReadFile("/landscape/accum/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("pinned for production"))
				Expect(string(content)).To(ContainSubstring(meta.GLKDefaultPrefix + "v1.2.0"))
				Expect(string(content)).NotTo(ContainSubstring(meta.GLKDefaultPrefix + "v1.1.0"))

				// GLK ships v1.3.0 — again replaced, never more than one annotation
				v130 := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.3.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": v130}, "/landscape", "accum", fs, configv1alpha1.MergeModeInformative)).To(Succeed())
				content, err = fs.ReadFile("/landscape/accum/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring("pinned for production"))
				Expect(string(content)).To(ContainSubstring(meta.GLKDefaultPrefix + "v1.3.0"))
				Expect(string(content)).NotTo(ContainSubstring(meta.GLKDefaultPrefix + "v1.2.0"))
				Expect(string(content)).NotTo(ContainSubstring(meta.GLKDefaultPrefix + "v1.1.0"))
			})

			It("should remove the annotation entirely when the GLK default reverts to the operator's value", func() {
				initial := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": initial}, "/landscape", "revert", fs, configv1alpha1.MergeModeInformative)).To(Succeed())

				// Operator pins to v1.0.5
				Expect(fs.WriteFile("/landscape/revert/test.yaml", []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.5
`), 0600)).To(Succeed())

				// GLK ships v1.1.0 — annotation appears
				updated := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.1.0
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": updated}, "/landscape", "revert", fs, configv1alpha1.MergeModeInformative)).To(Succeed())
				content, err := fs.ReadFile("/landscape/revert/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(ContainSubstring(meta.GLKDefaultPrefix))

				// GLK reverts to v1.0.5 — operator's value now matches the default, no annotation
				reverted := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.5
`)
				Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": reverted}, "/landscape", "revert", fs, configv1alpha1.MergeModeInformative)).To(Succeed())
				content, err = fs.ReadFile("/landscape/revert/test.yaml")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).NotTo(ContainSubstring(meta.GLKDefaultPrefix))
				Expect(string(content)).NotTo(ContainSubstring("# Attention - new default:"))
			})
		})
	})

	Describe("#RelativePathFromDirDepth", func() {
		It("should not go up if the passed relativeDir has no real depth", func() {
			Expect(files.RelativePathFromDirDepth("")).To(Equal("."))
			Expect(files.RelativePathFromDirDepth(".")).To(Equal("."))
			Expect(files.RelativePathFromDirDepth("./")).To(Equal("."))
		})

		It("should go up the depth of the passed relativeDir", func() {
			Expect(files.RelativePathFromDirDepth("examples")).To(Equal(".."))
			Expect(files.RelativePathFromDirDepth("some/directory")).To(Equal("../.."))
			Expect(files.RelativePathFromDirDepth("some/path/to/dir")).To(Equal("../../../.."))
		})

		It("should resolve path switches within relativeDir correctly", func() {
			Expect(files.RelativePathFromDirDepth("enter/../not/../only/destination")).To(Equal("../.."))
			Expect(files.RelativePathFromDirDepth("./././")).To(Equal("."))
			Expect(files.RelativePathFromDirDepth("/")).To(Equal("."))
		})
	})
})
