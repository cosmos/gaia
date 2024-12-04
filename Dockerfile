# Info on how to use this docker image can be found in DOCKER_README.md
ARG IMG_TAG=latest

# Compile the gaiad binary
FROM golang:1.22-alpine AS gaiad-builder
WORKDIR /src/app/
ENV PACKAGES="curl make git libc-dev bash file gcc linux-headers eudev-dev"

# Combine installation of packages and checksum verification to reduce layers
RUN apk add --no-cache $PACKAGES \
    && curl -LO https://github.com/CosmWasm/wasmvm/releases/download/v2.1.3/libwasmvm_muslc.aarch64.a \
    && curl -LO https://github.com/CosmWasm/wasmvm/releases/download/v2.1.3/libwasmvm_muslc.x86_64.a \
    && sha256sum libwasmvm_muslc.aarch64.a | grep faea4e15390e046d2ca8441c21a88dba56f9a0363f92c5d94015df0ac6da1f2d \
    && sha256sum libwasmvm_muslc.x86_64.a | grep 8dab08434a5fe57a6fbbcb8041794bc3c31846d31f8ff5fb353ee74e0fcd3093 \
    && cp "/lib/libwasmvm_muslc.$(uname -m).a" /lib/libwasmvm_muslc.a

# First copy only the go.mod and go.sum to leverage caching
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Now copy the source code
COPY . .
RUN LEDGER_ENABLED=false LINK_STATICALLY=true BUILD_TAGS=muslc make build

RUN echo "Ensuring binary is statically linked ..."  \
    && file /src/app/build/gaiad | grep "statically linked"

# Final stage: minimal image with only the required binary
FROM alpine:$IMG_TAG
RUN apk add --no-cache build-base jq
RUN addgroup -g 1025 nonroot
RUN adduser -D nonroot -u 1025 -G nonroot

COPY --from=gaiad-builder /src/app/build/gaiad /usr/local/bin/

EXPOSE 26656 26657 1317 9090
USER nonroot

ENTRYPOINT ["gaiad", "start"]
