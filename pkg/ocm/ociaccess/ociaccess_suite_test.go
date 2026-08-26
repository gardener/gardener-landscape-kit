// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package ociaccess

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOCIAccess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OCI Access Suite")
}
