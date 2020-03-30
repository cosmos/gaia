#!/bin/sh

CHAINID=$1
GENACCT=$2

if [ -z "$1" ]; then
  echo "Need to input chain id..."
  exit 1
fi

if [ -z "$2" ]; then
  echo "Need to input genesis account address..."
  exit 1
fi

# Build genesis file incl account for passed address
coins="10000000000stake,100000000000samoleans"
gaiad init --chain-id $CHAINID $CHAINID
gaiacli keys add validator --keyring-backend="test"
gaiad add-genesis-account validator $coins --keyring-backend="test"
gaiad add-genesis-account $GENACCT $coins --keyring-backend="test"
gaiad gentx --name validator --keyring-backend="test"
gaiad collect-gentxs

# Set proper defaults and change ports
sed -i 's/"leveldb"/"goleveldb"/g' ~/.gaiad/config/config.toml
sed -i 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:26657"#g' ~/.gaiad/config/config.toml
sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ~/.gaiad/config/config.toml
sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' ~/.gaiad/config/config.toml
sed -i 's/index_all_keys = false/index_all_keys = true/g' ~/.gaiad/config/config.toml

# Start the gaia
gaiad start --pruning=nothing