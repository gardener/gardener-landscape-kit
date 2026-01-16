// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package componentvector

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"k8s.io/component-base/version"
)

const (
	githubUrlPrefix = "https://github.com"
	glkRepository   = "gardener/gardener-landscape-kit"
	filePath        = "componentvector/components.yaml"
)

// GetReleaseBranchName returns the release branch name based on the current GLK version.
func GetReleaseBranchName() string {
	glkVersion := version.Get()
	return fmt.Sprintf("release-v%s.%s", glkVersion.Major, glkVersion.Minor)
}

// GetDefaultComponentVectorFromGitRepository fetches the latest default component vector file
// from the release branch of the gardener-landscape-kit GitHub repository based on the current GLK version.
func GetDefaultComponentVectorFromGitRepository() ([]byte, error) {
	branch := GetReleaseBranchName()
	return getFileFromGitRepository(githubUrlPrefix+"/"+glkRepository, branch, filePath)
}

func getFileFromGitRepository(repositoryUrl, branch, filePath string) ([]byte, error) {
	rawURL := repositoryUrl
	if rawURL[len(rawURL)-1] == '/' {
		rawURL = rawURL[:len(rawURL)-1]
	}

	if !strings.HasPrefix(rawURL, githubUrlPrefix) {
		return nil, fmt.Errorf("unsupported Git provider for URL: %s", rawURL)
	}
	// Convert GitHub repo URL to raw content URL if necessary.
	// e.g. https://github.com/org/repo -> https://raw.githubusercontent.com/org/repo/branch/path
	rawURL = "https://raw.githubusercontent.com" + rawURL[len(githubUrlPrefix):] + "/" + branch + "/" + filePath

	resp, err := http.Get(rawURL) // #nosec G107 -- This is a GET request to a constrained public GitHub URL.
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file from '%s': %s", rawURL, resp.Status)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
