# Info on how to use this docker image can be found in DOCKER_README.md
ARG IMG_TAG=latest

# Compile the gaiad binary
FROM golang:1.24-alpine AS gaiad-builder
WORKDIR /src/app/
ENV PACKAGES="curl build-base git bash file linux-headers eudev-dev"
RUN apk add --no-cache $PACKAGES

ARG CGO_CFLAGS="-D__BLST_PORTABLE__"
ENV CGO_CFLAGS=$CGO_CFLAGS

# See https://github.com/CosmWasm/wasmvm/releases
ARG WASMVM_VERSION=v2.3.2
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 4b87af3c8aac1756ee1aa1e06daefe3a7f5a3469a3c8d77ad07513539606f8a6
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep 4d03a4bf508c89a303e8d7d0236feac44a40f6b6e221df4076968abe9d1e49c6
RUN cp "/lib/libwasmvm_muslc.$(uname -m).a" /lib/libwasmvm_muslc.a

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN LEDGER_ENABLED=false LINK_STATICALLY=true BUILD_TAGS=muslc make build
RUN echo "Ensuring binary is statically linked ..."  \
    && file /src/app/build/gaiad | grep "statically linked"

FROM alpine:$IMG_TAG
RUN apk add --no-cache build-base jq
RUN addgroup -g 1025 nonroot
RUN adduser -D nonroot -u 1025 -G nonroot
ARG IMG_TAG
COPY --from=gaiad-builder  /src/app/build/gaiad /usr/local/bin/
EXPOSE 26656 26657 1317 9090
USER nonroot

ENTRYPOINT ["gaiad", "start"]
