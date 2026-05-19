// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	BasePath = os.Getenv("GLK_BASE_PATH")
	LandscapePath = os.Getenv("GLK_LANDSCAPE_PATH")
	ConfigPath = os.Getenv("GLK_CONFIG_PATH")

	PrepareBinary()
})

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test E2E GLK Suite")
}
