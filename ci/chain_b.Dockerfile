FROM alpine
LABEL maintainer="hello@informal.systems"

EXPOSE 26656 26657 26660

ENTRYPOINT ["/usr/bin/gaiad"]

CMD ["start"]

VOLUME [ "/root" ]

COPY gaia/build/gaiad /usr/bin/gaiad
COPY gaia/build/chain_b/node0/gaiad /root/.gaia
COPY gaia/build/chain_b/node0/gaiad/key_seed.json /root/key_seed.json
