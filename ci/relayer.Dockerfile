#####################################################
####                 Relayer image               ####
#####################################################
FROM ubuntu:21.04
LABEL maintainer="hello@informal.systems"

ARG RELEASE

# Add Python 3 and Ping and Telnet (for testing)
RUN apt-get update -y && apt-get install python3 python3-toml iputils-ping telnet -y

# Copy relayer executable
COPY ./hermes /usr/bin/hermes

# Relayer folder
WORKDIR /relayer

# Copy configuration file
COPY ci/simple_config.toml .

# Copy setup script
COPY ci/e2e.sh .

# Copy end-to-end testing script
COPY e2e ./e2e

# Copy key files
COPY ci/chains/gaia/$RELEASE/ibc-0/user_seed.json  ./user_seed_ibc-0.json
RUN cat ./user_seed_ibc-0.json
COPY ci/chains/gaia/$RELEASE/ibc-1/user_seed.json  ./user_seed_ibc-1.json
RUN cat ./user_seed_ibc-1.json
COPY ci/chains/gaia/$RELEASE/ibc-0/user2_seed.json ./user2_seed_ibc-0.json
RUN cat ./user2_seed_ibc-0.json
COPY ci/chains/gaia/$RELEASE/ibc-1/user2_seed.json ./user2_seed_ibc-1.json
RUN cat ./user2_seed_ibc-1.json

# Make it executable
RUN chmod +x e2e.sh

ENTRYPOINT ["/bin/sh"]
