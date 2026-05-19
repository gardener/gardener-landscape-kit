// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var glkBinaryPath string

// PrepareBinary builds the GLK binary.
func PrepareBinary() {
	By("Building gardener-landscape-kit binary")
	var err error
	glkBinaryPath, err = gexec.Build("../../cmd/gardener-landscape-kit")
	Expect(err).NotTo(HaveOccurred())
	logf.Log.Info("Using binary", "path", glkBinaryPath)

	DeferCleanup(gexec.CleanupBuildArtifacts)
}

// NewCommand creates a new exec.Cmd for gardenadm.
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

// GardenerLandscapeKit runs GLK with the given arguments and returns the gexec.Session.
func GardenerLandscapeKit(args ...string) *gexec.Session {
	return runCommand(NewCommand(glkBinaryPath, ".", append([]string{"--log-level=debug"}, args...)...))
}

// Git runs Git with the given arguments and returns the gexec.Session.
func Git(workDir string, args ...string) *gexec.Session {
	return runCommand(NewCommand("git", workDir, args...))
}
