# Repositories

Git repositories are an essential part for landscape and configuration management. In essence, the landscape kit differentiates between two types of repositories:
- **Base Repository**: This repository contains the core landscape configurations, modules, and shared resources that are common across multiple landscapes.
- **Landscape Repository**: These repositories are specific to individual landscapes and typically contain overlay configurations that are merged with the ones found in the base repository.

Also see the Kustomize [base and overlay documentation](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/#bases-and-overlays) for more information.

Technically, the Gardener Landscape Kit doesn't use Git directly, but prepares and operates on the directory structure of the local filesystem, that can later be checked in and committed via Git by the user.

## Organization

Whenever Gardener Landscape Kit generates landscape assets, it must have access to both the base repository and the landscape repository. This requirement also applies to the Flux deployment system.
The separation of base and landscape repositories not only allows for a better modularization, but also helps to test changes, like version updates, through different landscape stages (e.g. `development`, `staging`, `production`).

![Base Graduation Diagram](content/base-graduation.png)

To facilitate this, a reference from the landscape repository to the base repository is necessary. This reference can be organized in two ways:

### Monorepo

> [!NOTE]
> A monorepo is not recommended for production usage.

Both the base repository (directories) and landscape repository (directories) are maintained within a single Git repository. The landscape repository contains the base repository as a subdirectory.

### Separate Repositories (Submodule) — Recommended
The base repository and landscape repository are maintained as two distinct Git repositories. The landscape repository includes a reference to the base repository through a [Git submodule](https://git-scm.com/book/en/v2/Git-Tools-Submodules).
In a setup with multiple landscape stages (e.g. `development`, `staging`, `production`), each stage can reference a different commit, branch, or tag of the base repository through its submodule configuration.
This way, changes in the base repository can graduate through the different landscape stages before being applied to production.

Flux natively supports Git submodules via the `recurseSubmodules` field on the [`GitRepository`](https://fluxcd.io/flux/components/source/gitrepositories/) resource, which GLK enables by default in the generated Flux sync configuration.

> [!IMPORTANT]
> Do **not** use the Flux GitRepository [`include`](https://fluxcd.io/flux/components/source/gitrepositories/#include) mechanism to combine base and landscape repositories. The `include` artifact is assembled asynchronously — when Flux reconciles, the included repository may not yet reflect the latest commit. This can cause the landscape to be applied with a stale base, leading to unintended or inconsistent states. Git submodules avoid this problem because Flux fetches the submodule content atomically together with the parent repository.

## Generation

The repository layout must be considered when executing GLK's `generate` command.
Generating the base repository is as simple as providing the path to the base repository directory, e.g.:

```bash
glk generate base -c config-file /path/to/base-repo
```

The generation of the landscape repositories requires the config to describe both the base and landscape repository.
For each repository, every path is interpreted relative to *its own* repository root:

```yaml
apiVersion: landscape.config.gardener.cloud/v1alpha1
kind: LandscapeKitConfiguration
repositories:
  base:
    target: ./config
  landscape:
    url: https://github.tools.sap/d066080/test-landscape-dev
    ref:
      branch: main
    baseLink: ./base
    target: ./
```

```bash
glk generate landscape -c config-file /path/to/landscape-repo
```

The fields are anchored as follows:
- `repositories.base.target` is the directory inside the **base** repository that holds the generated base content (the output of `glk generate base`).
- `repositories.landscape.url` and `repositories.landscape.ref` identify the **landscape** Git repository and the ref to check out.
- `repositories.landscape.target` is the directory inside the **landscape** repository that holds the landscape configuration (e.g. `./` when the landscape lives at the repo root, or `./landscapes/first` when multiple landscapes share one repo).
- `repositories.landscape.baseLink` is the path inside the **landscape** repository where the base repository's content is mounted. In a [submodule setup](#separate-repositories-submodule--recommended) this is the submodule mount point; in a [monorepo setup](#monorepo) it is the in-tree directory holding the base content. It is the explicit cross-repo glue that tells GLK how to reach base content from within the landscape repo.

In the example above, the landscape repository is checked out at `/path/to/landscape-repo`, the landscape configuration sits at the repository root (`target: ./`), and the base repository's content is reachable from within the landscape repo at `./base/config` (`baseLink: ./base` + `base.target: ./config`), either as a Git submodule or as an in-tree subdirectory, depending on the chosen [organization](#organization).
