#!/bin/bash
set -eux
# your gaiad binary name
BIN=althea

CHAIN_ID="gravity-test"

NODES=$1

ALLOCATION="10000000000stake,10000000000footoken"

# first we start a genesis.json with validator 1
# validator 1 will also collect the gentx's once gnerated
STARTING_VALIDATOR=1
STARTING_VALIDATOR_HOME="--home /validator$STARTING_VALIDATOR"
# todo add git hash to chain name
$BIN init $STARTING_VALIDATOR_HOME --chain-id=$CHAIN_ID validator1
mv /validator$STARTING_VALIDATOR/config/genesis.json /genesis.json

# copy over the legacy RPC enabled config for the first validator
# this is the only validator we want taking up this port
cp /legacy-api-enable /validator1/config/app.toml

# Sets up an arbitrary number of validators on a single machine by manipulating
# the --home parameter on gaiad
for i in $(seq 1 $NODES);
do
GAIA_HOME="--home /validator$i"
GENTX_HOME="--home-client /validator$i"
ARGS="$GAIA_HOME --keyring-backend test"
$BIN keys add $ARGS validator$i 2>> /validator-phrases
KEY=$($BIN keys show validator$i -a $ARGS)
# move the genesis in
mkdir -p /validator$i/config/
mv /genesis.json /validator$i/config/genesis.json
$BIN add-genesis-account $ARGS $KEY $ALLOCATION
# move the genesis back out
mv /validator$i/config/genesis.json /genesis.json
done


for i in $(seq 1 $NODES);
do
cp /genesis.json /validator$i/config/genesis.json
GAIA_HOME="--home /validator$i"
ARGS="$GAIA_HOME --keyring-backend test"
# the /8 containing 7.7.7.7 is assigned to the DOD and never routable on the public internet
# we're using it in private to prevent gaia from blacklisting it as unroutable
# and allow local pex
$BIN gentx $ARGS $GAIA_HOME --moniker validator$i --chain-id=$CHAIN_ID --ip 7.7.7.$i validator$i 500000000stake
# obviously we don't need to copy validator1's gentx to itself
if [ $i -gt 1 ]; then
cp /validator$i/config/gentx/* /validator1/config/gentx/
fi
done


$BIN collect-gentxs $STARTING_VALIDATOR_HOME test
GENTXS=$(ls /validator1/config/gentx | wc -l)
cp /validator1/config/genesis.json /genesis.json
echo "Collected $GENTXS gentx"

# put the now final genesis.json into the correct folders
for i in $(seq 1 $NODES);
do
cp /genesis.json /validator$i/config/genesis.json
done