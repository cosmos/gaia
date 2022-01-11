#!/bin/sh

# This file can be used to initialize a chain

# coins to add to each account
coins="100000000000stake,100000000000samoleans"
STAKE="100000000000stake"
# - the user also needs stake to perform actions
USER_COINS="${STAKE},${SAMOLEANS}samoleans"
#home="/chain"

echo Node: "$MONIKER"
echo Chain: "$CHAIN_ID"
echo Chain IP: "$CHAIN_IP"
echo RPC Port: "$RPC_PORT"
echo GRPC Port: "$GRPC_PORT"
echo Home_Dir: "$CHAIN_HOME"

# Clean home dir if exists
rm -Rf "$CHAIN_HOME"

# Create home dir
mkdir -p "$CHAIN_HOME"

ls -allh "$CHAIN_HOME"

# Check gaia version
echo "-------------------------------------------------------------------------------------------------------------------"
echo "Gaiad version"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad version --long

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Initialize chain $CHAIN_ID"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad init "$MONIKER" --chain-id "$CHAIN_ID" --home "$CHAIN_HOME"

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Replace addresses and ports in the config file and some performance tweaks"
echo "-------------------------------------------------------------------------------------------------------------------"
sed -i 's#"tcp://127.0.0.1:26657"#"tcp://'"$CHAIN_IP"':'"$RPC_PORT"'"#g' "$CHAIN_HOME"/config/config.toml
#sed -i 's#"tcp://0.0.0.0:26656"#"tcp://'"$CHAIN_ID"':'"$P2P_PORT"'"#g' "$CHAIN_HOME"/config/config.toml
#sed -i 's#grpc_laddr = ""#grpc_laddr = "tcp://'"$CHAIN_IP"':'"$GRPC_PORT"'"#g' "$CHAIN_HOME"/config/config.toml
sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' "$CHAIN_HOME"/config/config.toml
sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' "$CHAIN_HOME"/config/config.toml
sed -i 's/index_all_keys = false/index_all_keys = true/g' "$CHAIN_HOME"/config/config.toml

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Adding validator key"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad keys add validator --keyring-backend="test" --home "$CHAIN_HOME" --output json > "$CHAIN_HOME"/validator_seed.json
cat "$CHAIN_HOME"/validator_seed.json

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Adding user key"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad keys add user --keyring-backend="test" --home $CHAIN_HOME --output json > "$CHAIN_HOME"/user_seed.json
cat "$CHAIN_HOME"/user_seed.json

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Adding user2 key"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad keys add user2 --keyring-backend="test" --home $CHAIN_HOME --output json > "$CHAIN_HOME"/user2_seed.json
cat "$CHAIN_HOME"/user2_seed.json

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Adding user account to genesis"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad --home "$CHAIN_HOME" add-genesis-account $(gaiad --home "$CHAIN_HOME" keys --keyring-backend="test" show user -a) 1000000000stake
echo "Done!"

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Adding user2 account to genesis"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad --home "$CHAIN_HOME" add-genesis-account $(gaiad --home "$CHAIN_HOME" keys --keyring-backend="test" show user2 -a) 1000000000stake
echo "Done!"

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Adding validator account to genesis"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad --home "$CHAIN_HOME" add-genesis-account $(gaiad --home "$CHAIN_HOME" keys --keyring-backend="test" show validator -a) 1000000000stake,1000000000validatortoken
echo "Done!"

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Generate a genesis transaction that creates a validator with a self-delegation"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad --home "$CHAIN_HOME" gentx validator 1000000000stake --keyring-backend="test" --chain-id "$CHAIN_ID"
echo "Done!"

echo "-------------------------------------------------------------------------------------------------------------------"
echo "Collect genesis txs and output a genesis.json file"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad collect-gentxs --home "$CHAIN_HOME"
echo "Done!"
