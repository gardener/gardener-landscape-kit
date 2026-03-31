// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	. "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
)

var _ = Describe("Component Vector", func() {
	Describe("#NewWithOverride without override", func() {
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
			cv, err := NewWithOverride([]byte(yaml))
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
			cv, err := NewWithOverride([]byte(yaml))
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
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError(`failed to parse base component vector: error converting YAML to JSON: yaml: line 7: could not find expected ':'`))
			Expect(cv).To(BeNil())
		})

		It("should fail with empty components list", func() {
			yaml := `
components: []
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError("invalid base component vector: [].components: Required value: at least one component must be specified"))
			Expect(cv).To(BeNil())
		})

		It("should fail when component name is missing", func() {
			yaml := `
components:
  - name: ""
    sourceRepository: https://github.com/org/repo
    version: 1.0.0
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError("invalid base component vector: [].components[0].name: Required value: component name must not be empty"))
			Expect(cv).To(BeNil())
		})

		It("should fail when source repository is missing", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: ""
    version: 1.0.0
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError("invalid base component vector: [].components[0].sourceRepository: Required value: source repository must not be empty"))
			Expect(cv).To(BeNil())
		})

		It("should fail when version is missing", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo
    version: ""
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError("invalid base component vector: [].components[0].version: Required value: component version must not be empty"))
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
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError(`invalid base component vector: [].components[1].name: Duplicate value: "duplicate"`))
			Expect(cv).To(BeNil())
		})

		It("should fail with invalid source repository URL", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: not-a-valid-url
    version: 1.0.0
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError(`invalid base component vector: [].components[0].sourceRepository: Invalid value: "not-a-valid-url": must have a valid URL scheme (e.g., https, http)`))
			Expect(cv).To(BeNil())
		})

		It("should fail with source repository URL without scheme", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: github.com/org/repo
    version: 1.0.0
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError(`invalid base component vector: [].components[0].sourceRepository: Invalid value: "github.com/org/repo": must have a valid URL scheme (e.g., https, http)`))
			Expect(cv).To(BeNil())
		})

		It("should fail with multiple validation errors", func() {
			yaml := `
components:
  - name: ""
    sourceRepository: invalid-url
    version: ""
`
			cv, err := NewWithOverride([]byte(yaml))
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
			cv, err := NewWithOverride([]byte{}, nil)
			Expect(err).To(MatchError("invalid base component vector: [].components: Required value: at least one component must be specified"))
			Expect(cv).To(BeNil())
		})
	})

	Describe("#NewWithOverride with override", func() {
		const baseYAML = `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
  - name: component2
    sourceRepository: https://github.com/org/repo2
    version: 2.0.0
  - name: component3
    sourceRepository: https://github.com/org/repo3
    version: 3.0.0
`

		It("should override a single component version", func() {
			overrideYAML := `
components:
  - name: component2
    sourceRepository: https://github.com/org/repo2
    version: 2.99.0
`
			cv, err := NewWithOverride([]byte(baseYAML), []byte(overrideYAML))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())

			v, ok := cv.FindComponentVersion("component1")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal("1.0.0"))

			v, ok = cv.FindComponentVersion("component2")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal("2.99.0"), "component2 should be overridden")

			v, ok = cv.FindComponentVersion("component3")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal("3.0.0"))
		})

		It("should override multiple component versions", func() {
			overrideYAML := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.99.0
  - name: component3
    sourceRepository: https://github.com/org/repo3
    version: 3.99.0
`
			cv, err := NewWithOverride([]byte(baseYAML), []byte(overrideYAML))
			Expect(err).NotTo(HaveOccurred())

			v, _ := cv.FindComponentVersion("component1")
			Expect(v).To(Equal("1.99.0"))
			v, _ = cv.FindComponentVersion("component2")
			Expect(v).To(Equal("2.0.0"), "component2 should remain unchanged")
			v, _ = cv.FindComponentVersion("component3")
			Expect(v).To(Equal("3.99.0"))
		})

		It("should override multiple component versions from multiple overrides", func() {
			baseOverrideYAML := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.99.0
  - name: component3
    sourceRepository: https://github.com/org/repo3
    version: 3.99.0
`

			landscapeOverrideYAML := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.100.0
  - name: component2
    sourceRepository: https://github.com/org/repo3
    version: 2.34.0
`
			cv, err := NewWithOverride([]byte(baseYAML), []byte(baseOverrideYAML), []byte(landscapeOverrideYAML))
			Expect(err).NotTo(HaveOccurred())

			v, _ := cv.FindComponentVersion("component1")
			Expect(v).To(Equal("1.100.0"))
			v, _ = cv.FindComponentVersion("component2")
			Expect(v).To(Equal("2.34.0"))
			v, _ = cv.FindComponentVersion("component3")
			Expect(v).To(Equal("3.99.0"))
		})

		It("should append a new component not present in base", func() {
			overrideYAML := `
components:
  - name: new-component
    sourceRepository: https://github.com/org/new
    version: 0.1.0
`
			cv, err := NewWithOverride([]byte(baseYAML), []byte(overrideYAML))
			Expect(err).NotTo(HaveOccurred())

			v, ok := cv.FindComponentVersion("new-component")
			Expect(ok).To(BeTrue())
			Expect(v).To(Equal("0.1.0"))

			// Base components still present
			_, ok = cv.FindComponentVersion("component1")
			Expect(ok).To(BeTrue())
		})

		It("should return base unchanged when override is empty", func() {
			overrideYAML := `components: []`
			cv, err := NewWithOverride([]byte(baseYAML), []byte(overrideYAML))
			Expect(err).NotTo(HaveOccurred())

			Expect(cv.ComponentNames()).To(ConsistOf("component1", "component2", "component3"))
		})

		It("should fail when the base YAML is invalid", func() {
			cv, err := NewWithOverride([]byte(`components: []`), []byte(`components:
  - name: x
    version: 1.0.0`))
			Expect(err).To(HaveOccurred())
			Expect(cv).To(BeNil())
		})

		It("should fail when a new override entry has an empty version", func() {
			overrideYAML := `
components:
  - name: component4
    version: ""
`
			cv, err := NewWithOverride([]byte(baseYAML), []byte(overrideYAML))
			Expect(err).To(HaveOccurred())
			Expect(cv).To(BeNil())
		})

		It("should merge overridden items even with omitted version field", func() {
			overrideYAML := `
components:
  - name: component2
`
			cv, err := NewWithOverride([]byte(baseYAML), []byte(overrideYAML))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())
			version, ok := cv.FindComponentVersion("component2")
			Expect(ok).To(BeTrue())
			Expect(version).To(Equal("2.0.0"))
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
			cv, err = NewWithOverride([]byte(yaml))
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

	Describe("#FindComponentVector", func() {
		var cv Interface

		BeforeEach(func() {
			yaml := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
`
			var err error
			cv, err = NewWithOverride([]byte(yaml))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())
		})

		It("should find the ComponentVector of an existing component", func() {
			component := cv.FindComponentVector("component1")
			Expect(component).NotTo(BeNil())
			Expect(component.Name).To(Equal("component1"))
			Expect(component.SourceRepository).To(Equal(new("https://github.com/org/repo1")))
			Expect(component.Version).To(Equal("1.0.0"))
		})

		It("should return nil when component does not exist", func() {
			component := cv.FindComponentVector("non-existent-component")
			Expect(component).To(BeNil())
		})
	})

	Describe("#ComponentNames", func() {
		var cv Interface

		BeforeEach(func() {
			yaml := `
components:
  - name: component2
    sourceRepository: https://github.com/org/repo2
    version: 2.5.3
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
`
			var err error
			cv, err = NewWithOverride([]byte(yaml))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())
		})

		It("should return the sorted component names", func() {
			Expect(cv.ComponentNames()).To(Equal([]string{"component1", "component2"}))
		})
	})
})

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
			cv, err := NewWithOverride([]byte(yaml))
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
			cv, err := NewWithOverride([]byte(yaml))
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
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError(`failed to parse base component vector: error converting YAML to JSON: yaml: line 7: could not find expected ':'`))
			Expect(cv).To(BeNil())
		})

		It("should fail with empty components list", func() {
			yaml := `
components: []
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError("invalid base component vector: [].components: Required value: at least one component must be specified"))
			Expect(cv).To(BeNil())
		})

		It("should fail when component name is missing", func() {
			yaml := `
components:
  - name: ""
    sourceRepository: https://github.com/org/repo
    version: 1.0.0
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError("invalid base component vector: [].components[0].name: Required value: component name must not be empty"))
			Expect(cv).To(BeNil())
		})

		It("should fail when source repository is missing", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: ""
    version: 1.0.0
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError("invalid base component vector: [].components[0].sourceRepository: Required value: source repository must not be empty"))
			Expect(cv).To(BeNil())
		})

		It("should fail when version is missing", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo
    version: ""
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError("invalid base component vector: [].components[0].version: Required value: component version must not be empty"))
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
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError(`invalid base component vector: [].components[1].name: Duplicate value: "duplicate"`))
			Expect(cv).To(BeNil())
		})

		It("should fail with invalid source repository URL", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: not-a-valid-url
    version: 1.0.0
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError(`invalid base component vector: [].components[0].sourceRepository: Invalid value: "not-a-valid-url": must have a valid URL scheme (e.g., https, http)`))
			Expect(cv).To(BeNil())
		})

		It("should fail with source repository URL without scheme", func() {
			yaml := `
components:
  - name: component1
    sourceRepository: github.com/org/repo
    version: 1.0.0
`
			cv, err := NewWithOverride([]byte(yaml))
			Expect(err).To(MatchError(`invalid base component vector: [].components[0].sourceRepository: Invalid value: "github.com/org/repo": must have a valid URL scheme (e.g., https, http)`))
			Expect(cv).To(BeNil())
		})

		It("should fail with multiple validation errors", func() {
			yaml := `
components:
  - name: ""
    sourceRepository: invalid-url
    version: ""
`
			cv, err := NewWithOverride([]byte(yaml))
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
			cv, err := NewWithOverride([]byte{}, nil)
			Expect(err).To(MatchError("invalid base component vector: [].components: Required value: at least one component must be specified"))
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
			cv, err = NewWithOverride([]byte(yaml))
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

	Describe("#FindComponentVector", func() {
		var cv Interface

		BeforeEach(func() {
			yaml := `
components:
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
`
			var err error
			cv, err = NewWithOverride([]byte(yaml))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())
		})

		It("should find the ComponentVector of an existing component", func() {
			component := cv.FindComponentVector("component1")
			Expect(component).NotTo(BeNil())
			Expect(component.Name).To(Equal("component1"))
			Expect(component.SourceRepository).To(Equal(new("https://github.com/org/repo1")))
			Expect(component.Version).To(Equal("1.0.0"))
		})

		It("should return nil when component does not exist", func() {
			component := cv.FindComponentVector("non-existent-component")
			Expect(component).To(BeNil())
		})
	})

	Describe("#ComponentNames", func() {
		var cv Interface

		BeforeEach(func() {
			yaml := `
components:
  - name: component2
    sourceRepository: https://github.com/org/repo2
    version: 2.5.3
  - name: component1
    sourceRepository: https://github.com/org/repo1
    version: 1.0.0
`
			var err error
			cv, err = NewWithOverride([]byte(yaml))
			Expect(err).NotTo(HaveOccurred())
			Expect(cv).NotTo(BeNil())
		})

		It("should return the sorted component names", func() {
			Expect(cv.ComponentNames()).To(Equal([]string{"component1", "component2"}))
		})
	})

	Describe("#WriteComponentVectorFile", func() {
		const outputDir = "/output"

		// componentNames parses the written components.yaml and returns the list of component names.
		componentNames := func(fs afero.Afero) []string {
			data, err := fs.ReadFile(outputDir + "/components.yaml")
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			var comps struct {
				Components []struct {
					Name string `json:"name"`
				} `json:"components"`
			}
			ExpectWithOffset(1, yaml.Unmarshal(data, &comps)).NotTo(HaveOccurred())
			names := make([]string, 0, len(comps.Components))
			for _, c := range comps.Components {
				names = append(names, c.Name)
			}
			return names
		}

		cv := func(contents []byte) Interface {
			cv, err := NewWithOverride(contents)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			return cv
		}

		BeforeEach(func() {
			componentvector.DefaultComponentsYAML = []byte(`components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
  version: v1.137.1
- name: github.com/gardener/other-component
  sourceRepository: https://github.com/gardener/other-component
  version: v2.0.0
`)
		})

		It("should not produce duplicate entries when the user edits the injected default-version comment", func() {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			overrideCV := []byte(`components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
  version: v1.99.0
- name: github.com/gardener/other-component
  sourceRepository: https://github.com/gardener/other-component
  version: v2.0.0
`)

			// Run 1: write from default CV so no comment is injected.
			Expect(WriteComponentVectorFile(fs, outputDir, cv(componentvector.DefaultComponentsYAML))).To(Succeed())

			// User changes the gardener version.
			writtenFile := outputDir + "/components.yaml"
			writtenData, err := fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(fs.WriteFile(writtenFile,
				[]byte(strings.ReplaceAll(string(writtenData), "version: v1.137.1", "version: v1.99.0")),
				0600)).To(Succeed())

			// Run 2: the injected default-version comment appears.
			Expect(WriteComponentVectorFile(fs, outputDir, cv(overrideCV))).To(Succeed())

			writtenData, err = fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(writtenData)).To(ContainSubstring("# version: v1.137.1 # <-- gardener-landscape-kit version default"))

			// User edits the injected comment (e.g. adds a personal annotation).
			writtenData, err = fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(fs.WriteFile(writtenFile,
				[]byte(strings.ReplaceAll(string(writtenData),
					"# version: v1.137.1 # <-- gardener-landscape-kit version default",
					"# version: v1.137.1 # <-- default (my annotation)")),
				0600)).To(Succeed())

			// Run 3: must not duplicate gardener entry and amend the comment with a new default version comment.
			Expect(WriteComponentVectorFile(fs, outputDir, cv(overrideCV))).To(Succeed())

			Expect(componentNames(fs)).To(ConsistOf(
				"github.com/gardener/gardener",
				"github.com/gardener/other-component",
			))

			// The correct default-version comment must have been restored.
			writtenData, err = fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(writtenData)).To(ContainSubstring("# <-- gardener-landscape-kit version default"))
			Expect(string(writtenData)).To(ContainSubstring("my annotation"))
		})

		It("should bump a component version when GLK updates the default and the user has no override", func() {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			writtenFile := outputDir + "/components.yaml"

			// Run 1: write with the original default (v1.137.1).
			Expect(WriteComponentVectorFile(fs, outputDir, cv(componentvector.DefaultComponentsYAML))).To(Succeed())

			// User removes the other-component entry from components.yaml.
			writtenData, err := fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			idx := strings.Index(string(writtenData), "- name: github.com/gardener/other-component")
			Expect(idx).To(BeNumerically(">", 0))
			Expect(fs.WriteFile(writtenFile, writtenData[:idx], 0600)).To(Succeed())

			// Simulate a GLK version bump: gardener goes from v1.137.1 → v1.138.0.
			componentvector.DefaultComponentsYAML = []byte(`components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
  version: v1.138.0
- name: github.com/gardener/other-component
  sourceRepository: https://github.com/gardener/other-component
  version: v2.0.0
`)

			// Run 2: user has no pin, so the new default version is passed in directly.
			// The three-way merge inside WriteComponentVectorFile must accept the new default version.
			// The user-deleted other-component must not come back.
			Expect(WriteComponentVectorFile(fs, outputDir, cv(componentvector.DefaultComponentsYAML))).To(Succeed())

			writtenData, err = fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(writtenData)).To(ContainSubstring("version: v1.138.0"))
			Expect(string(writtenData)).NotTo(ContainSubstring("version: v1.137.1"))
			Expect(componentNames(fs)).To(ConsistOf("github.com/gardener/gardener"))

			// Run 3 (same GLK version): a subsequent run must be stable.
			Expect(WriteComponentVectorFile(fs, outputDir, cv(componentvector.DefaultComponentsYAML))).To(Succeed())

			writtenData, err = fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(writtenData)).To(ContainSubstring("version: v1.138.0"))
			Expect(string(writtenData)).NotTo(ContainSubstring("version: v1.137.1"))
			Expect(componentNames(fs)).To(ConsistOf("github.com/gardener/gardener"))
		})

		It("should preserve a user-pinned version when GLK updates the default", func() {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			writtenFile := outputDir + "/components.yaml"

			pinnedCV := func() Interface {
				return cv([]byte(`components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
  version: v1.99.0
- name: github.com/gardener/other-component
  sourceRepository: https://github.com/gardener/other-component
  version: v2.0.0
`))
			}

			// Run 1: write with the original default (v1.137.1).
			Expect(WriteComponentVectorFile(fs, outputDir, cv(componentvector.DefaultComponentsYAML))).To(Succeed())

			// User pins gardener to v1.99.0.
			writtenData, err := fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(fs.WriteFile(writtenFile,
				[]byte(strings.ReplaceAll(string(writtenData), "version: v1.137.1", "version: v1.99.0")),
				0600)).To(Succeed())

			// Run 2: GLK picks up the user pin and injects the default-version comment.
			Expect(WriteComponentVectorFile(fs, outputDir, pinnedCV())).To(Succeed())

			writtenData, err = fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(writtenData)).To(ContainSubstring("version: v1.99.0"))
			Expect(string(writtenData)).To(ContainSubstring("# version: v1.137.1 # <-- gardener-landscape-kit version default"))

			// GLK bumps its default: v1.137.1 → v1.138.0.
			componentvector.DefaultComponentsYAML = []byte(`components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
  version: v1.138.0
- name: github.com/gardener/other-component
  sourceRepository: https://github.com/gardener/other-component
  version: v2.0.0
`)

			// Run 3: the default-version comment must be updated to v1.138.0.
			// The user's pin (v1.99.0) must be preserved.
			Expect(WriteComponentVectorFile(fs, outputDir, pinnedCV())).To(Succeed())

			writtenData, err = fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(writtenData)).To(ContainSubstring("version: v1.99.0"))
			Expect(string(writtenData)).To(ContainSubstring("# version: v1.138.0 # <-- gardener-landscape-kit version default"))
			Expect(string(writtenData)).NotTo(ContainSubstring("# version: v1.137.1"))

			// Run 4: must be stable, user pin must not be overwritten.
			Expect(WriteComponentVectorFile(fs, outputDir, pinnedCV())).To(Succeed())

			writtenData, err = fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(writtenData)).To(ContainSubstring("version: v1.99.0"))
			Expect(string(writtenData)).To(ContainSubstring("# version: v1.138.0 # <-- gardener-landscape-kit version default"))
		})

		It("should not re-add entries that the user removed from the file", func() {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			// Run 1: write both entries.
			Expect(WriteComponentVectorFile(fs, outputDir, cv(componentvector.DefaultComponentsYAML))).To(Succeed())

			// User removes the other-component entry entirely.
			writtenFile := outputDir + "/components.yaml"
			writtenData, err := fs.ReadFile(writtenFile)
			Expect(err).NotTo(HaveOccurred())
			idx := strings.Index(string(writtenData), "- name: github.com/gardener/other-component")
			Expect(idx).To(BeNumerically(">", 0))
			Expect(fs.WriteFile(writtenFile, writtenData[:idx], 0600)).To(Succeed())

			// Run 2: same vector — the removed entry must not come back.
			Expect(WriteComponentVectorFile(fs, outputDir, cv(componentvector.DefaultComponentsYAML))).To(Succeed())

			Expect(componentNames(fs)).To(ConsistOf("github.com/gardener/gardener"))
		})
	})
})
