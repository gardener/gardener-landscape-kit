FROM kindest/node:v1.34.8@sha256:02722c2dedddcfc00febf5d27fbeb9b7b2c14294c82109ff4a85d89ac9ba3256

# copy containerd hosts configurations for local registry mirrors
COPY containerd /etc/containerd/certs.d/
# names with colon are not allowed in the directory name checked into Git as it breaks go mod if gardener/gardener is used as a module
RUN mv /etc/containerd/certs.d/glk-registry.local.gardener.cloud_5001 /etc/containerd/certs.d/glk-registry.local.gardener.cloud:5001

ENTRYPOINT ["/usr/local/bin/entrypoint", "/sbin/init"]
