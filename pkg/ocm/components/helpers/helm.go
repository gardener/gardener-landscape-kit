// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"encoding/json"
	"fmt"
	"strings"
)

// HelmChartImageMap represents a Helm chart image map.
type HelmChartImageMap struct {
	HelmChartResource HelmChartResource `json:"helmchartResource"`
	ImageMapping      []ImageMapping    `json:"imageMapping"`
}

// HelmChartResource represents a Helm chart resource.
type HelmChartResource struct {
	Name string `json:"name"`
}

// ImageMapping represents an image mapping.
type ImageMapping struct {
	Resource   Resource `json:"resource"`
	Repository string   `json:"repository"`
	Tag        string   `json:"tag"`
}

// Resource represents a resource.
type Resource struct {
	Name string `json:"name"`
}

// ParseHelmChartImageMap parses the given JSON data into a HelmChartImageMap.
func ParseHelmChartImageMap(data []byte) (*HelmChartImageMap, error) {
	var helmChartImageMap HelmChartImageMap
	if err := json.Unmarshal(data, &helmChartImageMap); err != nil {
		return nil, err
	}
	return &helmChartImageMap, nil
}

// SplitOCIImageReference splits an OCI image reference into repository and tag.
func SplitOCIImageReference(ref string) (string, string, error) {
	lastIndex := strings.LastIndex(ref, "/")
	if lastIndex == -1 {
		return "", "", fmt.Errorf("unexpected reference '%s'", ref)
	}
	idx := strings.IndexAny(ref[lastIndex:], ":@")
	if idx == -1 {
		return "", "", fmt.Errorf("unexpected reference '%s'", ref)
	}
	return ref[:lastIndex+idx], ref[lastIndex+idx+1:], nil
}

// RepoTagFromRefOrParts resolves repository and tag from either explicit
// repository+tag fields or by splitting a ref string. It mirrors the field
// layout of imagevector.ImageSource and componentvector.HelmChart / OCIImage.
func RepoTagFromRefOrParts(repository, tag, ref *string) (string, string, error) {
	if repository != nil && tag != nil {
		return *repository, *tag, nil
	}
	if ref != nil {
		return SplitOCIImageReference(*ref)
	}
	return "", "", fmt.Errorf("neither repository+tag nor ref is set")
}
