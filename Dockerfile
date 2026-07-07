# builder
FROM --platform=$BUILDPLATFORM golang:1.26.5 AS builder

ARG EFFECTIVE_VERSION
ARG BUILD_DATE
ARG TARGETOS
ARG TARGETARCH
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=$GOPROXY
ENV BUILD_DATE=$BUILD_DATE
WORKDIR /go/src/github.com/gardener/gardener-landscape-kit

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build EFFECTIVE_VERSION=$EFFECTIVE_VERSION GOOS=$TARGETOS GOARCH=$TARGETARCH BUILD_OUTPUT_FILE="/output/bin/"

# distroless-static
FROM gcr.io/distroless/static-debian13:nonroot AS distroless-static

# gardener-landscape-kit
FROM distroless-static AS gardener-landscape-kit
COPY --from=builder /output/bin/gardener-landscape-kit /gardener-landscape-kit
WORKDIR /
ENTRYPOINT ["/gardener-landscape-kit"]
