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


CHECK=$(shasum -a 256 $NODE_HOME/config/genesis.json | awk '{print $1}')
ALREADY_DOWNLOADED=false
if [[ $CHECK = "f1d17c898df187c99a98f02e84fe9129ab92ab8b1b99bdbf53ca898d6f02fe94" ]]; then
  ALREADY_DOWNLOADED=true
  cp $NODE_HOME/config/genesis.json ./build/genesis.json
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

$BINARY config broadcast-mode block --home $NODE_HOME
$BINARY config chain-id $CHAINID --home $NODE_HOME
$BINARY config keyring-backend test --home $NODE_HOME

# Get the correct genesis file

if [[ $ALREADY_DOWNLOADED = "true" ]]; then
  cp ./build/genesis.json $NODE_HOME/config/genesis.json
else
  wget https://files.polypore.xyz/genesis/mainnet-genesis-tinkered/tinkered-genesis_2022-09-11T07%3A43%3A05.6452382Z_v7.0.3_12010083.json.gz
  gunzip tinkered-genesis_2022-09-11T07:43:05.6452382Z_v7.0.3_12010083.json.gz 
  mv tinkered-genesis_2022-09-11T07:43:05.6452382Z_v7.0.3_12010083.json $NODE_HOME/config/genesis.json
fi

CHECK=$(shasum -a 256 $NODE_HOME/config/genesis.json | awk '{print $1}')

if [[ $CHECK != "f1d17c898df187c99a98f02e84fe9129ab92ab8b1b99bdbf53ca898d6f02fe94" ]]; then
  echo "SHA256 mismatch for genesis.json CHECK=($CHECK)"
  exit
else
  echo "Using local testnet genesis.json from polypore.xyz"
fi

# Replace the validator and node keys
wget https://raw.githubusercontent.com/cosmos/testnets/master/local/priv_validator_key.json
mv priv_validator_key.json $NODE_HOME/config/priv_validator_key.json
wget https://raw.githubusercontent.com/cosmos/testnets/master/local/node_key.json
mv node_key.json $NODE_HOME/config/node_key.json

# add the user account that has over 75% tokens bonded to our validator

echo $USER_MNEMONIC | $BINARY --home $NODE_HOME keys add val --recover --keyring-backend=test

# Set the minimum gas prices to 0
sed -i.bak'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uatom"/' $NODE_HOME/config/app.toml

# Set block sync to be false. This allow us to achieve liveness without additional peers.
# For details see https://github.com/osmosis-labs/osmosis/issues/735
sed -i -e '/fast_sync =/ s/= .*/= false/' $NODE_HOME/config/config.toml

# Enable the API server
perl -i~ -0777 -pe 's/# Enable defines if the API server should be enabled.
enable = false/# Enable defines if the API server should be enabled.
enable = true/g' $NODE_HOME/config/app.toml

$COSMOVISOR start --home $NODE_HOME --x-crisis-skip-assert-invariants

