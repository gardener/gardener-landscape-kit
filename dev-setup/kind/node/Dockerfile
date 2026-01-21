FROM kindest/node:v1.34.3@sha256:08497ee19eace7b4b5348db5c6a1591d7752b164530a36f855cb0f2bdcbadd48

# copy containerd hosts configurations for local registry mirrors
COPY containerd /etc/containerd/certs.d/
# names with colon are not allowed in the directory name checked into Git as it breaks go mod if gardener/gardener is used as a module
RUN mv /etc/containerd/certs.d/glk-registry.local.gardener.cloud_5001 /etc/containerd/certs.d/glk-registry.local.gardener.cloud:5001

ENTRYPOINT ["/usr/local/bin/entrypoint", "/sbin/init"]
