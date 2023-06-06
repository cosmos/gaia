FROM informalsystems/hermes:1.4.1 AS hermes-builder

FROM debian:buster
USER root

COPY --from=hermes-builder /usr/bin/hermes /usr/local/bin/
RUN chmod +x /usr/local/bin/hermes

EXPOSE 3031
# ENTRYPOINT ["hermes", "start"]