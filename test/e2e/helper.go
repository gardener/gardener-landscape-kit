// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

// newForgejoClient creates a Forgejo-compatible Gitea SDK client using the package-level variables.
func newForgejoClient() *forgejo.Client {
	GinkgoHelper()
	c, err := forgejo.NewClient(ForgejoURL, forgejo.SetBasicAuth(ForgejoUser, ForgejoPassword))
	Expect(err).NotTo(HaveOccurred())
	return c
}

// forgejoPushAndCreatePR creates a new branch by pushing local commits from workDir to the
// remote, then opens a PR. Returns the PR index and branch name.
func forgejoPushAndCreatePR(c *forgejo.Client, branchName, repoName, workDir string) (string, int64) {
	GinkgoHelper()

	session := Git(workDir, "push", "origin", fmt.Sprintf("HEAD:refs/heads/%s", branchName))
	Eventually(session).Should(gexec.Exit(0))

	pr, _, err := c.CreatePullRequest(ForgejoOwner, repoName, forgejo.CreatePullRequestOption{
		Head:  branchName,
		Base:  "main",
		Title: fmt.Sprintf("e2e: generate %s", repoName),
	})
	Expect(err).NotTo(HaveOccurred(), "creating PR in %s", repoName)

	return branchName, pr.Index
}

// forgejoWaitForActionSuccess polls the Forgejo Actions API until the workflow run triggered by
// the PR on branchName completes successfully, or the context deadline is exceeded.
func forgejoWaitForActionSuccess(ctx context.Context, c *forgejo.Client, repoName, branchName, commitSHA string) {
	GinkgoHelper()

	Eventually(ctx, func(g Gomega) {
		runs, _, err := c.ListRepoActionRuns(ForgejoOwner, repoName, forgejo.ListActionRunsOption{
			HeadSHA: commitSHA,
		})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(runs.WorkflowRuns).NotTo(BeEmpty(), "no workflow run found for commit %s - number of runs %d", commitSHA, len(runs.WorkflowRuns))

		g.Expect(runs.WorkflowRuns[0].Status).To(Equal("success"),
			"workflow run %d in %s/%s has status %q", runs.WorkflowRuns[0].ID, repoName, branchName, runs.WorkflowRuns[0].Status)
	}).WithPolling(15 * time.Second).Should(Succeed())
}

// forgejoVerifyActionCommit verifies that the latest commit on branchName was made by the
// github-actions bot, confirming the workflow committed generated content back to the branch.
func forgejoVerifyActionCommit(ctx context.Context, c *forgejo.Client, repoName, branchName string) {
	GinkgoHelper()

	commits, _, err := c.ListRepoCommits(ForgejoOwner, repoName, forgejo.ListCommitOptions{
		SHA: branchName,
		ListOptions: forgejo.ListOptions{
			Page:     1,
			PageSize: 1,
		},
	})
	Expect(err).NotTo(HaveOccurred())
	Expect(commits).NotTo(BeEmpty())
	Expect(commits[0].RepoCommit.Author.Name).To(Equal("github-actions[bot]"),
		"expected latest commit on %s to be by github-actions[bot], got %s", branchName, commits[0].RepoCommit.Author.Name)
}

// forgejoMergePR merges the pull request with the given index in the given repo.
func forgejoMergePR(c *forgejo.Client, repoName string, prIndex int64) {
	GinkgoHelper()

	_, _, err := c.MergePullRequest(ForgejoOwner, repoName, prIndex, forgejo.MergePullRequestOption{
		Style:   forgejo.MergeStyleMerge,
		Title:   "e2e: merge generated content",
		Message: "Merge generated content from e2e test",
	})
	Expect(err).NotTo(HaveOccurred(), "merging PR %d in %s", prIndex, repoName)
}

type gitCommand struct {
	repoPath string
}

func (g *gitCommand) checkout(branchName string, createBranch bool) {
	GinkgoHelper()

	args := []string{"checkout"}
	if createBranch {
		args = append(args, "-b")
	}

	session := Git(g.repoPath, append(args, branchName)...)
	Eventually(session).Should(gexec.Exit(0))
}

func (g *gitCommand) add(paths ...string) {
	GinkgoHelper()

	session := Git(g.repoPath, append([]string{"add"}, paths...)...)
	Eventually(session).Should(gexec.Exit(0))
}

func (g *gitCommand) commit(message string) {
	GinkgoHelper()

	session := Git(g.repoPath, "commit", "--allow-empty", "-m", message)
	Eventually(session).Should(gexec.Exit(0))
}

func (g *gitCommand) headCommit() string {
	GinkgoHelper()

	session := Git(g.repoPath, "rev-parse", "HEAD")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimRight(string(session.Out.Contents()), "\n")
}

func (g *gitCommand) submoduleUpdate() {
	session := Git(g.repoPath, "submodule", "update", "--remote", "--rebase", "base")
	Eventually(session).Should(gexec.Exit(0))
}

func (g *gitCommand) rebase(branchName string) {
	session := Git(g.repoPath, "fetch", "origin")
	Eventually(session).Should(gexec.Exit(0))

	session = Git(g.repoPath, "rebase", fmt.Sprintf("origin/%s", branchName))
	Eventually(session).Should(gexec.Exit(0))
}

func (g *gitCommand) push(branchName string) {
	session := Git(g.repoPath, "push", "origin", branchName)
	Eventually(session).Should(gexec.Exit(0))
}

func (g *gitCommand) deleteBranch(branchName string) {
	session := Git(g.repoPath, "branch", "-D", branchName)
	Eventually(session).Should(gexec.Exit(0))
}
