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
	})

	Describe("#WriteObjectsToFilesystem - MergeModeInformative", func() {
		It("should annotate operator-overridden scalar values with the GLK default and preserve user comments idempotently", func() {
			initial := []byte(`apiVersion: v1
kind: ConfigMap
metadata:
  name: test
data:
  version: v1.0.0
`)
			// First generate: establish defaults
			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": initial}, "/landscape", "manifest", fs, configv1alpha1.MergeModeInformative)).To(Succeed())

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
			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": updated}, "/landscape", "manifest", fs, configv1alpha1.MergeModeInformative)).To(Succeed())

			content, err := fs.ReadFile("/landscape/manifest/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			// Operator's override is preserved
			Expect(string(content)).To(ContainSubstring("version: v1.0.5"))
			// User comment is preserved
			Expect(string(content)).To(ContainSubstring("pinned for production"))
			// GLK default annotation is added
			Expect(string(content)).To(ContainSubstring("# glk default: v1.1.0"))
			Expect(string(content)).To(ContainSubstring(meta.GLKManagedMarker))

			// Re-run with the same inputs — annotation and user comment must not be doubled
			Expect(files.WriteObjectsToFilesystem(map[string][]byte{"test.yaml": updated}, "/landscape", "manifest", fs, configv1alpha1.MergeModeInformative)).To(Succeed())

			content2, err := fs.ReadFile("/landscape/manifest/test.yaml")
			Expect(err).NotTo(HaveOccurred())
			Expect(string(content2)).To(Equal(string(content)))
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
			Expect(string(content)).To(ContainSubstring(meta.GLKManagedMarker))

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
			Expect(string(content)).NotTo(ContainSubstring(meta.GLKManagedMarker))
			Expect(string(content)).NotTo(ContainSubstring("# glk default:"))
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
