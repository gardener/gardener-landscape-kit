# Relative OCI References

## Background

The standard [Open Component Model (OCM)](https://ocm.software/) defines a fixed set of access types for resources in a component descriptor. For OCI images, the canonical access type is `ociArtifact` (or its predecessor `ociRegistry`), where the resource carries a fully-qualified image reference such as `registry.example.com/path/my-image:1.2.3`.

In addition to the OCM-defined access types, the Gardener Landscape Kit (GLK) supports a non-standard access type called **`relativeOciReference`**. This access type carries only a sub-path (e.g. `my-image:1.2.3` or `charts/my-chart:v0.0.1@sha256:deadbeef`) instead of an absolute image reference. The reference is resolved at extraction time by prepending the URL of the repository the component descriptor was found in.

`relativeOciReference` is **not part of the OCM specification**. GLK supports it solely to remain compatible with component descriptors produced by legacy version of the OCM model.

The legacy OCM CLI documents `relativeOciReference` as an experimental, transitional access method created by the OCI uploader when `preferRelativeAccess` is enabled — see the legacy [OCM CLI config file reference](https://github.com/open-component-model/open-component-model/blob/9070bdbd2d3bb04513bd200a110cedfc0316d676/website/content_versioned/version-legacy/docs/reference/ocm-cli/help/configfile.md?plain=1#L305-L310). It was intended as a temporary stand-in until native local-blob support became available in OCM, and is not part of the current OCM specification.

## Schema

A resource access of type `relativeOciReference` looks like this in the component descriptor:

```yaml
resources:
  - name: my-image
    version: v0.0.1
    type: ociImage
    access:
      type: relativeOciReference
      reference: img/sub-path:v0.0.1@sha256:deadbeef
```

The `reference` field must contain a path with a tag (`name:tag`), a digest (`name@sha256:...`), or both (`name:tag@sha256:...`). It must **not** include a registry host — the host comes from the repository GLK loaded the component descriptor from.

## Resolution

When GLK looks up a component version it remembers the host of the repository the descriptor was found in. While extracting resources from the descriptor, every `relativeOciReference` access is resolved into a fully-qualified image reference using the rule:

```
<repository-base> + "/" + <reference>
```

By default, the repository base is the hostname (with optional port) parsed from the repository URL — the scheme and any path are dropped. This can be overridden by setting `ocm.customRepositoryBase` in the GLK configuration, which replaces the auto-detected host for all `relativeOciReference` resolutions. This is useful when the component descriptor was mirrored to a registry whose URL differs from the original producer's host.

The value must be a bare host without a scheme prefix, and may optionally include a port and/or a path prefix, for example:
- `registry.example.com`
- `registry.example.com:5000`
- `registry.example.com/myproject`

A leading `/` on the relative reference is also stripped to avoid producing a double slash. As a concrete example, given:

- repository URL: `https://registry.example.com/path/to/repo/`
- relative reference: `path/to/repo/img/sub-path:v0.0.1@sha256:deadbeef`

the resolved image reference becomes:

```
registry.example.com/path/to/repo/img/sub-path:v0.0.1@sha256:deadbeef
```

Resolution is applied uniformly to both `ociImage` and `helmChart` resources — Helm OCI charts use the same access shape and the same prepend logic.

If the resource access is still in raw (unparsed) form when GLK encounters it, GLK converts it through its scheme registry, where `relativeOciReference` is registered alongside the standard OCM access types. After conversion the same resolution rule applies. References that cannot be parsed (e.g. missing tag and digest) are rejected with an error.

## Where this is wired up

- The custom type is declared in [`pkg/ocm/ociaccess/relativeocireference.go`](../../../pkg/ocm/ociaccess/relativeocireference.go) and registered with the OCM runtime scheme in [`pkg/ocm/ociaccess/repoaccess.go`](../../../pkg/ocm/ociaccess/repoaccess.go).
- The resolution against the repository host is performed by `extractImageReference` in [`pkg/ocm/components/components.go`](../../../pkg/ocm/components/components.go), which is called from both the OCI image extraction path and the Helm chart extraction path.
- The repository host used as the base is captured by `FindComponentVersion` in [`pkg/ocm/ociaccess/repoaccess.go`](../../../pkg/ocm/ociaccess/repoaccess.go) and returned in `FindComponentVersionResult.RepositoryHost` (extracted from the repository URL via `hostFromURL`).

## Compatibility note

Component descriptors produced directly by the standard OCM tooling will use `ociArtifact` / `ociRegistry` (absolute references) and do **not** require this feature. The `relativeOciReference` support exists only to bridge to legacy produced descriptors and may be removed once those producers have migrated to the standard OCM access types.
