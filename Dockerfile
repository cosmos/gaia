# Simple usage with a mounted data directory:
# > docker build -t gaia .
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.gaiad:/gaia/.gaiad -v ~/.gaiacli:/gaia/.gaiacli gaia gaiad init
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.gaiad:/gaia/.gaiad -v ~/.gaiacli:/gaia/.gaiacli gaia gaiad start
FROM golang:alpine AS build-env

# Set up dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/gaia

# Add source files
COPY . .

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make install

# Final image
FROM alpine:edge

ENV GAIA /gaia

# Install ca-certificates
RUN apk add --update ca-certificates

RUN addgroup gaiauser && \
    adduser -S -G gaiauser gaiauser -h "$GAIA"
    
USER gaiauser

WORKDIR $GAIA

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/gaiad /usr/bin/gaiad
COPY --from=build-env /go/bin/gaiacli /usr/bin/gaiacli

# Run gaiad by default, omit entrypoint to ease using container with gaiacli
CMD ["gaiad"]
