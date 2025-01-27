FROM golang:1.21-alpine AS go-builder

SHELL ["/bin/sh", "-ecuxo", "pipefail"]

RUN apk add --no-cache ca-certificates build-base git

WORKDIR /code

ADD go.mod go.sum ./
RUN set -eux; \
    export ARCH=$(uname -m); \
    WASM_VERSION=$(go list -m all | grep github.com/CosmWasm/wasmvm/v2 || true); \
    if [ ! -z "${WASM_VERSION}" ]; then \
      WASMVM_REPO=$(echo $WASM_VERSION | awk '{print $1}');\
      WASMVM_VERS=$(echo $WASM_VERSION | awk '{print $2}');\
      if [ $(echo $WASMVM_REPO | grep -c '/v2$') -gt 0 ]; then \
        WASMVM_REPO=$(echo $WASMVM_REPO | sed 's/\/v2$//');\
      fi; \
      wget -O /lib/libwasmvm_muslc.a https://${WASMVM_REPO}/releases/download/${WASMVM_VERS}/libwasmvm_muslc.$(uname -m).a;\
      # https://github.com/strangelove-ventures/heighliner/pull/263
      wget -O /lib/libwasmvm.so https://${WASMVM_REPO}/releases/download/${WASMVM_VERS}/libwasmvm.$(uname -m).so;\
      wget -O /lib/libwasmvm_muslc.$(uname -m).a https://${WASMVM_REPO}/releases/download/${WASMVM_VERS}/libwasmvm_muslc.$(uname -m).a;\
      wget -O /lib/libwasmvm.$(uname -m).so https://${WASMVM_REPO}/releases/download/${WASMVM_VERS}/libwasmvm.$(uname -m).so;\
    fi; \
    go mod download;

# Copy over code
COPY . /code

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
# then log output of file /code/bin/tokend
# then ensure static linking
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc LINK_STATICALLY=true make build \
  && file /code/build/tokend \
  && echo "Ensuring binary is statically linked ..." \
  && (file /code/build/tokend | grep "statically linked")

# --------------------------------------------------------
FROM alpine:3.16

COPY --from=go-builder /code/build/tokend /usr/bin/tokend

# Install dependencies used for Starship
RUN apk add --no-cache curl make bash jq sed

WORKDIR /opt

# rest server, tendermint p2p, tendermint rpc
EXPOSE 1317 26656 26657

CMD ["/usr/bin/tokend", "version"]
