// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// NewCommand creates a new exec.Cmd.
func NewCommand(binaryPath string, workDir string, args ...string) *exec.Cmd { // #nosec G204 -- Used for e2e tests only.
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workDir
	return cmd
}

// runCommand runs the given exec.Cmd and returns the gexec.Session.
func runCommand(cmd *exec.Cmd) *gexec.Session {
	GinkgoHelper()

	session, err := gexec.Start(
		cmd,
		gexec.NewPrefixedWriter("[out] ", GinkgoWriter),
		gexec.NewPrefixedWriter("[err] ", GinkgoWriter),
	)
	Expect(err).NotTo(HaveOccurred())

	return session
}

// Git runs Git with the given arguments and returns the gexec.Session.
func Git(workDir string, args ...string) *gexec.Session {
	return runCommand(NewCommand("git", workDir, args...))
}
