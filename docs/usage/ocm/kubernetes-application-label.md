# Kubernetes Application Label

## Background

Some Gardener components (for example `gardenlet`) are deployed alongside Kubernetes images such as `pause`, `etcd`, or the `hyperkube` family. The image vector overwrites attached to those components reference these images by *repository* and expect the deployer to supply the matching *tag* — the Kubernetes version — for every supported Kubernetes minor.

The Kubernetes versions and digests are carried by a separate OCM component (the "Kubernetes component"). To wire the two together, GLK supports a non-standard component label called **`imagevector.gardener.cloud/application`**. A component labelled with the value `kubernetes` is treated as the authoritative source of Kubernetes image versions for the entire descriptor tree.

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