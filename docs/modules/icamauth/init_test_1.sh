BINARY=gaiad
MNEMONIC_RLY1="record gift you once hip style during joke field prize dust unique length more pencil transfer quit train device arrive energy sort steak upset"
MNEMONIC_BOB="uphold train large action document mixed exact cherry input evil sponsor digital used child engine fire attract sing little jeans decrease despair unfair what"
CHAINID_1=test-1
HOME_1=$HOME/test-1
P2PPORT_1=26656
RPCPORT_1=26657
GRPCPORT_1=9096
GRPCWEBPORT_1=9082
RESTPORT_1=1317
ROSETTA_1=8081

# Stop if it is already running
if pgrep -x  "$BINARY" >/dev/null; then
    echo "Terminating $BINARY..."
    killall gaiad
fi

echo "Removing previous data..."
rm -rf $HOME_1 &> /dev/null

# Add directories for both chains, exit if an error occurs
if ! mkdir -p $HOME_1 2>/dev/null; then
    echo "Failed to create gaiad folder. Aborting..."
    exit 1
fi

echo "Initializing $CHAINID_1..."
gaiad init test1 --chain-id=$CHAINID_1 --home $HOME_1
$BINARY config chain-id $CHAIN_ID_1 --home $HOME_1
$BINARY config keyring-backend test --home $HOME_1
$BINARY config broadcast-mode block --home $HOME_1
$BINARY config node tcp://localhost:$RPCPORT_1 --home $HOME_1


echo "Adding genesis accounts..."
$BINARY keys add val1  --home=$HOME_1
echo $MNEMONIC_BOB | $BINARY keys add bob --recover --home=$HOME_1
echo $MNEMONIC_RLY1 | $BINARY keys add rly1 --recover --home=$HOME_1
$BINARY add-genesis-account $($BINARY keys show val1 -a --home=$HOME_1) 100000000000uatom --home=$HOME_1
$BINARY add-genesis-account $($BINARY keys show bob -a --home=$HOME_1) 100000000000uatom --home=$HOME_1
$BINARY add-genesis-account $($BINARY keys show rly1 -a --home=$HOME_1) 100000000000uatom --home=$HOME_1

echo "Creating and collecting gentx..."
$BINARY gentx val1 7000000000uatom --chain-id $CHAINID_1 --home=$HOME_1
$BINARY collect-gentxs --home=$HOME_1

echo "Change setups in app.toml and config.toml files..."
sed -i -e 's#"tcp://0.0.0.0:26656"#"tcp://0.0.0.0:'"$P2PPORT_1"'"#g' $HOME_1/config/config.toml
sed -i -e 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:'"$RPCPORT_1"'"#g' $HOME_1/config/config.toml
sed -i -e 's/"0.0.0.0:9090"/"0.0.0.0:'"$GRPCPORT_1"'"/g' $HOME_1/config/app.toml
sed -i -e 's/"0.0.0.0:9091"/"0.0.0.0:'"$GRPCWEBPORT_1"'"/g' $HOME_1/config/app.toml
sed -i -e 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $HOME_1/config/config.toml
sed -i -e 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $HOME_1/config/config.toml
sed -i -e 's/index_all_keys = false/index_all_keys = true/g' $HOME_1/config/config.toml
sed -i -e 's/enable = false/enable = true/g' $HOME_1/config/app.toml
sed -i -e 's/swagger = false/swagger = true/g' $HOME_1/config/app.toml
sed -i -e 's#"tcp://0.0.0.0:1317"#"tcp://0.0.0.0:'"$RESTPORT_1"'"#g' $HOME_1/config/app.toml
sed -i -e 's#":8080"#":'"$ROSETTA_1"'"#g' $HOME_1/config/app.toml

#set min_gas_prices
sed -i '' 's/minimum-gas-prices = ""/minimum-gas-prices = "0.025uatom"/g' $HOME_1/config/app.toml

# Update host chain genesis to allow all msg types
sed -i '' 's/\"allow_messages\": \[\]/\"allow_messages\": \["*"\]/g' $HOME_0/config/genesis.json

 echo "Starting $CHAINID_1..."
 echo "Creating log file at gaia1.log"
# $BINARY start --home=$HOME_1 --log_level=trace --log_format=json --pruning=nothing > gaia1.log 2>&1 &
