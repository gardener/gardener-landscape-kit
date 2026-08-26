// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package oidcservice_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestShootOidcService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Shoot OIDC Service Suite")
}
