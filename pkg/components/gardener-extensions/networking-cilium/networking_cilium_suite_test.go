// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package cilium_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNetworkingCilium(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Networking Cilium Suite")
}
