#!/bin/sh

set -o errexit -o nounset

HOMEDIR=./build/.gaiad-v7
BINARY=./build/gaiad7
CHAINID=cosmoshub-4



if test -f "$BINARY"; then
	rm -rf $HOMEDIR
	$BINARY init upgrader --chain-id $CHAINID --home $HOMEDIR
	$BINARY config chain-id $CHAINID --home $HOMEDIR
	$BINARY config keyring-backend test --home $HOMEDIR
  tmp=$(mktemp)

  # add bank part of genesis
  jq --argjson foo "$(jq -c '.' denom.json)" '.app_state.bank.denom_metadata = $foo' $HOMEDIR/config/genesis.json > "$tmp" && mv "$tmp" $HOMEDIR/config/genesis.json

  # replace default stake token with uatom
  sed -i -e 's/stake/uatom/g' $HOMEDIR/config/genesis.json
  # min deposition amount (this one isn't working)
  sed -i -e 's%"amount": "10000000",%"amount": "1",%g' $HOMEDIR/config/genesis.json
  #   min voting power that a proposal requires in order to be a valid proposal
  sed -i -e 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $HOMEDIR/config/genesis.json
  # the minimum proportion of "yes" votes requires for the proposal to pass
  sed -i -e 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $HOMEDIR/config/genesis.json
  # voting period to 60s
  sed -i -e 's%"voting_period": "172800s"%"voting_period": "60s"%g' $HOMEDIR/config/genesis.json


  $BINARY keys add val --home $HOMEDIR --keyring-backend test
  $BINARY add-genesis-account val 10000000000000000000000000uatom --home $HOMEDIR --keyring-backend test
  $BINARY gentx val 1000000000uatom --home $HOMEDIR --chain-id $CHAINID
  $BINARY collect-gentxs --home $HOMEDIR

	sed -i.bak'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uatom"/' $HOMEDIR/config/app.toml
	$BINARY start --home $HOMEDIR --x-crisis-skip-assert-invariants
else
  echo "Please build gaia v7 and move to ./build/gaiad7"
fi
