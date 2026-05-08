# Gardener Landscape Kit Documentation

Welcome to the Gardener Landscape Kit (GLK) documentation. This guide will help you understand and use GLK to manage Gardener landscapes with GitOps best practices.

## Overview

The Gardener Landscape Kit is a toolkit for generating manifests to facilitate GitOps-based management of Gardener landscapes. It provides:
- Generation of a GitOps style directory structure
- Base manifests for Gardener and extensions
- Calculation of Image Vectors from OCM Component Descriptors
- Support for migration scenarios

## Getting Started

- [Glossary](glossary.md) - Key terms and definitions used throughout the documentation

## Core Concepts

Learn about the fundamental concepts and architecture of GLK:

- **[Integration](concepts/integration.md)** - How GLK integrates into landscape management workflows, from initial setup to ongoing automated operations with CI/CD systems
- **[Repositories](concepts/repositories.md)** - Understanding base and landscape repository organization patterns, including monorepo vs. separate repositories with Git submodules

## Usage Guides

Practical guides for working with GLK:

- **[Components](usage/components.md)** - How GLK renders component manifests using values from OCM component descriptors
- **[Component Versions](usage/versions.md)** - Managing component versions and component vector configuration
- **[Custom OCM Components](usage/custom-ocm-components.md)** - Creating and using custom Open Component Model components

## API Reference

- **[LandscapeKit Configuration v1alpha1](api-reference/landscapekit-v1alpha1.md)** - Complete API reference for the LandscapeKit configuration schema

## Development

Resources for contributors and developers:

- **[Tests](development/test.md)** - Information about the test suite, including OCM component extraction tests

## Quick Links

- [Main Repository README](../README.md) - Overview of the project scope and deployment system
- [Gardener Documentation](https://gardener.cloud/docs/) - Official documentation for the Gardener project
- [Open Component Model (OCM)](https://ocm.software/) - Specification for component descriptors
- [Flux Documentation](https://fluxcd.io/) - GitOps deployment engine used by GLK
- [Kustomize Documentation](https://kustomize.io/) - Tool for customizing Kubernetes configurations

## Contributing

GLK is under active development. For issues, feature requests, or contributions, please visit the [GitHub repository](https://github.com/gardener/gardener-landscape-kit).
