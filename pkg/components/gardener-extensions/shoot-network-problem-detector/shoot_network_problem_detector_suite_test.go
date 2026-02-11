// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package shoot_network_problem_detector_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestShootNetworkProblemDetector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Shoot Network Problem Detector Suite")
}
