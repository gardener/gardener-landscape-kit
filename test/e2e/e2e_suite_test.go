// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"testing"

	"github.com/gardener/gardener/pkg/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

var _ = BeforeSuite(func() {
	BasePath = os.Getenv("GLK_BASE_PATH")
	LandscapePath = os.Getenv("GLK_LANDSCAPE_PATH")
	ConfigPath = os.Getenv("GLK_CONFIG_PATH")

	ForgejoURL = getEnvOrDefault("FORGEJO_URL", "http://git.local.gardener.cloud:6080")
	ForgejoUser = getEnvOrDefault("FORGEJO_USER", "gitops")
	ForgejoPassword = getEnvOrDefault("FORGEJO_PASSWORD", "testtest")
	ForgejoOwner = getEnvOrDefault("FORGEJO_OWNER", "gitops")
	ForgejoBaseRepo = getEnvOrDefault("FORGEJO_BASE_REPO", "base")
	ForgejoLandscapeRepo = getEnvOrDefault("FORGEJO_LANDSCAPE_REPO", "test-landscape")
})

func TestE2E(t *testing.T) {
	logf.SetLogger(logger.MustNewZapLogger(logger.InfoLevel, logger.FormatJSON, zap.WriteTo(GinkgoWriter)))
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test E2E GLK Suite")
}
