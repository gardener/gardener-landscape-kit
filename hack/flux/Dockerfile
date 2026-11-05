# Rebuilds an upstream Flux controller on a distroless/minimal base by copying its (statically linked) binary out
# of the upstream image — no recompilation.
#
#   static    source-controller, helm-controller, notification-controller  (distroless/static)
#   kustomize kustomize-controller                                         (debian:13-slim + git/gpg/ssh/tini)
#
# Build: --target <static|kustomize> --build-arg FLUX_IMAGE=<upstream-ref> --build-arg CONTROLLER=<name>

ARG FLUX_IMAGE
ARG CONTROLLER
FROM ${FLUX_IMAGE} AS upstream

# Digest-pinned per gardener/gardener#16079; Renovate keeps the @sha256 current.
# renovate: datasource=docker
FROM gcr.io/distroless/static-debian13:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99 AS distroless-static

# static: the three controllers with no runtime deps. Binary goes to a fixed /controller path because a JSON-array
# ENTRYPOINT cannot expand the ${CONTROLLER} ARG and distroless has no shell (same pattern as the root Dockerfile).
FROM distroless-static AS static
ARG CONTROLLER
COPY --from=upstream /usr/local/bin/${CONTROLLER} /controller
USER 65534:65534
WORKDIR /
ENTRYPOINT ["/controller"]

# kustomize: kustomize-controller shells out to git/gpg/ssh and uses tini as PID 1, so it needs those at runtime.
# We install them from Debian packages rather than copying the upstream Alpine (musl) binaries, which are
# incompatible with a glibc base. debian:13-slim is used over distroless/base because the git/gpg/ssh shared-library
# tree is large and error-prone to hand-copy into distroless.
# Digest-pinned; Renovate keeps the @sha256 current.
# renovate: datasource=docker
FROM debian:13-slim@sha256:f6b08b09a05725f9431502d3b5714f5c4576e82c1a84fa1debe0c24a7e0c3ec4 AS kustomize
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        git \
        gnupg \
        openssh-client \
        tini \
    && rm -rf /var/lib/apt/lists/*
COPY --from=upstream /usr/local/bin/kustomize-controller /usr/local/bin/kustomize-controller
USER 65534:65534
WORKDIR /
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/kustomize-controller"]
