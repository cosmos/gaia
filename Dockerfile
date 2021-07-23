# Simple usage with a mounted data directory:
# $ docker build -t <ImageName> .
# $ docker run --rm -it -p 46657:46657 -p 46656:46656 -v ~/.althea:/althea/.althea <ImageName> althea init <Moniker>
# $ docker run --rm -it -p 46657:46657 -p 46656:46656 -v ~/.althea:/althea/.althea <ImageName> althea start
FROM golang:1.15-alpine AS build-env

# Set up dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3

# Set working directory for the build
WORKDIR /src

# Add source files
COPY . .

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make install

# Final image
FROM alpine:edge

ENV ALTHEA /althea

# Install ca-certificates
RUN apk add --update ca-certificates && \
    addgroup althea && \
    adduser -S -G althea althea -h "$ALTHEA"

USER althea

WORKDIR $ALTHEA

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/althea /usr/bin/althea

# Run althea by default
CMD ["althea"]
