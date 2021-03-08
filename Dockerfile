# Simple usage with a mounted data directory:
# > docker build -t gaia .
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.gaia:/gaia/.gaia gaia gaiad init
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.gaia:/gaia/.gaia gaia gaiad start
FROM golang:1.15-alpine AS build-env

# Set up dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/gaia

# Add source files
COPY . .

RUN go version

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make install

# Final image
FROM alpine:edge

ENV GAIA /gaia

# Install ca-certificates
RUN apk add --update ca-certificates

RUN addgroup gaia && \
    adduser -S -G gaia gaia -h "$GAIA"

USER gaia

WORKDIR $GAIA

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/gaiad /usr/bin/gaiad

# Run gaiad by default, omit entrypoint to ease using container with gaiacli
CMD ["gaiad"]
