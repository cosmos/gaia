ARG IMG_TAG=latest

# Compile the gaiad binary
FROM golang:1.21-alpine AS gaiad-builder
WORKDIR /src/app/
COPY go.mod go.sum* ./
RUN set -eux; \
    export ARCH=$(uname -m); \
    WASMVM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm | awk '{print $2}'); \
    if [ ! -z "${WASMVM_VERSION}" ]; then \
      wget -O /lib/libwasmvm_muslc.a https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.${ARCH}.a; \
    fi; \
    go mod download;
COPY . .
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3
RUN apk add --no-cache $PACKAGES
RUN CGO_ENABLED=0 LINK_STATICALLY=true BUILD_TAGS=muslc make install

# Add to a distroless container
FROM alpine:$IMG_TAG
RUN adduser -D nonroot
ARG IMG_TAG
COPY --from=gaiad-builder /go/bin/gaiad /usr/local/bin/
EXPOSE 26656 26657 1317 9090
USER nonroot

ENTRYPOINT ["gaiad", "start"]
