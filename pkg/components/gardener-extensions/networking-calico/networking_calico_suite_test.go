// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package networking_calico_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNetworkingCalico(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Networking Calico Suite")
}
