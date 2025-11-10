// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package components

const (
	// LabelImageVectorImages is the label name for the images mapping
	LabelImageVectorImages = "imagevector.gardener.cloud/images"
	// LabelImageVectorApplication is the label name to mark a component as a source for images of an application, e.g. "kubernetes"
	LabelImageVectorApplication = "imagevector.gardener.cloud/application"
	// LabelImageVectorApplicationValueKubernetes is the label value for the kubernetes application.
	LabelImageVectorApplicationValueKubernetes = "kubernetes"

	labelNameImageVectorName             = "imagevector.gardener.cloud/name"
	labelNameImageVectorRepository       = "imagevector.gardener.cloud/repository"
	labelNameImageVectorSourceRepository = "imagevector.gardener.cloud/source-repository"
	labelNameImageVectorTargetVersion    = "imagevector.gardener.cloud/target-version"
	labelNameCveCategorisation           = "gardener.cloud/cve-categorisation"
	labelNameOriginalRef                 = "cloud.gardener.cnudie/migration/original_ref"
)
