# Extra Component References

## Background

GLK walks an OCM descriptor tree starting from a configured root component to discover all components that contribute to a landscape. The standard way to add a component to that tree is to declare it as a `componentReferences` entry in the parent descriptor. Such standard references are replicated together with the parent: they are transported when the component is transported between OCM repositories and also show up in the bill of materials (BOM).

In some setups it is desirable to pull a component into GLK's view of the tree **without** replicating it — for example for experimental components.

For this case GLK supports a non-standard component label called **`ocm.software/ocm-gear/extra-component-references`**. Components listed in this label are added to GLK's dependency graph just like ordinary `componentReferences`, but are **not replicated**: they do not appear in the BOM and are not transported with the parent.

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
