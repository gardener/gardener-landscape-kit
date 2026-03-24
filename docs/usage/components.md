# Components

## Templates for rendering component manifests using values from OCM component descriptors

The GLK `generate` command typically renders manifests for each component using templates with values provided from a component descriptor of the Open Component Model (OCM).
This allows you to generate Kubernetes manifests for your components with externally provided versions, image overwrites and similar overwrites for Helm charts.

To extract these values, the GLK command `resolve --ocm` must be executed first.
It reads the component descriptors and extracts the versions and image vector overwrites.
The result is written into the `components.yaml` file in the target directory.

For each component, `components.yaml` contains these fields:

- `name` is the OCM component name, which is the value of the `name` field in the OCM component descriptor.
- `version` as extracted from the component descriptor.
- `imageVectorOverwrite` as extracted from the component descriptor, which contains the image overwrites for the component and its referenced components. This includes the image name, repository, tag, and optionally the target version and source repository for the component.
- `resources` contains a map from the component resource name in camelCase to the values as shown below:
  ```yaml
  resources:
    <componentResourceName>:
      ociImage: # if the resource has an OCI image
        ref: <repository>:<tag>[@<digest>] # oci image reference
      helmChart: # if the resource has a Helm chart
        ref: <repository>:<tag>[@<digest>] # Helm chart reference
        imageMap: # if the resource is a Helm chart with need to overwrite Helm values for images repository and tag.
          <imageOverwriteName>:
             ... # calculated Helm values for the image overwrite, which can be directly rendered into Helm values.yaml files
  ```
- `componentImageVectorOverwrites` contains the image vector overwrite images deployed by subcomponents. This is not relevant for most components. Notable exception is the Gardener operator.

For more details about the extracted data, see the `ComponentVector` struct in the package [`pkg/utils/componentvector`](../../pkg/utils/componentvector/types.go).
