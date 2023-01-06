#!/bin/sh

set -o errexit -o nounset

NODE_HOME=$(realpath ./build/.gaia)
echo "NODE_HOME = ${NODE_HOME}"
BINARY=$NODE_HOME/cosmovisor/genesis/bin/gaiad
echo "BINARY = ${BINARY}"
CHAINID=cosmoshub-4

USER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"

if ! test -f "./build/gaiad7"; then
  echo "gaiad v7 does not exist"
  exit
fi


rm -rf ./build/.gaia

mkdir -p "$NODE_HOME"/cosmovisor/genesis/bin
cp ./build/gaiad7 "$NODE_HOME"/cosmovisor/genesis/bin/gaiad
$BINARY init upgrader --chain-id $CHAINID --home "$NODE_HOME"


if ! test -f "./build/gaiad8"; then
  echo "gaiad v8 does not exist"
  exit
fi

mkdir -p "$NODE_HOME"/cosmovisor/upgrades/v8-Rho/bin
cp ./build/gaiad8 "$NODE_HOME"/cosmovisor/upgrades/v8-Rho/bin/gaiad

GOPATH=$(go env GOPATH)

export DAEMON_NAME=gaiad
export DAEMON_HOME=$NODE_HOME
COSMOVISOR=$GOPATH/bin/cosmovisor


$BINARY config chain-id $CHAINID --home $NODE_HOME
$BINARY config keyring-backend test --home $NODE_HOME
tmp=$(mktemp)

# add bank part of genesis
jq --argjson foo "$(jq -c '.' contrib/denom.json)" '.app_state.bank.denom_metadata = $foo' $NODE_HOME/config/genesis.json > "$tmp" && mv "$tmp" $NODE_HOME/config/genesis.json

# replace default stake token with uatom
sed -i -e 's/stake/uatom/g' $NODE_HOME/config/genesis.json
# min deposition amount (this one isn't working)
sed -i -e 's%"amount": "10000000",%"amount": "1",%g' $NODE_HOME/config/genesis.json
#   min voting power that a proposal requires in order to be a valid proposal
sed -i -e 's%"quorum": "0.334000000000000000",%"quorum": "0.000000000000000001",%g' $NODE_HOME/config/genesis.json
# the minimum proportion of "yes" votes requires for the proposal to pass
sed -i -e 's%"threshold": "0.500000000000000000",%"threshold": "0.000000000000000001",%g' $NODE_HOME/config/genesis.json
# voting period to 30s
sed -i -e 's%"voting_period": "172800s"%"voting_period": "30s"%g' $NODE_HOME/config/genesis.json

echo $USER_MNEMONIC | $BINARY --home $NODE_HOME keys add val --recover --keyring-backend=test
$BINARY add-genesis-account val 10000000000000000000000000uatom --home $NODE_HOME --keyring-backend test
$BINARY gentx val 1000000000uatom --home $NODE_HOME --chain-id $CHAINID
$BINARY collect-gentxs --home $NODE_HOME

sed -i.bak'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uatom"/' $NODE_HOME/config/app.toml

perl -i~ -0777 -pe 's/# Enable defines if the API server should be enabled.
enable = false/# Enable defines if the API server should be enabled.
enable = true/g' $NODE_HOME/config/app.toml

$COSMOVISOR start --home $NODE_HOME --x-crisis-skip-assert-invariants

