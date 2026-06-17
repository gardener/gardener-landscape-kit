// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
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
