FROM kindest/node:v1.34.0@sha256:7416a61b42b1662ca6ca89f02028ac133a309a2a30ba309614e8ec94d976dc5a

# copy containerd hosts configurations for local registry mirrors
COPY containerd /etc/containerd/certs.d/
# names with colon are not allowed in the directory name checked into Git as it breaks go mod if gardener/gardener is used as a module
RUN mv /etc/containerd/certs.d/glk-registry.local.gardener.cloud_5001 /etc/containerd/certs.d/glk-registry.local.gardener.cloud:5001

ENTRYPOINT ["/usr/local/bin/entrypoint", "/sbin/init"]
