// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package flux

import (
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/pkg/utilities/files"
)

const (
	// NamespaceKind is the kind of the namespace resource.
	NamespaceKind = "Namespace"
	// GardenNamespaceFileName is the name of the namespace manifest file.
	GardenNamespaceFileName = "garden-namespace.yaml"
	// GardenNamespaceName is the name of the namespace created by this component.
	GardenNamespaceName = "garden"

	// FluxSystemNamespaceName is the name of the namespace used by flux components.
	FluxSystemNamespaceName = "flux-system"
)

// GenerateGardenerNamespaceManifest generates the garden namespace in the given landscape directory.
func GenerateGardenerNamespaceManifest(landscapeDir, dirName string, fs afero.Afero) error {
	objects := make(map[string][]byte)

	var err error
	objects[GardenNamespaceFileName], err = yaml.Marshal(&corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       NamespaceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: GardenNamespaceName,
		},
	})
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, landscapeDir, dirName, fs)
}
