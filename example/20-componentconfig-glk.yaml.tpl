apiVersion: landscape.config.gardener.cloud/v1alpha1
kind: LandscapeKitConfiguration
# ocm:
#   repositories:
#   - <repo-url>
#   rootComponent:
#     name: <component-name>
#     version: <component-version>
#   originalRefs: true
#   ignoreMissingComponents: false
# repositories:
#   base:
#     target: ./
#     componentsFiles:
#     - components.yaml
#     - custom-components.yaml
#   landscape:
#     kind: GitRepository # or OCIRepository
#     url: https://github.com/<org>/<repo> # or: oci://<oci-registry-url>/<repo-name>
#     ref:
#       branch: <branch-name> # only for GitRepository
#       tag: <tag-name>
#     baseLink: ./base
#     target: ./
#     componentsFiles:
#     - components.yaml
#     - custom-components.yaml
# components:
{{- range .Components}}
# - name: {{.}}
{{- end}}
# versionConfig:
#   defaultVersionsUpdateStrategy: ReleaseBranch
#   checkMode: Strict # or Warning
# mergeMode: Hint
