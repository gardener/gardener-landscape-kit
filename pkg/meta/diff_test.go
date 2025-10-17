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

		Expect(fs.MkdirAll("/landscape/.glk", 0700)).To(Succeed())
		Expect(fs.MkdirAll("/landscape/manifest", 0700)).To(Succeed())

		Expect(meta.CreateOrUpdateManifest(obj, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

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

		Expect(fs.MkdirAll("/landscape/.glk", 0700)).To(Succeed())
		Expect(fs.MkdirAll("/landscape/manifest", 0700)).To(Succeed())
		Expect(meta.CreateOrUpdateManifest(obj, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

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

		Expect(meta.CreateOrUpdateManifest(obj, "/landscape", "manifest/config.yaml", fs)).To(Succeed())

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
})
