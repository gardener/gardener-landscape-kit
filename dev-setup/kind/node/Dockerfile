FROM kindest/node:v1.37.0@sha256:a1ed56cfb0e7b93589bdf97c8cd566405a265939e3620fc4f5de89adff580ae5

# copy containerd hosts configurations for local registry mirrors
COPY containerd /etc/containerd/certs.d/
# names with colon are not allowed in the directory name checked into Git as it breaks go mod if gardener/gardener is used as a module
RUN mv /etc/containerd/certs.d/glk-registry.local.gardener.cloud_6001 /etc/containerd/certs.d/glk-registry.local.gardener.cloud:6001

ENTRYPOINT ["/usr/local/bin/entrypoint", "/sbin/init"]
