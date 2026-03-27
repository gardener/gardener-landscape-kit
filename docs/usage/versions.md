# Component Versions

The Gardener Landscape Kit (GLK) needs to know which versions of Gardener and related components to generate. This document explains how GLK handles component version configuration.

## Component Vector

GLK uses a component vector to determine the versions. A component vector is a YAML file that specifies:
- Component names
- Source repository URLs
- Component versions

### Default Component Vector

GLK includes a default component vector file located at [`componentvector/components.yaml`](../../componentvector/components.yaml). This file is automatically used when no custom vector is configured in the GLK configuration file.

Example of the default component vector structure:

```yaml
components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
  version: v1.134.1
```

#### Updating Default Component Vector from GLK Release Branch

The included default component vector can be updated from the GLK release branch, where patch updates are regularly applied.
GLK can automatically retrieve the updated `components.yaml` file and use the latest versions from the release branch if a specific configuration flag is set in your GLK configuration.
This allows you to keep your component versions up-to-date with the latest patches provided by the maintainers.

You can enable this feature by adding the following configuration to your `LandscapeKitConfiguration`:

```yaml
apiVersion: landscape.config.gardener.cloud/v1alpha1
kind: LandscapeKitConfiguration
versionConfig:
  defaultVersionsUpdateStrategy: ReleaseBranch
```

### Custom Component Vector

You can pin or override component versions by placing a `components.yaml` file in the root of your base or landscape directory.
This file is created (or updated) by running the `resolve plain` command, pointing `--target-dir` at the respective directory:

```bash
gardener-landscape-kit resolve plain \
    --target-dir /path/to/base/dir \
    --config path/to/config-file
```

The file is written to `<target-dir>/components.yaml`. GLK uses the three-way merge strategy when re-writing it, so any versions you have manually pinned are preserved across subsequent `resolve` runs.

To pin a version, simply edit the file and set the desired version for a component. On the next `generate` run GLK will use the pinned version and show the default version as a comment next to it:

```yaml
components:
- name: github.com/gardener/gardener
  sourceRepository: https://github.com/gardener/gardener
# version: v1.134.1 # <-- gardener-landscape-kit version default
  version: v1.133.0
```

> [!IMPORTANT]
> The landscape-level `components.yaml` (if present) overrides the base-level file, which in turn overrides the built-in defaults.

### Prefer components.yaml Over Editing Generated Manifests

The `components.yaml` file is the recommended way to control component versions for your landscape.
While GLK's three-way merge preserves user modifications to generated manifests across runs, editing image tags or other version-specific fields directly in generated manifests is not recommended:
the version pins are scattered across many files, not visible in one place, and the manual modifications are not updated automatically even when a new default would have been a meaningful upgrade.

By maintaining versions in `components.yaml`, your version pins survive regeneration and are visible in one place.

### Custom Components

If your landscape includes components that are not part of the GLK default set, add them to `components.yaml` as well:

```yaml
components:
- name: github.com/my-org/my-component
  sourceRepository: https://github.com/my-org/my-component
  version: v1.2.3
```

GLK will read and preserve custom component entries from `components.yaml` during `resolve` and `generate` runs.
This keeps all version information in one place and makes it easy to update or audit.

### Best Effort Maintenance

The component versions in the default vector file ([`componentvector/components.yaml`](../../componentvector/components.yaml)) are maintained on a **best effort basis**. This means:

- Versions are updated periodically by the GLK maintainers
- Updates may not always be immediate or synchronized with upstream releases
- The default vector provides a convenient starting point but may not always reflect the latest versions

### No Automated Qualification

**Important**: There is currently no automated qualification or testing process to verify that the component versions in the default vector work together correctly. This means:

- The default versions are not guaranteed to be compatible with each other
- No automated tests validate version combinations
- Users should test deployments thoroughly in non-production environments first

### Expected Compatibility

Despite the lack of automated qualification, **the maintained versions are expected to work together most of the time**. The GLK maintainers aim to:

- Select component versions that are known to be compatible
- Update versions based on community feedback and issue reports
- Provide reasonable default versions for typical use cases

## Recommendations

For production usage, consider these best practices:

1. **Use Custom Vectors**: Create and maintain your own component vector file with versions that you have tested and validated for your specific use case

2. **Test Before Deploying**: Always test version combinations in development or staging environments before deploying to production

3. **Pin Specific Versions**: Use explicit version tags rather than relying on default versions for critical production landscapes

4. **Track Updates**: Monitor the GLK repository for updates to the default component vector and evaluate whether to adopt new versions
