#!/bin/bash

BINARY=$(which gaiad)
# please do not reveal your mnemonic in production !!!
MNEMONIC_RLY0=`cat docs/modules/icamauth_scripts/rly0-mnemonic.txt`
MNEMONIC_ALICE="captain six loyal advice caution cost orient large mimic spare radar excess quote orchard error biology choice shop dish master quantum dumb accident between"
CHAINID_0=test-0
HOME_0=$HOME/test-0
P2PPORT_0=16656
RPCPORT_0=16657
GRPCPORT_0=9095
GRPCWEBPORT_0=9081
RESTPORT_0=1316
ROSETTA_0=8080

# Stop if it is already running
if pgrep -x "$BINARY" >/dev/null; then
    echo "Terminating $BINARY..."
    killall gaiad
fi

echo "Removing previous data..."
rm -rf $HOME_0 &> /dev/null

# Add directories for both chains, exit if an error occurs
if ! mkdir -p $HOME_0 2>/dev/null; then
    echo "Failed to create gaiad folder. Aborting..."
    exit 1
fi

echo "Initializing $CHAINID_0..."
$BINARY init test0 --chain-id=$CHAINID_0 --home $HOME_0

$BINARY config chain-id $CHAIN_ID_0 --home $HOME_0
$BINARY config keyring-backend test --home $HOME_0
$BINARY config broadcast-mode block --home $HOME_0
$BINARY config node tcp://localhost:$RPCPORT_0 --home $HOME_0


echo "Adding genesis accounts..."
$BINARY keys add val0  --home=$HOME_0
echo $MNEMONIC_ALICE | $BINARY keys add alice --recover --home=$HOME_0
echo $MNEMONIC_RLY0 | $BINARY keys add rly0 --recover --home=$HOME_0
$BINARY add-genesis-account $($BINARY keys show val0 -a --home=$HOME_0) 100000000000stake --home=$HOME_0
$BINARY add-genesis-account $($BINARY keys show alice -a --home=$HOME_0) 100000000000stake --home=$HOME_0
$BINARY add-genesis-account $($BINARY keys show rly0 -a --home=$HOME_0) 100000000000stake --home=$HOME_0

echo "Creating and collecting gentx..."
$BINARY gentx val0 7000000000stake --chain-id $CHAINID_0 --home=$HOME_0
$BINARY collect-gentxs --home=$HOME_0

echo "Change setups in app.toml and config.toml files..."
sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_0"'"#g' $HOME_0/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_0"'"#g' $HOME_0/config/config.toml
sed -i -e 's/"0.0.0.0:9090"/"0.0.0.0:'"$GRPCPORT_0"'"/g' $HOME_0/config/app.toml
sed -i -e 's/"0.0.0.0:9091"/"0.0.0.0:'"$GRPCWEBPORT_0"'"/g' $HOME_0/config/app.toml
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $HOME_0/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $HOME_0/config/config.toml
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' $HOME_0/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $HOME_0/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $HOME_0/config/app.toml
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$RESTPORT_0"'"#g' $HOME_0/config/app.toml
sed -i -e 's#":8080"#":'"$ROSETTA_0"'"#g' $HOME_0/config/app.toml

#set min_gas_prices
sed -i '' 's/minimum-gas-prices = ""/minimum-gas-prices = "0.025stake"/g' $HOME_0/config/app.toml

# Update host chain genesis to allow all msg types
sed -i '' 's/\"allow_messages\": \[\]/\"allow_messages\": \["*"\]/g' $HOME_0/config/genesis.json

echo "Starting $CHAINID_0..."
echo "Creating log file at gaia0.log"
$BINARY start --home=$HOME_0 --log_level=trace --log_format=json --pruning=nothing
