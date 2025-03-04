# TODO: dynamically fetch this from cargo file
ARG SP1_VERSION="v4.1.0"
FROM ghcr.io/succinctlabs/sp1:${SP1_VERSION} as BUILD

RUN apt update && apt install -y just protobuf-compiler

WORKDIR /solidity-ibc-eureka/
COPY . .
RUN just build-relayer

FROM gcr.io/distroless/base-debian11:debug
WORKDIR /usr/local/bin
COPY --from=BUILD /solidity-ibc-eureka/target/release/relayer /usr/local/bin/relayer
ENTRYPOINT [ "/usr/local/bin/relayer", "--config", "/usr/local/relayer/relayer.json"]
