// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package calico_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNetworkingCalico(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Networking Calico Suite")
}
