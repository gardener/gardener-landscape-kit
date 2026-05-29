// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package github exposes the subset of this repository's .github/ assets that are intended for re-use in consumer landscape repositories.
// The github GLK component (pkg/components/github) materializes these into a target repo.
package github

import "embed"

// DotGitHubActions embeds the composite actions consumed by the reusable workflows shipped with GLK.
// Only the action(s) usable by GLK consumers are included; CI-only actions (e.g. prepare-release) are intentionally excluded.
//
//go:embed actions/glk
var DotGitHubActions embed.FS

// DotGitHubWorkflows embeds the reusable workflows shipped with GLK for use in consumer landscape repositories.
// CI-only workflows that build/release GLK itself are intentionally excluded.
//
//go:embed workflows/glk.yaml
var DotGitHubWorkflows embed.FS
