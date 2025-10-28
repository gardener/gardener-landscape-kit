// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package meta_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/meta"
)

var _ = Describe("Meta Dir Config Diff", func() {
	var fs afero.Afero

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
	})

	It("should overwrite the manifest file if no meta file is present yet", func() {
		obj := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			Data: map[string]string{
				"key": "value",
			},
		}

		objYaml, err := yaml.Marshal(obj)
		Expect(err).NotTo(HaveOccurred())

		Expect(meta.CreateOrUpdateManifest(objYaml, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		expectedDefaultOutput := `apiVersion: v1
kind: ConfigMap
data:
  key: value
metadata: {}`

		content, err := fs.ReadFile("/landscape/.glk/defaults/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(MatchYAML(expectedDefaultOutput))

		content, err = fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(MatchYAML(expectedDefaultOutput))
	})

	It("should patch only changed default values on subsequent generates and retain custom modifications", func() {
		obj := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			Data: map[string]string{
				"key": "value",
			},
		}

		objYaml, err := yaml.Marshal(obj)
		Expect(err).NotTo(HaveOccurred())

		Expect(meta.CreateOrUpdateManifest(objYaml, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		// Modify the manifest on disk
		content, err := fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		modifiedContent := []byte(strings.ReplaceAll(string(content), "value", "changedValue"))
		Expect(fs.WriteFile("/landscape/manifest/config.yaml", modifiedContent, 0600)).To(Succeed())

		// Patch the default object and generate again
		obj = obj.DeepCopy()
		obj.Data = map[string]string{
			"key":    "value",
			"newKey": "anotherValue",
		}

		objYaml, err = yaml.Marshal(obj)
		Expect(err).NotTo(HaveOccurred())

		Expect(meta.CreateOrUpdateManifest(objYaml, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		content, err = fs.ReadFile("/landscape/.glk/defaults/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(MatchYAML(`apiVersion: v1
kind: ConfigMap
data:
  key: value
  newKey: anotherValue
metadata: {}`))

		content, err = fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(MatchYAML(`apiVersion: v1
kind: ConfigMap
data:
  key: changedValue
  newKey: anotherValue
metadata: {}`))
	})

	It("should support patching raw yaml manifests with comments", func() {
		initialManifest := []byte(`---
apiVersion: v1
# Comment at kind to be removed later on.
kind: ConfigMap
data:
  # This is the key. Please change it.
  key: value`)
		Expect(meta.CreateOrUpdateManifest(initialManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		content, err := fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal(`apiVersion: v1
# Comment at kind to be removed later on.
kind: ConfigMap
data:
  # This is the key. Please change it.
  key: value
`))

		// Modify the manifest on disk
		Expect(fs.WriteFile("/landscape/manifest/config.yaml", []byte(`---
apiVersion: v1
kind: ConfigMap
# I need to fill in the data:
data:
  # I've changed the value to my needs.
  key: changedValue`), 0600)).To(Succeed())

		// Run the generation again with an updated default manifest
		updatedDefaultManifest := []byte(`---
apiVersion: v1
# Comment at kind to be removed later on.
kind: ConfigMap
# Metadata has been added.
metadata:
  name: my-config
  # This namespace could also be changed.
  namespace: default
# data needs to be filled by user (new comment on this key).
data:
  # Generally, this section needs to be customized. (For this test case: Generating data before and after the former comment which has been changed.)
  # This is the key. Please change it.
  # Once it has been changed it will be reflected.
  key: value`)
		Expect(meta.CreateOrUpdateManifest(updatedDefaultManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		content, err = fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal(`apiVersion: v1
kind: ConfigMap
# Metadata has been added.
metadata:
  name: my-config
  # This namespace could also be changed.
  namespace: default
# I need to fill in the data:
data:
  # I've changed the value to my needs.
  key: changedValue
`))
	})

	It("should support patching raw yaml manifests with comments", func() {
		initialManifest := []byte(`---
apiVersion: v1
# Comment at kind to be removed later on.
kind: ConfigMap
# Please fill in your data here:
data:
  # This is the key. Please change it.
  key: value
  key2: value1
  key3: value2
  key4: value3
`)
		Expect(meta.CreateOrUpdateManifest(initialManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		// Modify the manifest on disk
		Expect(fs.WriteFile("/landscape/manifest/config.yaml", []byte(`---
apiVersion: v1
# Comment at kind to be removed later on.
kind: ConfigMap
# USERCOMMENT HERE
# Please fill in your data here:
data:
  # This is the key. Please change it.
  key: value
  key2: value1
  key3: value2
  key4: value3
`), 0600)).To(Succeed())

		// Run the generation again with an updated default manifest
		updatedDefaultManifest := []byte(`---
apiVersion: v1
# Comment at kind to be removed later on.
kind: ConfigMap
# New metadata node
metadata:
  labels:
    # this foo label needs to be adapted
    foo: bar
# Please fill in your data here:
data:
  # This is the key. Please change it.
  key: value
  key2: value7
  keyNeu: valueTest
`)
		Expect(meta.CreateOrUpdateManifest(updatedDefaultManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		content, err := fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal(`apiVersion: v1
# Comment at kind to be removed later on.
kind: ConfigMap
# New metadata node
metadata:
  labels:
    # this foo label needs to be adapted
    foo: bar
# USERCOMMENT HERE
# Please fill in your data here:
data:
  # This is the key. Please change it.
  key: value
  key2: value7
  keyNeu: valueTest
`))
	})

	It("should retain user modifications and order while patching raw yaml manifests", func() {
		initialManifest := []byte(`# Top-most comment
apiVersion: v1 # Comment in line
# Comment at kind.
kind: ConfigMap
# Comment at data
data:
  # Comment for key
  key: value # Comment in key line
  # Comment after key before key2
  key2: value1
  # Comment after key2 within data
`)
		Expect(meta.CreateOrUpdateManifest(initialManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		// Modify the manifest on disk
		Expect(fs.WriteFile("/landscape/manifest/config.yaml", []byte(`# Top-most comment (modified)
apiVersion: v1 # Comment in line (modified)
# Comment at kind. (modified)
kind: ConfigMap
# Comment at data (modified)
data:
  # Comment for key (modified)
  key: value2 # Comment in key line (modified)
  # Comment after key before key2 (modified)
  key2: value7
  # Comment after key2 within data (modified)`), 0600)).To(Succeed())

		// Run the generation again with an updated default manifest
		updatedDefaultManifest := []byte(`# Top-most comment
apiVersion: v1 # Comment in line
# Comment at kind.
kind: ConfigMap
# Comment at data
data:
  # Comment for key
  key: value # Comment in key line
  # Comment after key before key2
  key3: value9
  # Comment after key2 within data
  # This key is new - key2 has been removed, key3 renamed.
  key4: value10`)
		Expect(meta.CreateOrUpdateManifest(updatedDefaultManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		content, err := fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal(`# Top-most comment (modified)
apiVersion: v1 # Comment in line (modified)
# Comment at kind. (modified)
kind: ConfigMap
# Comment at data (modified)
data:
  # Comment for key (modified)
  key: value2 # Comment in key line (modified)
  # Comment after key before key2
  key3: value9
  # Comment after key2 within data
  # This key is new - key2 has been removed, key3 renamed.
  key4: value10
`))
	})

	It("should handle a non-existent default file gracefully", func() {
		// User manifest is there first, no default file yet
		Expect(fs.WriteFile("/landscape/manifest/config.yaml", []byte(`apiVersion: v1
kind: ConfigMap
# Data will be added later on
data:
  # we have a key already
  key: custom # Just a test
  # this will be only here
  onlyHere: 72`), 0600)).To(Succeed())

		// Run the generation
		defaultManifest := []byte(`# New default comments on existing keys are ignored
apiVersion: v1 # also here
# Also on the kind
kind: ConfigMap
# Maybe on the data (as its content changes)
data:
  # Absolutely on the key as it exists in the user manifest already
  key: value # ignored
  # new key3 should be added incl. comment
  key3: value9`)
		Expect(meta.CreateOrUpdateManifest(defaultManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		// The previous user manifest content should have been retained, only new keys added
		content, err := fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(Equal(`apiVersion: v1
kind: ConfigMap
# Data will be added later on
data:
  # we have a key already
  key: custom # Just a test
  # new key3 should be added incl. comment
  key3: value9
  # this will be only here
  onlyHere: 72
`))
	})

	It("should patch arrays correctly", func() {
		initialManifest := []byte(`apiVersion: v1
kind: ConfigMap
data:
  - one # with comment
  - two
  - three
complex:
  - name: item1
    # with comment here
    value: value1
  - name: item2
    value: value2
  - name: item3
    value: value3
`)
		Expect(meta.CreateOrUpdateManifest(initialManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		Expect(fs.WriteFile("/landscape/manifest/config.yaml", []byte(`apiVersion: v1
kind: ConfigMap
data:
  - one # with comment
  - two
  - threeModified
  - ninety-nine
complex:
  - name: item1
    # with comment here
    value: value1Modified
  - name: item2
    value: value2
  - name: item3
    value: value3Modified
  - name: item10
    value: value10 # with comment
`), 0600)).To(Succeed())

		defaultManifest := []byte(`apiVersion: v1
kind: ConfigMap
data:
  - one # with comment
complex:
  - name: item1
    # with comment here
    value: value1`)
		Expect(meta.CreateOrUpdateManifest(defaultManifest, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

		content, err := fs.ReadFile("/landscape/manifest/config.yaml")
		Expect(err).ToNot(HaveOccurred())

		// Arrays are treated as unordered sets, so order may vary
		// Just verify the content is correct (MatchYAML checks structure, not array order for our purposes)
		// Expected items: data=[threeModified, ninety-nine, one], complex=[item1Modified, item3Modified, item10]
		Expect(string(content)).To(ContainSubstring("threeModified"))
		Expect(string(content)).To(ContainSubstring("ninety-nine"))
		Expect(string(content)).To(ContainSubstring("one # with comment"))
		Expect(string(content)).To(ContainSubstring("value: value1Modified"))
		Expect(string(content)).To(ContainSubstring("value: value3Modified"))
		Expect(string(content)).To(ContainSubstring("value: value10"))
		// Ensure no duplicates or unwanted items
		Expect(string(content)).NotTo(ContainSubstring("two"))
		Expect(string(content)).NotTo(ContainSubstring("three\n"))
		Expect(string(content)).NotTo(ContainSubstring("item2"))
		Expect(string(content)).NotTo(ContainSubstring("value: value1\n"))
	})
})
