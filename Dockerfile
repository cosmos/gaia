# Info on how to use this docker image can be found in DOCKER_README.md
ARG IMG_TAG=latest

# Compile the gaiad binary
FROM golang:1.26.10-alpine AS gaiad-builder
WORKDIR /src/app/
ENV PACKAGES="curl build-base git bash file linux-headers eudev-dev"
RUN apk add --no-cache $PACKAGES

ARG CGO_CFLAGS="-D__BLST_PORTABLE__"
ENV CGO_CFLAGS=$CGO_CFLAGS

ARG GIT_VERSION=""
ARG GIT_COMMIT=""

# See https://github.com/CosmWasm/wasmvm/releases
ARG WASMVM_VERSION=v2.3.6
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep dd4beaf9e0092b8f6d398d09015edd400fef656cb0879b09e7441d02451ca03d
RUN cp "/lib/libwasmvm_muslc.x86_64.a" /lib/libwasmvm_muslc.a

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN LEDGER_ENABLED=false LINK_STATICALLY=true BUILD_TAGS=muslc \
    VERSION=$GIT_VERSION COMMIT=$GIT_COMMIT make build
RUN echo "Ensuring binary is statically linked (static-pie) ..."  \
    && file /src/app/build/gaiad | grep "static-pie linked"

FROM alpine:$IMG_TAG
RUN apk add --no-cache build-base jq
RUN addgroup -g 1025 nonroot
RUN adduser -D nonroot -u 1025 -G nonroot
ARG IMG_TAG
COPY --from=gaiad-builder  /src/app/build/gaiad /usr/local/bin/
EXPOSE 26656 26657 1317 9090
USER nonroot

ENTRYPOINT ["gaiad", "start"]
