apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: gardener-operator
  namespace: garden
spec:
  interval: 0s
  path: {{ .landscapeComponentPath }}
  prune: false
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
