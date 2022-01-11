FROM alpine
LABEL maintainer="hello@informal.systems"

EXPOSE 26656 26657 26660 1317

ENTRYPOINT ["/usr/bin/simd"]

CMD ["--home", "/root/.simapp", "start"]

VOLUME [ "/root" ]

#Commit ID: c2d40e1099d2c00c02f68bc156c57603640e3590
COPY cosmos-sdk/build/simd /usr/bin/simd
COPY simapp/ /root/.simapp
