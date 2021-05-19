FROM golang:1.16-alpine AS build-env

# Set up dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3 wget

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/gaia

# Add source files
COPY . .

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make install

WORKDIR /

RUN wget https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz && \
        gzip -d genesis.cosmoshub-4.json.gz


# Final image
FROM alpine:edge                                                                                                                                                                              
                                                                                                                                                                                              
# Install ca-certificates                                                                                                                                                                     
RUN apk add --update ca-certificates                                                                                                                                                          
                                                                                                                                                                                              
                                                                                                                                                                                              
# Copy over binaries from the build-env                                                                                                                                                       
COPY --from=build-env /go/bin/gaiad /usr/bin/gaiad                                                                                                                                            
                                                                                                                                                                                              
                                                                                                                                                                                              
COPY --from=build-env /genesis.cosmoshub-4.json /root/.gaia/config/genesis.json.tmp                                                                                                           
                                                                                                                                                                                              
                                                                                                                                                                                              
# Run gaiad by default, omit entrypoint to ease using container with gaiacli                                                                                                                  
CMD gaiad init instagaia$RANDOM && cp /root/.gaia/config/genesis.json.tmp /root/.gaia/config/genesis.json && gaiad start --p2p.seeds bf8328b66dceb4987e5cd94430af66045e59899f@public-seed.cosmos.vitwit.com:26656,cfd785a4224c7940e9a10f6c1ab24c343e923bec@164.68.107.188:26656,d72b3011ed46d783e369fdf8ae2055b99a1e5074@173.249.50.25:26656,ba3bacc714817218562f743178228f23678b2873@public-seed-node.cosmoshub.certus.one:26656,3c7cad4154967a294b3ba1cc752e40e8779640ad@84.201.128.115:26656,366ac852255c3ac8de17e11ae9ec814b8c68bddb@51.15.94.196:26656 --x-crisis-skip-assert-invariants 
