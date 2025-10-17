// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	_ "embed"
	"time"

	v1 "github.com/controlplaneio-fluxcd/flux-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// FluxInstanceFileName is the name of the flux instance manifest file.
	FluxInstanceFileName = "flux-instance.yaml"
)

func generateFluxInstance() *v1.FluxInstance {
	return &v1.FluxInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FluxInstance",
			APIVersion: "fluxcd.controlplane.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "flux",
			Namespace: "flux-system",
			Annotations: map[string]string{
				"fluxcd.controlplane.io/reconcile":        "enabled",
				"fluxcd.controlplane.io/reconcileEvery":   "1h",
				"fluxcd.controlplane.io/reconcileTimeout": "5m",
			},
		},
		Spec: v1.FluxInstanceSpec{
			Distribution: v1.Distribution{
				Version:  "2.7.x",
				Registry: "ghcr.io/fluxcd",
			},
			Components: []v1.Component{
				"source-controller",
				"kustomize-controller",
				"helm-controller",
				"notification-controller",
				"image-reflector-controller",
				"image-automation-controller",
			},
			Cluster: &v1.Cluster{
				Type:          "kubernetes",
				Size:          "medium",
				Multitenant:   false,
				NetworkPolicy: true,
				Domain:        "cluster.local",
			},
			Storage: &v1.Storage{
				Class: "standard",
				Size:  "10Gi",
			},
			CommonMetadata: &v1.CommonMetadata{
				Labels: map[string]string{
					"app.kubernetes.io/name": "flux",
				},
			},
			Sync: &v1.Sync{
				Kind:       "GitRepository",
				URL:        "https://github.com/<user>/<repo>.git",
				Ref:        "refs/heads/main",
				Path:       RuntimeClusterDirName,
				PullSecret: "flux-system-git-auth",
				Interval:   &metav1.Duration{Duration: 1 * time.Minute},
			},
		},
	}
}
