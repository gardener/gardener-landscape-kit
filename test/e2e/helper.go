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

// forgejoCommand is the base struct offering various command executions against the given server.
type forgejoCommand struct {
	client *forgejo.Client
}

// newForgejoCommand returns a new Forgejo command.
func newForgejoCommand(url, username, password string) *forgejoCommand {
	GinkgoHelper()

	c, err := forgejo.NewClient(url, forgejo.SetBasicAuth(username, password))
	Expect(err).NotTo(HaveOccurred())

	return &forgejoCommand{
		client: c,
	}
}

// pushAndCreatePR creates a new branch by pushing local commits from workDir to the
// remote, then opens a PR. Returns the PR index and branch name.
func (f *forgejoCommand) pushAndCreatePR(branchName, repoName, workDir string) (string, int64) {
	GinkgoHelper()

	session := Git(workDir, "push", "origin", fmt.Sprintf("HEAD:refs/heads/%s", branchName))
	Eventually(session).Should(gexec.Exit(0))

	pr, _, err := f.client.CreatePullRequest(RepoOwner, repoName, forgejo.CreatePullRequestOption{
		Head:  branchName,
		Base:  "main",
		Title: fmt.Sprintf("e2e: generate %s", repoName),
	})
	Expect(err).NotTo(HaveOccurred(), "creating PR in %s", repoName)

	return branchName, pr.Index
}

// waitForActionSuccess polls the Forgejo Actions API until the workflow run triggered by
// the PR on branchName completes successfully, or the context deadline is exceeded.
func (f *forgejoCommand) waitForActionSuccess(ctx context.Context, repoName, branchName, commitSHA string) {
	GinkgoHelper()

	Eventually(ctx, func(g Gomega) {
		runs, _, err := f.client.ListRepoActionRuns(RepoOwner, repoName, forgejo.ListActionRunsOption{
			HeadSHA: commitSHA,
		})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(runs.WorkflowRuns).NotTo(BeEmpty(), "no workflow run found for commit %s - number of runs %d", commitSHA, len(runs.WorkflowRuns))

		g.Expect(runs.WorkflowRuns[0].Status).To(Equal("success"),
			"workflow run %d in %s/%s has status %q", runs.WorkflowRuns[0].ID, repoName, branchName, runs.WorkflowRuns[0].Status)
	}).WithPolling(15 * time.Second).Should(Succeed())
}

// verifyActionCommit verifies that the latest commit on branchName was made by the
// github-actions bot, confirming the workflow committed generated content back to the branch.
func (f *forgejoCommand) verifyActionCommit(repoName, branchName string) {
	GinkgoHelper()

	commits, _, err := f.client.ListRepoCommits(RepoOwner, repoName, forgejo.ListCommitOptions{
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

// mergePR merges the pull request with the given index in the given repo.
func (f *forgejoCommand) mergePR(repoName string, prIndex int64) {
	GinkgoHelper()

	_, _, err := f.client.MergePullRequest(RepoOwner, repoName, prIndex, forgejo.MergePullRequestOption{
		Style:   forgejo.MergeStyleMerge,
		Title:   "e2e: merge generated content",
		Message: "Merge generated content from e2e test",
	})
	Expect(err).NotTo(HaveOccurred(), "merging PR %d in %s", prIndex, repoName)
}

// gitCommand is the base struct offering various git command executions in the given repo.
type gitCommand struct {
	repoPath string
}

// newGitCommand creates a new git command.
func newGitCommand(repoPath string) *gitCommand {
	return &gitCommand{
		repoPath: repoPath,
	}
}

// checkout switches to the given branch, creating it first if createBranch is true.
func (g *gitCommand) checkout(branchName string, createBranch bool) {
	GinkgoHelper()

	args := []string{"checkout"}
	if createBranch {
		args = append(args, "-b")
	}

	session := Git(g.repoPath, append(args, branchName)...)
	Eventually(session).Should(gexec.Exit(0))
}

// add stages the given paths for the next commit.
func (g *gitCommand) add(paths ...string) {
	GinkgoHelper()

	session := Git(g.repoPath, append([]string{"add"}, paths...)...)
	Eventually(session).Should(gexec.Exit(0))
}

// commit creates a commit with the given message, allowing empty commits.
func (g *gitCommand) commit(message string) {
	GinkgoHelper()

	session := Git(g.repoPath, "commit", "--allow-empty", "-m", message)
	Eventually(session).Should(gexec.Exit(0))
}

// headCommit returns the SHA of the current HEAD commit.
func (g *gitCommand) headCommit() string {
	GinkgoHelper()

	session := Git(g.repoPath, "rev-parse", "HEAD")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimRight(string(session.Out.Contents()), "\n")
}

// submoduleUpdate updates the "base" submodule.
func (g *gitCommand) submoduleUpdate() {
	session := Git(g.repoPath, "submodule", "update", "--remote", "--rebase", "base")
	Eventually(session).Should(gexec.Exit(0))
}

// rebase fetches from origin and rebases the current branch onto the given branch.
func (g *gitCommand) rebase(branchName string) {
	session := Git(g.repoPath, "fetch", "origin")
	Eventually(session).Should(gexec.Exit(0))

	session = Git(g.repoPath, "rebase", fmt.Sprintf("origin/%s", branchName))
	Eventually(session).Should(gexec.Exit(0))
}

// push pushes the given branch to origin.
func (g *gitCommand) push(branchName string) {
	session := Git(g.repoPath, "push", "origin", branchName)
	Eventually(session).Should(gexec.Exit(0))
}

// deleteBranch force-deletes branchName from the local repository.
func (g *gitCommand) deleteBranch(branchName string) {
	session := Git(g.repoPath, "branch", "-D", branchName)
	Eventually(session).Should(gexec.Exit(0))
}
