FROM golang:1.17-alpine3.15 as builder

# Set up dependencies
ENV PACKAGES bash curl make git libc-dev gcc linux-headers eudev-dev python3

WORKDIR /gaia

# Add source files
COPY . .

# Install minimum necessary dependencies
RUN apk add --no-cache $PACKAGES && make build

# ----------------------------
FROM alpine:edge

RUN apk add --update ca-certificates

# p2p port
EXPOSE 26656
# rpc port
EXPOSE 26657
# metrics port
EXPOSE 26660
# rest server
EXPOSE 1317

COPY --from=builder /gaia/build/gaiad /usr/local/bin/gaiad