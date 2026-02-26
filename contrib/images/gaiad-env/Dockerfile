# Info on how to use this docker image can be found in DOCKER_README.md
ARG IMG_TAG=latest

# Compile the gaiad binary
FROM golang:1.24-alpine AS gaiad-builder
WORKDIR /work
ENV PACKAGES="curl build-base git bash file linux-headers eudev-dev"
RUN apk add --no-cache $PACKAGES

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
    && file /work/build/gaiad | grep "statically linked"

FROM alpine:$IMG_TAG AS run
RUN apk add --no-cache build-base jq bash curl
RUN addgroup -g 1025 nonroot
RUN adduser -D nonroot -u 1025 -G nonroot

# Set up the runtime environment
EXPOSE 26656 26657 1317 9090
STOPSIGNAL SIGTERM
VOLUME /gaiad
WORKDIR /gaiad

# Copy the wrapper script and binary to expected locations
COPY contrib/images/gaiad-env/wrapper.sh /usr/bin/wrapper.sh
COPY --from=gaiad-builder /work/build/gaiad /gaiad/
COPY --from=gaiad-builder /work/build/gaiad /usr/local/bin/

# Set proper ownership and permissions before switching to nonroot user
RUN chown nonroot:nonroot /usr/bin/wrapper.sh && chmod +x /usr/bin/wrapper.sh
RUN chown -R nonroot:nonroot /gaiad

USER nonroot

ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start", "--log_format", "plain", "--minimum-gas-prices", "0uatom"]