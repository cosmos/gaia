###################################################################################################
# Build image
###################################################################################################
FROM golang:alpine AS build-env

# Add dependencies
RUN apk add --no-cache curl make git libc-dev bash gcc linux-headers eudev-dev python3

# Gaia Branch or Release
ARG RELEASE

# Clone repository
RUN git clone https://github.com/cosmos/gaia /go/src/github.com/cosmos/gaia

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/gaia

# Checkout branch
RUN git checkout $RELEASE

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make tools && \
    make install

# Show version
RUN gaiad version --long

###################################################################################################
# Final image
###################################################################################################
FROM alpine:edge
LABEL maintainer="hello@informal.systems"

ARG RELEASE
ARG CHAIN
ARG NAME

# Add jq for debugging
RUN apk add --no-cache jq curl tree

WORKDIR /$NAME

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/gaiad /usr/bin/gaiad

COPY --chown=root:root ./chains/$CHAIN/$RELEASE/$NAME /chain/$CHAIN

# Copy entrypoint script
COPY ./run-gaiad.sh /chain/$CHAIN
RUN chmod 755 /chain/$CHAIN/run-gaiad.sh

RUN tree -pug /chain

ENTRYPOINT "/bin/sh"
