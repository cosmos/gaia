ARG IMG_TAG=latest

# Compile the gaiad binary
FROM golang:1.21-alpine3.18 AS gaiad-builder
WORKDIR /src/app/
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3
RUN apk add --no-cache $PACKAGES
COPY go.mod go.sum* ./

# See https://github.com/CosmWasm/wasmvm/releases
ARG WASMVM_VERSION=v1.5.2
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep e78b224c15964817a3b75a40e59882b4d0e06fd055b39514d61646689cef8c6e
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep e660a38efb2930b34ee6f6b0bb12730adccb040b6ab701b8f82f34453a426ae7
RUN cp "/lib/libwasmvm_muslc.$(uname -m).a" /lib/libwasmvm_muslc.a

RUN go mod download
COPY . .
RUN CGO_ENABLED=0 LINK_STATICALLY=true BUILD_TAGS=muslc make install

# Add to a distroless container
FROM cgr.dev/chainguard/static:$IMG_TAG
ARG IMG_TAG
COPY --from=gaiad-builder /go/bin/gaiad /usr/local/bin/
EXPOSE 26656 26657 1317 9090
USER 0

ENTRYPOINT ["gaiad", "start"]