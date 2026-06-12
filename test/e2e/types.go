// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

var (
	// BasePath is the path to the base dir.
	BasePath string
	// LandscapePath is the path to the landscape dir.
	LandscapePath string
	// ConfigPath is the path to the GLK config file.
	ConfigPath string

	// GitServerURL is the base URL of the Forgejo server.
	GitServerURL string
	// GitUserName is the Forgejo user for authentication.
	GitUserName string
	// GitUserPassword is the Forgejo password for authentication.
	GitUserPassword string
	// GLKBaseRepoName is the name of the base repository in Forgejo.
	GLKBaseRepoName string
	// GLKLandscapeRepoName is the name of the landscape repository in Forgejo.
	GLKLandscapeRepoName string
	// RepoOwner is the owner of the Forgejo repositories.
	RepoOwner string
)
