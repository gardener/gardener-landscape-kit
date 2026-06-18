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

var _ = BeforeSuite(func() {
	BasePath = os.Getenv("GLK_BASE_REPO_PATH")
	LandscapePath = os.Getenv("GLK_LANDSCAPE_REPO_PATH")
	ConfigPath = os.Getenv("GLK_CONFIG_PATH")

	GitServerURL = os.Getenv("GIT_SERVER_URL")
	GitUserName = os.Getenv("GIT_USER_NAME")
	GitUserPassword = os.Getenv("GIT_USER_PASSWORD")
	GLKBaseRepoName = os.Getenv("GLK_BASE_REPO_NAME")
	GLKLandscapeRepoName = os.Getenv("GLK_LANDSCAPE_REPO_NAME")
	RepoOwner = GitUserName
})

func TestE2E(t *testing.T) {
	logf.SetLogger(logger.MustNewZapLogger(logger.InfoLevel, logger.FormatJSON, zap.WriteTo(GinkgoWriter)))
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test E2E GLK Suite")
}
