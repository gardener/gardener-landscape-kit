# Custom OCM Components

The Gardener Landscape Kit (GLK) generates Flux kustomizations for public Gardener components. For such included components, GLK can extract versions and image vector overrides automatically from their Open Component Model (OCM) component descriptors.

If you add a custom component to the base or landscape directory, normally you need to deal with version updates yourself.
But if your component is backed by an OCM component descriptor, you can benefit from automatic version and image vector extraction.

### Requirements for a custom OCM Components

1. The component directory contains a file named `component-name` containing the OCM component name (e.g. `my.private-github.com/gardener/my-gardener-extension`).
2. The custom component must be included in the OCM descriptor tree. This means it must be referenced by the root component descriptor or any component referenced directly or indirectly.
3. The component directory contains template files with suffix `.template` (e.g. `my-extension-deployment.yaml.template`). It can reference values for resources or image vector overrides as extracted from the OCM component descriptor.

### How it works

For additional custom OCM components, GLK can inject information for the OCM component descriptor by rendering templates.

GLK needs two steps to extract and use the information from the OCM component descriptors.
1. The command `glk resolve ocm` walks all referenced component descriptors starting from the given root in the GLK configuration file at `.ocm.rootComponent`. It tries to fetch the component descriptors from any of the specified OCI repositories at `.ocm.repositories` and extracts the versions and image vector overrides for all GLK components and all custom OCM components found in the landscape directory tree.

   - The extracted information is written to `components.yaml` in the provided landscape (or base) directory.
2. The command `glk generate [base|landscape]` generates the Flux kustomizations for its active components and additionally renders templates for custom OCM components.

### OCI Registry Authentication

To access private OCI registries, GLK reads credentials from two environment variables:

| Variable | Description |
|---|---|
| `GLK_OCM_REG_USERNAME` | Registry username |
| `GLK_OCM_REG_PASSWORD` | Registry password or token |

Both variables must be set to access private registries. Without them, only anonymous access is attempted.

#### Example: Google Artifact Registry / GCR

```bash
# Authenticate with gcloud and export credentials for GLK
gcloud auth login
eval "$(echo "europe-docker.pkg.dev" | docker-credential-gcloud get | jq -r '"export GLK_OCM_REG_USERNAME=\(.Username)\nexport GLK_OCM_REG_PASSWORD=\(.Secret)"')"
```

> [!NOTE]
> Replace `europe-docker.pkg.dev` with your registry host if it differs.

After exporting the variables, run `glk resolve ocm` as usual.

### Example

Assuming you have a custom OCM component named `my.private-github.com/gardener/my-gardener-extension` stored in a private OCI registry and it is somehow referenced in the OCM descriptor tree starting from the root component descriptor `my.private-github.com/gardener/my-root-compoment`.

1. Adjust the GLK configuration file to include the OCM settings:
```yaml
apiVersion: landscape.config.gardener.cloud/v1alpha1
kind: LandscapeKitConfiguration
ocm:
  rootComponent:
    name: my.private-github.com/gardener/my-root-component
    version: 0.123.0
  repositories:
  - oci://my.private-registry.com/ocm-repo
  - oci://europe-docker.pkg.dev/gardener-project/releases
...
```

2. Create a custom component directory `my-gardener-extension` in your base or landscape directory:
```
components/my-gardener-extension
  ├── component-name
  ├── my-extension-deployment.yaml.template
  ├── kustomization.yaml
  ...
```

Here the file `component-name` contains the OCM component name:
```
my.private-github.com/gardener/my-gardener-extension
```

3. Run the `resolve ocm` command to extract the versions and image vector overrides:
```bash
glk resolve ocm -c landscapekitconfiguration.yaml --target-dir /path/to/landscape-or-base-repo
```

4. Make use of the extracted resource information stored in the file `ocm-components.yaml`.

If your component is a gardener extension, the `my-extension-deployment.yaml.template` may look like this:
```yaml
apiVersion: operator.gardener.cloud/v1alpha1
kind: Extension
metadata:
  name: my-extension
spec:
  deployment:
    extension:
      helm:
        ociRepository:
          ref: {{ .resources.myExtension.helmChart.ref }}
      values:
        {{- if .resources.myExtension.helmChart.imageMap }}
{{ toIndentYAML 2 .resources.myExtension.helmChart.imageMap.gardenerExtensionMyExtension | indent 8 }}
        {{- end }}
        {{- if .imageVectorOverwrite }}
        imageVectorOverwrite: |
{{ indent 10 .imageVectorOverwrite }}
        {{- end }}
```
