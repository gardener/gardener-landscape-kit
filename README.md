# gardener-landscape-kit

[![reuse compliant](https://reuse.software/badge/reuse-compliant.svg)](https://reuse.software/)

> [!WARNING]
> This project is under active development. Breaking changes may occur frequently. Not ready for production use.

## Getting Started

Run the installer without arguments to install the latest GLK release:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/gardener/gardener-landscape-kit/HEAD/install.sh)"
```

To install a specific release, pass its version to the installer:

```bash
GLK_VERSION=v0.6.0
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/gardener/gardener-landscape-kit/HEAD/install.sh)" -- \
  --version "${GLK_VERSION}"
```

The installer downloads the latest release's `component-descriptor.yaml` by default, or
the specified release's descriptor when `--version` is set. It determines the matching
`glk` binary version, installs it to `~/.glk/bin/`, and prints next steps. Using the latest
installer also ensures that descriptors from older releases are handled correctly.

To use a descriptor downloaded manually, get the asset named
`component-descriptor.yaml` from the desired [GitHub release](https://github.com/gardener/gardener-landscape-kit/releases)
and pass it to the installer:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/gardener/gardener-landscape-kit/HEAD/install.sh)" -- \
  --components-file path/to/component-descriptor.yaml
```

`gardener-landscape-kit`, short GLK, is a toolkit for generating manifests to facilitate GitOps-based management of Gardener landscapes.

## Scope

This repository provides a set of tools and templates to help you create and manage Gardener landscapes which includes:
- The generation of a GitOps style directory structure
- The generation of base manifests for Gardener and extensions
- The calculation of [Image Vectors](https://github.com/gardener/gardener/blob/3de34e4283ba25895f6012cfaf8b84271942262d/docs/deployment/image_vector.md) from [OCM Component Descriptors](https://ocm.software/docs/getting-started/component-descriptor-example/)
- The support of migration scenarios of checked-in manifests and deployed resources

## Deployment System

All generated manifests are intended to be applied via [Flux](https://fluxcd.io/), a popular GitOps operator for Kubernetes.
Therefore, the toolkit's components produce Kubernetes manifests that Flux automatically applies, based on the corresponding Flux configuration manifests.

## Configuration Overlays

Besides the deployment system, generated manifests must allow configuration overlays for various landscapes.
For this purpose, the toolkit heavily relies on [Kustomize](https://kustomize.io/), a tool to customize Kubernetes configurations.
The resulting repository structure can be realized in various ways. However, the recommended approach is to use one repository per landscape (see [Repo per environment](https://fluxcd.io/flux/guides/repository-structure/#repo-per-environment)).
