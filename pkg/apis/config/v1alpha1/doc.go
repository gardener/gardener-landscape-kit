// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// +k8s:deepcopy-gen=package
// +k8s:openapi-gen=true
// +k8s:defaulter-gen=TypeMeta

//go:generate crd-ref-docs --source-path=. --config=../../../../hack/api-reference/config.yaml --renderer=markdown --log-level=ERROR --output-path=../../../../docs/api-reference/landscapekit-v1alpha1.md

// +groupName=landscape.config.gardener.cloud
package v1alpha1 // import "github.com/gardener/gardener-landscape-kit/pkg/apis/config/v1alpha1"
