// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package helpers

import "encoding/json"

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
