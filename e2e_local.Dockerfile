ARG IMG_TAG=latest

# Add to a distroless container
FROM gcr.io/distroless/cc:$IMG_TAG
ARG IMG_TAG
COPY ./build/gaiad /usr/local/bin/
EXPOSE 26656 26657 1317 9090

ENTRYPOINT ["gaiad", "start"]
