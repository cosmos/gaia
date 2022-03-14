#!/bin/bash
# microtick and bitcanna contributed significantly here.
set -uxe

# Set Golang environment variables.
export GOPATH=~/go
export PATH=$PATH:~/go/bin

# Install Gaia.
make install

# Initialize chain.
gaiad init test 

# Prepare genesis file for cosmoshub-4.
wget https://github.com/cosmos/mainnet/raw/master/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json $HOME/.gaia/config/genesis.json

# Set minimum gas price.
sed -i'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0.0025uatom"/' $HOME/.gaia/config/app.toml

# Get "trust_hash" and "trust_height".
INTERVAL=1000
LATEST_HEIGHT=$(curl -s https://rpc.cosmos.network/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL)) 
TRUST_HASH=$(curl -s "https://rpc.cosmos.network/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export GAIAD_STATESYNC_ENABLE=true
export GAIAD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export GAIAD_STATESYNC_RPC_SERVERS="https://rpc.cosmos.network:443,https://rpc.cosmos.network:443"
export GAIAD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export GAIAD_STATESYNC_TRUST_HASH=$TRUST_HASH

# Fetch and set list of seeds from chain registry.
export GAIAD_P2P_SEEDS=$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/cosmoshub/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')

# Start chain.
gaiad start --x-crisis-skip-assert-invariants
