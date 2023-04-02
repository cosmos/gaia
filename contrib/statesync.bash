#!/bin/bash
# microtick and bitcanna contributed significantly here.
# Pebbledb state sync script.
# invoke like: bash contrib/statesync.bash

## USAGE RUNDOWN
# Not for use on live nodes
# For use when testing.

set -uxe

# Set Golang environment variables.
# ! Adapt as required, depending on your system configuration
export GOPATH=~/go
export PATH=$PATH:~/go/bin

# Install with pebbledb (uncomment for incredible performance)
# go mod edit -replace github.com/tendermint/tm-db=github.com/baabeetaa/tm-db@pebble
# go mod tidy

# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb -X github.com/tendermint/tm-db.ForceSync=1' -tags pebbledb ./...

# install (comment if using pebble for incredible performance)
# go install ./...

# NOTE: ABOVE YOU CAN USE ALTERNATIVE DATABASES, HERE ARE THE EXACT COMMANDS
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb' -tags badgerdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb' -tags boltdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb -X github.com/tendermint/tm-db.ForceSync=1' -tags pebbledb ./...


# Initialize chain.
gaiad init test

# Get Genesis
wget https://github.com/cosmos/mainnet/raw/master/genesis/genesis.cosmoshub-4.json.gz
gzip -d genesis.cosmoshub-4.json.gz
mv genesis.cosmoshub-4.json ~/.gaia/config/genesis.json

# Get "trust_hash" and "trust_height".
INTERVAL=100
LATEST_HEIGHT=$(curl -s https://cosmos-rpc.polkachu.com/block | jq -r .result.block.header.height)
BLOCK_HEIGHT=$((LATEST_HEIGHT - INTERVAL))
TRUST_HASH=$(curl -s "https://cosmos-rpc.polkachu.com/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)

# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export GAIAD_STATESYNC_ENABLE=true
export GAIAD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export GAIAD_STATESYNC_RPC_SERVERS="https://cosmos-rpc.polkachu.com:443,https://rpc-cosmoshub-ia.cosmosia.notional.ventures:443"
export GAIAD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export GAIAD_STATESYNC_TRUST_HASH=$TRUST_HASH

# Fetch and set list of seeds from chain registry.
GAIAD_P2P_SEEDS=$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/cosmoshub/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')
export GAIAD_P2P_SEEDS

# Start chain.
gaiad start --x-crisis-skip-assert-invariants --iavl-disable-fastnode false
