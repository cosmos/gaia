#!/bin/bash
# microtick and bitcanna contributed significantly here.
# Pebbledb state sync script.
# invoke like: bash contrib/statesync.bash

## USAGE RUNDOWN
# Not for use on live nodes
# For use when testing.

## THIS IS FOR THE REPLICATED SECURITY TESTNET

set -uxe

# Set Golang environment variables.
# ! Adapt as required, depending on your system configuration
export GOPATH=~/go
export PATH=$PATH:~/go/bin


# Initialize chain.
gaiad init test --home ~/.gaia-rs

# Get Genesis for testnet
wget -O ~/.gaia-rs/config/genesis.json https://github.com/cosmos/testnets/raw/master/replicated-security/provider/provider-genesis.json


# Get "trust_hash" and "trust_height" for testnet
INTERVAL=100
LATEST_HEIGHT=$(curl -s https://rpc.provider-state-sync-01.rs-testnet.polypore.xyz:443/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$((LATEST_HEIGHT - INTERVAL))
TRUST_HASH=$(curl -s "https://rpc.provider-state-sync-01.rs-testnet.polypore.xyz:443/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)



# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export GAIAD_STATESYNC_ENABLE=true
export GAIAD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export GAIAD_STATESYNC_RPC_SERVERS="https://rpc.provider-state-sync-01.rs-testnet.polypore.xyz:443,https://rpc.provider-state-sync-01.rs-testnet.polypore.xyz:443"
export GAIAD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export GAIAD_STATESYNC_TRUST_HASH=$TRUST_HASH

# Fetch and set list of seeds from chain registry.
export GAIAD_P2P_SEEDS="08ec17e86dac67b9da70deb20177655495a55407@provider-seed-01.rs-testnet.polypore.xyz:26656,4ea6e56300a2f37b90e58de5ee27d1c9065cf871@provider-seed-02.rs-testnet.polypore.xyz:26656"

# Start chain.
gaiad start --x-crisis-skip-assert-invariants --iavl-disable-fastnode false --home ~/.gaia-rs
