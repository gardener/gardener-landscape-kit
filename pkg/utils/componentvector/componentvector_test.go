// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

var _ = Describe("Component Vector", func() {
	Describe("#New", func() {
		It("should successfully create a component vector from valid YAML", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
  - name: component2
    sourceRepository: https://github.com/org/repo2
    version: 2.0.0
`
			cv, err := New([]byte(yaml))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())
		})

		It("should successfully create a component vector with a single component", func() {
			yaml := `
components:
  - name: single-component
    sourceRepository: https://github.com/org/repo
    version: 1.0.0
`
			cv, err := New([]byte(yaml))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())
		})

		It("should fail with invalid YAML syntax", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
  invalid yaml syntax here!!!
`
			cv, err := New([]byte(yaml))
			Expect(err).To(MatchError(`error converting YAML to JSON: yaml: line 7: could not find expected ':'`))
			Expect(cv).To(BeNil())
		})

		It("should fail with empty components list", func() {
			yaml := `
components: []
`
			cv, err := New([]byte(yaml))
			Expect(err).To(MatchError("[].components: Required value: at least one component must be specified"))
			Expect(cv).To(BeNil())
		})

		It("should fail when component name is missing", func() {
			yaml := `
components:
  - name: ""
    sourceRepository: https://github.com/org/repo
    version: 1.0.0
`
			cv, err := New([]byte(yaml))
			Expect(err).To(MatchError("[].components[0].name: Required value: component name must not be empty"))
			Expect(cv).To(BeNil())
		})

		It("should fail when source repository is missing", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: ""
    version: 1.0.0
`
			cv, err := New([]byte(yaml))
			Expect(err).To(MatchError("[].components[0].sourceRepository: Required value: source repository must not be empty"))
			Expect(cv).To(BeNil())
		})

		It("should fail when version is missing", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo
    version: ""
`
			cv, err := New([]byte(yaml))
			Expect(err).To(MatchError("[].components[0].version: Required value: component version must not be empty"))
			Expect(cv).To(BeNil())
		})

		It("should fail with duplicate component names", func() {
			yaml := `
components:
  - name: duplicate
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
  - name: duplicate
    sourceRepository: https://github.com/org/repo2
    version: 2.0.0
`
			cv, err := New([]byte(yaml))
			Expect(err).To(MatchError(`[].components[1].name: Duplicate value: "duplicate"`))
			Expect(cv).To(BeNil())
		})

		It("should fail with invalid source repository URL", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: not-a-valid-url
    version: 1.0.0
`
			cv, err := New([]byte(yaml))
			Expect(err).To(MatchError(`[].components[0].sourceRepository: Invalid value: "not-a-valid-url": must have a valid URL scheme (e.g., https, http)`))
			Expect(cv).To(BeNil())
		})

		It("should fail with source repository URL without scheme", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: github.com/org/repo
    version: 1.0.0
`
			cv, err := New([]byte(yaml))
			Expect(err).To(MatchError(`[].components[0].sourceRepository: Invalid value: "github.com/org/repo": must have a valid URL scheme (e.g., https, http)`))
			Expect(cv).To(BeNil())
		})

		It("should fail with multiple validation errors", func() {
			yaml := `
components:
  - name: ""
    sourceRepository: invalid-url
    version: ""
`
			cv, err := New([]byte(yaml))
			Expect(err).To(HaveOccurred())
			// The aggregate error should contain multiple validation errors
			Expect(err.Error()).To(SatisfyAny(
				ContainSubstring("name"),
				ContainSubstring("sourceRepository"),
				ContainSubstring("version"),
			))
			Expect(cv).To(BeNil())
		})

		It("should fail with empty input", func() {
			cv, err := New([]byte{})
			Expect(err).To(MatchError("[].components: Required value: at least one component must be specified"))
			Expect(cv).To(BeNil())
		})
	})

	Describe("#FindComponentVersion", func() {
		var cv Interface

		BeforeEach(func() {
			yaml := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
  - name: component2
    sourceRepository: https://github.com/org/repo2
    version: 2.5.3
  - name: gardener
    sourceRepository: https://github.com/gardener/gardener
    version: v1.134.1
`
			var err error
			cv, err = New([]byte(yaml))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())
		})

		It("should find the version of an existing component", func() {
			version, exists := cv.FindComponentVersion("component1")
			Expect(exists).To(BeTrue())
			Expect(version).To(Equal("1.0.0"))
		})

		It("should find the version of another existing component", func() {
			version, exists := cv.FindComponentVersion("component2")
			Expect(exists).To(BeTrue())
			Expect(version).To(Equal("2.5.3"))
		})

		It("should find the version of a component with semantic versioning", func() {
			version, exists := cv.FindComponentVersion("gardener")
			Expect(exists).To(BeTrue())
			Expect(version).To(Equal("v1.134.1"))
		})

		It("should return an error when component is not exists", func() {
			version, exists := cv.FindComponentVersion("non-existent-component")
			Expect(exists).To(BeFalse())
			Expect(version).To(Equal(""))
		})

		It("should return an error for empty component name", func() {
			version, exists := cv.FindComponentVersion("")
			Expect(exists).To(BeFalse())
			Expect(version).To(Equal(""))
		})
	})
})
