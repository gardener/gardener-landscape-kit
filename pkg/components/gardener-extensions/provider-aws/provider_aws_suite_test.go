// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package aws_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProviderAWS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Provider AWS Suite")
}
