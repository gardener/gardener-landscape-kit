# Extra Component References

## Background

GLK walks an OCM descriptor tree starting from a configured root component to discover all components that contribute to a landscape. The standard way to add a component to that tree is to declare it as a `componentReferences` entry in the parent descriptor. In some setups it is undesirable or impossible to modify the parent descriptor for that purpose — for example because the parent is published by an upstream pipeline that does not include the additional component.

To still pull such components into GLK's view of the tree, GLK supports a non-standard component label called **`ocm.software/ocm-gear/extra-component-references`**. Components listed in this label are added as dependencies of the labelled component, exactly as if they had been declared via `componentReferences`, but without requiring replication into the upstream descriptor.

This label is **not part of the OCM specification**.

## Schema

The label is attached to a component (not to a resource) and carries a JSON array of references:

```yaml
component:
  name: example.com/my-root
  version: 1.2.3
  labels:
    - name: ocm.software/ocm-gear/extra-component-references
      value:
        - component_reference:
            name: github.com/gardener/diki
            version: v0.25.0
        - component_reference:
            name: example.com/another-component
            version: 0.42.0
```

Each entry has a single `component_reference` object with `name` and `version`. Other fields are ignored.

## Resolution

When GLK loads a component descriptor in `AddComponentDependencies`, it processes the regular `componentReferences` first and then reads the extra-references label. For every entry in the label:

- a dependency keyed by `<name>:<version>` is added to the labelled component
- the referenced component is fetched from the configured OCM repositories and walked recursively, the same way an ordinary `componentReferences` entry would be

Extra references behave like plain dependencies — they do **not** carry an `imageSource`, an alias, or a `LocalName`. If the same component is also declared as a regular `componentReferences` entry, the regular entry wins and the extra reference is silently merged into it.

## Where this is wired up

- The label name is declared as `LabelExtraComponentReferences` in [`pkg/ocm/components/const.go`](../../../pkg/ocm/components/const.go).
- Parsing of the label happens in `extraReferences` in [`pkg/ocm/components/components.go`](../../../pkg/ocm/components/components.go).
- The extra references are merged into the dependency map by `AddComponentDependencies` in [`pkg/ocm/components/components.go`](../../../pkg/ocm/components/components.go), right after the standard `componentReferences` are processed.

## Compatibility note

`extra-component-references` exists to support setups in which the root component descriptor is produced by tooling that cannot be extended with additional `componentReferences` entries. If you control the producer of the descriptor, prefer adding standard `componentReferences` — they are portable to any OCM-aware tool, while this label is interpreted only by GLK.