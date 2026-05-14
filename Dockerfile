# Info on how to use this docker image can be found in DOCKER_README.md
ARG IMG_TAG=latest

# Compile the gaiad binary
FROM golang:1.25.7-alpine AS gaiad-builder
WORKDIR /src/app/
ENV PACKAGES="curl build-base git bash file linux-headers eudev-dev"
RUN apk add --no-cache $PACKAGES

ARG CGO_CFLAGS="-D__BLST_PORTABLE__"
ENV CGO_CFLAGS=$CGO_CFLAGS

# See https://github.com/CosmWasm/wasmvm/releases
ARG WASMVM_VERSION=v2.3.3
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 3e7846f80ef707d673eada84f4a57a4cc788edc6950c490108e963d2fb3280fc
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep 097dccf8c71a28410e7bbe4cddaf506fadc1bda080cfd58fb0a4edeba8d5eb45
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
