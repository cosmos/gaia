#!/bin/sh

set -o errexit -o nounset

UPGRADE_HEIGHT=$1
SLEEP=$2

if [ -z "$1" ]; then
  echo "Need to add an upgrade height"
  exit 1
fi

# NODE_HOME=./build/.gaia
NODE_HOME=$(realpath ./build/.gaia)
echo "NODE_HOME = ${NODE_HOME}"

BINARY=$NODE_HOME/cosmovisor/genesis/bin/gaiad
echo "BINARY = ${BINARY}"

USER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
CHAINID=local-testnet

if test -f "$BINARY"; then

  echo "wait $SLEEP seconds for blockchain to start"
  sleep $SLEEP

	$BINARY config chain-id $CHAINID --home $NODE_HOME
	$BINARY config output json --home $NODE_HOME
	$BINARY config keyring-backend test --home $NODE_HOME
  $BINARY config --home $NODE_HOME


  key=$($BINARY keys show val --home $NODE_HOME)
  if [ key == "" ]; then
    echo $USER_MNEMONIC | $BINARY --home $NODE_HOME keys add val --recover --keyring-backend=test
  fi

  # $BINARY keys list --home $NODE_HOME

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
  --home $NODE_HOME \
  --node tcp://localhost:26657 \
  --yes
  echo "Done \n"

  sleep 6

  PROPOSAL_ID=$($BINARY q gov proposals --home $NODE_HOME --output json | jq -r '.proposals | last.proposal_id')


  echo "Casting vote for proposal $PROPOSAL_ID... \n"


  $BINARY tx gov vote $PROPOSAL_ID yes \
  --from val \
  --keyring-backend test \
  --chain-id $CHAINID \
  --home $NODE_HOME \
  --gas auto \
  --fees 400uatom \
  --node tcp://localhost:26657 \
  --yes

  echo "Done \n"

else
  echo "Please build gaia v7 and move to ./build/gaiad7"
fi
