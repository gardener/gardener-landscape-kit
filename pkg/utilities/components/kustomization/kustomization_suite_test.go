// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package kustomization_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestKustomization(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kustomization Suite")
}
