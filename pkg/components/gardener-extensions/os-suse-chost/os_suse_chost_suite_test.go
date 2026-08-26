// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package suse_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOSSuseChost(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components OS SUSE CHOST Suite")
}
