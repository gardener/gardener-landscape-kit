## Start server

```bash
make git-server-up
```

## Access Web UI

http://git.local.gardener.cloud:6080/gitops

```txt
> User: `gitops`  
> Password: `testtest`
```

> [!NOTE]
> Ensure that the host `git.local.gardener.cloud` resolves to `127.0.0.1`

## Clone

- Base repo
```bash
git clone http://gitops:testtest@git.local.gardener.cloud:6080/gitops/base.git
```

- Test landscape repo
```bash
git clone http://gitops:testtest@git.local.gardener.cloud:6080/gitops/test-landscape.git
```

## Configure Git Remote in Landscape Repo

`git-sync-secret.yaml`:
```yaml
stringData:
  password: testtest
  username: gitops
```

`gotk-sync.yaml`:
```yaml
  url: http://git.local.gardener.cloud:6080/gitops/test-landscape
```

## Gardener Local Configurations

`helm-release.yaml`:
```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: gardener-operator
  namespace: garden
spec:
  values:
    env:
      - name: GARDENER_OPERATOR_LOCAL
        value: "true"
    hostAliases:
      - hostnames:
          - api.virtual-garden.local.gardener.cloud
        ip: 10.2.10.2
        # config:
        # featureGates:
        # IstioTLSTermination: true
        # VPAInPlaceUpdates: true
```

`kustomization.yaml`:
```yaml
patches:
  - path: helm-release.yaml
```
