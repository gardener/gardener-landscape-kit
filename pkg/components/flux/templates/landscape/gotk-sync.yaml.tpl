---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: flux-system
  namespace: flux-system
spec:
  interval: 1m0s
  ref:
    {{ .repo_ref }}
  secretRef:
    name: flux-system
  url: {{ .repo_url }}
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: flux-system
  namespace: flux-system
spec:
  interval: 10m0s
  path: ./{{ .flux_path }}
  prune: false
  sourceRef:
    kind: GitRepository
    name: flux-system
