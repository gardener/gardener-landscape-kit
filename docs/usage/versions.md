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

### Custom Component Vector

You can override the default component vector by specifying a custom vector file in your GLK configuration:

```yaml
apiVersion: landscape.config.gardener.cloud/v1alpha1
kind: LandscapeKitConfiguration
versionConfig:
  componentsVectorFile: ./path/to/your/versions.yaml
```

When a custom component vector file is configured, GLK will use that file instead of the default one. The custom vector file must follow the same structure as the default component vector.

## Version Maintenance and Compatibility

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
