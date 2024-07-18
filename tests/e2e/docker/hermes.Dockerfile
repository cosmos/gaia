FROM --platform=linux/amd64 informalsystems/hermes:1.10.0 AS hermes-builder

FROM --platform=linux/amd64 debian:buster-slim
USER root

COPY --from=hermes-builder /usr/bin/hermes /usr/local/bin/
RUN chmod +x /usr/local/bin/hermes

EXPOSE 3031