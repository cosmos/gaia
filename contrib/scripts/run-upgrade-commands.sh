#!/bin/sh

set -o errexit -o nounset

UPGRADE_HEIGHT=$1

if [ -z "$1" ]; then
  echo "Need to add an upgrade height"
  exit 1
fi


# if [ -z "$2" ]; then
#   echo "Need to add an amount of time to wait for upgrade height"
#   exit 1
# fi

# HOMEDIR=./build/.gaia
NODE_HOME=$(realpath ./build/.gaia)
echo "NODE_HOME = ${NODE_HOME}"

BINARY=$NODE_HOME/cosmovisor/genesis/bin/gaiad
echo "BINARY = ${BINARY}"



BINARY=./build/gaiad7
CHAINID=cosmoshub-4

if test -f "$BINARY"; then

  echo "wait 10 seconds for blockchain to start"
  sleep 10


	$BINARY config chain-id $CHAINID --home $HOMEDIR
	$BINARY config output json --home $HOMEDIR
	$BINARY config keyring-backend test --home $HOMEDIR
  $BINARY keys list --home $HOMEDIR
  echo "\n"
  echo "Submitting proposal... \n"
  $BINARY tx gov submit-proposal software-upgrade v8-Rho \
  --title v8-Rho \
  --deposit 10000000uatom \
  --upgrade-height $UPGRADE_HEIGHT \
  --upgrade-info "upgrade to v8-Rho" \
  --description "upgrade to v8-Rho" \
  --gas auto \
  --fees 400uatom \
  --from val \
  --keyring-backend test \
  --chain-id $CHAINID \
  --home $HOMEDIR \
  --node tcp://localhost:26657 \
  --yes
  echo "Done \n"

  sleep 6
  echo "Casting vote... \n"

  $BINARY tx gov vote 1 yes \
  --from val \
  --keyring-backend test \
  --chain-id $CHAINID \
  --home $HOMEDIR \
  --gas auto \
  --fees 400uatom \
  --node tcp://localhost:26657 \
  --yes

  echo "Done \n"
  # echo "Waiting $TIME_TO_SLEEP sec for upgrade height... \n"
  # sleep $TIME_TO_SLEEP

  # ./run-gaia-v8.sh



else
  echo "Please build gaia v7 and move to ./build/gaiad7"
fi
