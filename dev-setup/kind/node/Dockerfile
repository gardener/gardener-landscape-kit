FROM kindest/node:v1.34.11@sha256:44e222ee2132dab25ff87301682f89eb82c7880ea3a1bf543bfe9708fd08d67d

# copy containerd hosts configurations for local registry mirrors
COPY containerd /etc/containerd/certs.d/
# names with colon are not allowed in the directory name checked into Git as it breaks go mod if gardener/gardener is used as a module
RUN mv /etc/containerd/certs.d/glk-registry.local.gardener.cloud_6001 /etc/containerd/certs.d/glk-registry.local.gardener.cloud:6001

ENTRYPOINT ["/usr/local/bin/entrypoint", "/sbin/init"]
