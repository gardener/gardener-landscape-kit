apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: gardener-operator
  namespace: garden
spec:
  chartRef:
    kind: OCIRepository
    name: gardener-operator
    namespace: garden
  install:
    crds: CreateReplace
  interval: 10m0s
  upgrade:
    crds: CreateReplace
  values:
    replicaCount: 2
