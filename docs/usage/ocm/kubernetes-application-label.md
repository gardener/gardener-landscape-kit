# Kubernetes Application Label

## Background

Gardener usually specifies component and image versions in the image vector (see [imagevector/containers.yaml](https://github.com/gardener/gardener/blob/master/imagevector/containers.yaml)). Kubernetes images such as `kube-apiserver`, `kube-controller-manager`, or the `hyperkube` family are the exception: their versions are derived from the desired Kubernetes version of a shoot cluster, not pinned in the vector. The image vector attached to components that consume these images (for example `gardenlet`) therefore reference them by *repository* only and expect the deployer to supply the matching *tag* — the Kubernetes version — for every supported Kubernetes minor. This means the overwrites cannot simply be taken from the image vector and require special handling.

To bridge this gap, GLK expects the Kubernetes image versions and digests to be carried by a dedicated OCM component, referred to here as the "Kubernetes component". This component is **not provided by Gardener out of the box**; it is a custom component that the landscape operator must build and maintain themselves, listing one resource per Kubernetes image (e.g. `kube-apiserver`, `kube-controller-manager`, `hyperkube`) for every supported Kubernetes minor version, with the Kubernetes version encoded in the resource's `version` field.

To wire this component into the descriptor tree, GLK supports a non-standard component label called **`imagevector.gardener.cloud/application`**. A component labelled with the value `kubernetes` is treated as the authoritative source of Kubernetes image versions for the entire descriptor tree.

This label is **not part of the OCM specification**.

## Schema

The label is attached to a component (not to a resource):

```yaml
component:
  name: example.com/kubernetes-root-example
  version: 0.1499.0
  labels:
    - name: imagevector.gardener.cloud/application
      value: kubernetes
  resources:
    - name: kube-apiserver
      version: 1.35.3
      labels:
        - name: imagevector.gardener.cloud/name
          value: registry.k8s.io/kube-apiserver
      type: ociImage
      relation: external
      access:
        imageReference: repo.example.com/registry_k8s_io/kube-apiserver:v1.35.3@sha256:8888888855555a3f57748c600c8cdec6279b996cccfbe35272c8e5b87b69f5cb
        type: ociRegistry  
    ... # more resources for kube-controller-manager, hyperkube, etc.

```

Only the value `kubernetes` is recognized today; any other value is ignored.

## Resolution

While walking the descriptor tree, GLK records exactly **one** component as the Kubernetes component:

1. Each component is inspected for the `imagevector.gardener.cloud/application` label.
2. If the value equals `kubernetes`, the component reference is stored as the tree's Kubernetes component.
3. Encountering a second component with the same label value is treated as a configuration error (`non-unique kubernetes component`).

When the image vector for any other component is later assembled, GLK consults the Kubernetes component to resolve "mapped" images (those carried in the regular `imagevector.gardener.cloud/images` label). The matching rule is:

- For each entry in the consumer's image-vector mapping, the entry's `repository` is looked up in the Kubernetes component's resources by repository.
- When a match is found, the Kubernetes resource is cloned, renamed to the consumer's `name`, and the Kubernetes resource's `version` becomes the cloned image's `targetVersion`.
- The cloned entry is appended to the consumer's image vector.

The result is an image vector that pairs each consumer-side image name with every supported Kubernetes version, ready to be rendered into Gardener's `imageVectorOverwrite`.

If a component declares mapped images but no `application: kubernetes` component was found in the tree, GLK fails with `could not determine kubernetes component reference`.

## Where this is wired up

- The label name and the recognized value are declared as `LabelImageVectorApplication` and `LabelImageVectorApplicationValueKubernetes` in [`pkg/ocm/components/const.go`](../../../pkg/ocm/components/const.go).
- The label is detected during component ingestion in `addComponent` in [`pkg/ocm/components/components.go`](../../../pkg/ocm/components/components.go), which records the unique Kubernetes component reference and rejects duplicates.
- The mapping is applied by `GetImageVector` in [`pkg/ocm/components/components.go`](../../../pkg/ocm/components/components.go), using `getKubernetesComponentRef` to locate the Kubernetes component.

## Compatibility note

The `imagevector.gardener.cloud/application` label is a GLK-internal convention used to model Gardener's "Kubernetes image versions live in a separate OCM component" topology. Component descriptors that do not need cross-component Kubernetes-version mapping can omit it entirely. If your descriptor tree contains no mapped-image references (i.e. no `imagevector.gardener.cloud/images` entries that depend on Kubernetes versions), the absence of an `application: kubernetes` component is harmless.