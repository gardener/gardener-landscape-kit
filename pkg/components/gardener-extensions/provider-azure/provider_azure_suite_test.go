// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package azure_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProviderAzure(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Provider Azure Suite")
}
