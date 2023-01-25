#!/bin/bash

NETWORK=bigbang-1
DAEMON=akashd
HOME_DIR=~/.akashd
CONFIG=~/.akashd/config
TOKEN_DENOM=star
FAUCET_ACCOUNTS=("akash1czxh6ewhuy00tsv5zu50gz7lz2cxcpufdrarty" "akash1qjcvelu4rud75jztawcls48luxmapcajvfdhuy")

rm -rf $HOME_DIR

$DAEMON init $NETWORK --chain-id $NETWORK

rm -rf $CONFIG/gentx && mkdir $CONFIG/gentx

sed -i "s/\"stake\"/\"$TOKEN_DENOM\"/g" $HOME_DIR/config/genesis.json

for i in $NETWORK/gentx/*.json; do
  echo $i
  $DAEMON add-genesis-account $(jq -r '.value.msg[0].value.delegator_address' $i) 1000000000$TOKEN_DENOM
  cp $i $CONFIG/gentx/
done

for addr in "${FAUCET_ACCOUNTS[@]}"; do
    echo "Adding faucet addr: $addr"
    $DAEMON add-genesis-account $addr 100000000000$TOKEN_DENOM
done

$DAEMON collect-gentxs

$DAEMON validate-genesis

cp $CONFIG/genesis.json $NETWORK