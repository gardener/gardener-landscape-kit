// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package virtualgardenaccess_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVirtualGardenAccess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Virtual Garden Access Suite")
}
