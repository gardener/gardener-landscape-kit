// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package version_test

import (
	"encoding/json"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/gardener/gardener-landscape-kit/pkg/utils/version"
)

var _ = Describe("Version Metadata", func() {
	var (
		fs         afero.Afero
		targetPath string
	)

	BeforeEach(func() {
		fs = afero.Afero{Fs: afero.NewMemMapFs()}
		targetPath = "/test/target"
	})

	Describe("#WriteVersionMetadata", func() {
		It("should create metadata directory and write version.json", func() {
			err := version.WriteVersionMetadata(targetPath, fs)
			Expect(err).NotTo(HaveOccurred())

			versionFile := filepath.Join(targetPath, ".glk", version.MetaDirName, version.VersionFileName)
			exists, err := fs.Exists(versionFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			content, err := fs.ReadFile(versionFile)
			Expect(err).NotTo(HaveOccurred())

			var versionInfo version.Info
			err = json.Unmarshal(content, &versionInfo)
			Expect(err).NotTo(HaveOccurred())

			Expect(versionInfo.Version).NotTo(BeEmpty())
			Expect(versionInfo.GitVersion).NotTo(BeEmpty())
		})

		It("should write valid JSON with proper formatting", func() {
			err := version.WriteVersionMetadata(targetPath, fs)
			Expect(err).NotTo(HaveOccurred())

			versionFile := filepath.Join(targetPath, ".glk", version.MetaDirName, version.VersionFileName)
			content, err := fs.ReadFile(versionFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(content).NotTo(BeEmpty())
		})
	})

	Describe("#ReadVersionMetadata", func() {
		It("should read version metadata successfully", func() {
			// First write metadata
			err := version.WriteVersionMetadata(targetPath, fs)
			Expect(err).NotTo(HaveOccurred())

			// Then read it back
			metadata, err := version.ReadVersionMetadata(targetPath, fs)
			Expect(err).NotTo(HaveOccurred())
			Expect(metadata).NotTo(BeNil())
			Expect(metadata.Version).NotTo(BeEmpty())
		})

		It("should return error when version file does not exist", func() {
			metadata, err := version.ReadVersionMetadata(targetPath, fs)
			Expect(metadata).To(BeNil())
			Expect(err).To(MatchError(And(
				ContainSubstring("older version of gardener-landscape-kit"),
				ContainSubstring("regenerate the base directory"),
			)))
		})

		It("should return error when version file contains invalid JSON", func() {
			// Create directory and write invalid JSON
			metaDir := filepath.Join(targetPath, ".glk", version.MetaDirName)
			err := fs.MkdirAll(metaDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			versionFile := filepath.Join(metaDir, version.VersionFileName)
			err = fs.WriteFile(versionFile, []byte("invalid json"), 0644)
			Expect(err).NotTo(HaveOccurred())

			metadata, err := version.ReadVersionMetadata(targetPath, fs)
			Expect(metadata).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("failed to parse version metadata")))
		})
	})

	Describe("#ValidateVersionCompatibility", func() {
		It("should allow when landscape version equals base version", func() {
			err := version.ValidateVersionCompatibility("v0.2.0", "v0.2.0")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should allow when landscape version is older than base version", func() {
			err := version.ValidateVersionCompatibility("v0.3.0", "v0.2.0")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should block when landscape version is newer than base version", func() {
			err := version.ValidateVersionCompatibility("v0.2.0", "v0.3.0")
			Expect(err).To(MatchError(And(
				ContainSubstring("landscape generation version (v0.3.0) is newer than base generation version (v0.2.0)"),
				ContainSubstring("regenerate the base directory"),
			)))
		})

		It("should handle dev versions correctly", func() {
			// Dev version of same release should be compatible
			err := version.ValidateVersionCompatibility("v0.2.0", "v0.2.0-dev")
			Expect(err).NotTo(HaveOccurred())

			// Newer dev version should be blocked
			err = version.ValidateVersionCompatibility("v0.2.0", "v0.3.0-dev")
			Expect(err).To(HaveOccurred())
		})

		It("should handle versions without 'v' prefix", func() {
			err := version.ValidateVersionCompatibility("0.2.0", "0.2.0")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle versions with build metadata", func() {
			err := version.ValidateVersionCompatibility("v0.2.0", "v0.2.0+build")
			Expect(err).NotTo(HaveOccurred())

			// Special development version format
			err = version.ValidateVersionCompatibility("v0.2.0", "v0.2.0-master+123a")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error for invalid base version", func() {
			err := version.ValidateVersionCompatibility("invalid", "v0.2.0")
			Expect(err).To(MatchError(ContainSubstring("failed to parse base version")))
		})

		It("should return error for invalid landscape version", func() {
			err := version.ValidateVersionCompatibility("v0.2.0", "invalid")
			Expect(err).To(MatchError(ContainSubstring("failed to parse landscape version")))
		})

		It("should handle complex version comparisons", func() {
			// Major version difference
			err := version.ValidateVersionCompatibility("v1.0.0", "v2.0.0")
			Expect(err).To(HaveOccurred())

			// Minor version difference
			err = version.ValidateVersionCompatibility("v0.5.0", "v0.6.0")
			Expect(err).To(HaveOccurred())

			// Patch version difference
			err = version.ValidateVersionCompatibility("v0.2.1", "v0.2.2")
			Expect(err).To(HaveOccurred())
		})

		It("should allow older landscape with newer base", func() {
			err := version.ValidateVersionCompatibility("v0.5.0", "v0.2.0")
			Expect(err).NotTo(HaveOccurred())

			err = version.ValidateVersionCompatibility("v2.0.0", "v1.0.0")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("#ValidateLandscapeVersionCompatibility", func() {
		It("should validate successfully when versions are compatible", func() {
			// Write metadata with a high version to ensure current version is compatible
			metaDir := filepath.Join(targetPath, ".glk", version.MetaDirName)
			err := fs.MkdirAll(metaDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			metadata := version.Info{
				Version:    "v99.99.99",
				GitVersion: "v99.99.99",
				GitCommit:  "test-commit",
				BuildDate:  "2024-01-01",
				Major:      "99",
				Minor:      "99",
			}

			data, err := json.Marshal(metadata)
			Expect(err).NotTo(HaveOccurred())

			versionFile := filepath.Join(metaDir, version.VersionFileName)
			err = fs.WriteFile(versionFile, data, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = version.ValidateLandscapeVersionCompatibility(targetPath, fs)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return error when version file is missing", func() {
			err := version.ValidateLandscapeVersionCompatibility(targetPath, fs)
			Expect(err).To(MatchError(ContainSubstring("older version of gardener-landscape-kit")))
		})

		It("should return error when base version is older", func() {
			// Write metadata with a very old version to ensure current version is newer
			metaDir := filepath.Join(targetPath, ".glk", version.MetaDirName)
			err := fs.MkdirAll(metaDir, 0755)
			Expect(err).NotTo(HaveOccurred())

			metadata := version.Info{
				Version:    "v0.0.0-alpha",
				GitVersion: "v0.0.0-alpha",
				GitCommit:  "test-commit",
				BuildDate:  "2024-01-01",
				Major:      "0",
				Minor:      "0",
			}

			data, err := json.Marshal(metadata)
			Expect(err).NotTo(HaveOccurred())

			versionFile := filepath.Join(metaDir, version.VersionFileName)
			err = fs.WriteFile(versionFile, data, 0644)
			Expect(err).NotTo(HaveOccurred())

			err = version.ValidateLandscapeVersionCompatibility(targetPath, fs)
			Expect(err).To(MatchError(ContainSubstring("is newer than base generation version")))
		})
	})
})
