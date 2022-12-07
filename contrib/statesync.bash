#!/bin/bash
# microtick and bitcanna contributed significantly here.
set -uxe

# set environment variables
export GOPATH=~/go
export PATH=$PATH:~/go/bin


# Install Gaia
make install

# MAKE HOME FOLDER AND GET GENESIS
gaiad init test 
wget https://github.com/cosmos/mainnet/raw/master/genesis/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json ~/.gaia/config/genesis.json
rm genesis.cosmoshub-4.json.gz

# IPFS hosted alternative download link
# wget -O ~/.gaia/config/genesis.json https://cloudflare-ipfs.com/ipfs/Qmc54DreioPpPDUdJW6bBTYUKepmcPsscfqsfFcFmTaVig

INTERVAL=1000

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s https://rpc-cosmoshub-ia.cosmosia.notional.ventures/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL)) 
TRUST_HASH=$(curl -s "https://rpc-cosmoshub-ia.cosmosia.notional.ventures/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export GAIAD_STATESYNC_ENABLE=true
export GAIAD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export GAIAD_STATESYNC_RPC_SERVERS="https://cosmos-rpc.polkachu.com:443,https://rpc-cosmoshub-ia.cosmosia.notional.ventures:443,https://rpc.cosmos.network:443"
export GAIAD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export GAIAD_STATESYNC_TRUST_HASH=$TRUST_HASH
# Fetch and set list of seeds from chain registry.
GAIAD_P2P_SEEDS=$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/cosmoshub/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')
export GAIAD_P2P_SEEDS

gaiad start --x-crisis-skip-assert-invariants
