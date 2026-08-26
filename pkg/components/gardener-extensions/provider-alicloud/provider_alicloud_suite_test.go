// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package alicloud_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProviderAlicloud(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Provider Alicloud Suite")
}
