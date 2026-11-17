# Rebuilds an upstream Flux controller on a distroless/minimal base by copying its (statically linked) binary out
# of the upstream image — no recompilation.
#
#   static    source-controller, helm-controller, notification-controller  (distroless/static)
#   kustomize kustomize-controller                                         (alpine + git/gpg/ssh/tini)
#
# Build: --target <static|kustomize> --build-arg FLUX_IMAGE=<upstream-ref> --build-arg CONTROLLER=<name>

ARG FLUX_IMAGE
ARG CONTROLLER
FROM ${FLUX_IMAGE} AS upstream

# Digest-pinned per gardener/gardener#16079; Renovate keeps the @sha256 current.
# renovate: datasource=docker
FROM gcr.io/distroless/static-debian13:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931 AS distroless-static

# static: the three controllers with no runtime deps. Binary goes to a fixed /controller path because a JSON-array
# ENTRYPOINT cannot expand the ${CONTROLLER} ARG and distroless has no shell (same pattern as the root Dockerfile).
FROM distroless-static AS static
ARG CONTROLLER
COPY --from=upstream /usr/local/bin/${CONTROLLER} /controller
USER 65534:65534
WORKDIR /
ENTRYPOINT ["/controller"]

# kustomize: keep the upstream Alpine runtime tools and GPG configuration, but refresh packages.
# This does not fix vulnerabilities embedded in the copied Go binary.
# renovate: datasource=docker
FROM alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf AS kustomize
RUN apk --no-cache upgrade \
    && apk --no-cache add ca-certificates git openssh-client gnupg tini \
    && update-ca-certificates
COPY --from=upstream /usr/local/bin/kustomize-controller /usr/local/bin/kustomize-controller
ENV GNUPGHOME=/tmp
USER 65534:65534
WORKDIR /
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/kustomize-controller"]
