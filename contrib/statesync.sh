#!/bin/bash
# microtick and bitcanna contributed significantly here.


INTERVAL=1000

# GET TRUST HASH AND TRUST HEIGHT

LATEST_HEIGHT=$(curl -s 162.55.132.230:2011/block | jq -r .result.block.header.height);
BLOCK_HEIGHT=$(($LATEST_HEIGHT-$INTERVAL)) 
TRUST_HASH=$(curl -s "162.55.132.230:2011/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)


# TELL USER WHAT WE ARE DOING
echo "TRUST HEIGHT: $BLOCK_HEIGHT"
echo "TRUST HASH: $TRUST_HASH"


# export state sync vars
export GAIAD_STATESYNC_ENABLE=true
export GAIAD_STATESYNC_RPC_SERVERS="162.55.132.230:2011,162.55.132.230:2011"
export GAIAD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export GAIAD_STATESYNC_TRUST_HASH=$TRUST_HASH
export GAIAD_P2P_PERSISTENT_PEERS="7f317a82192462dcf0dca48111afcce895cf68d0@162.55.132.230:2010"

gaiad unsafe-reset-all
gaiad start
