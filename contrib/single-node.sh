#!/bin/sh

set -o errexit -o nounset

HOME_DIR="${1:-$HOME}"
CHAINID="test-gaia"
USER_COINS="100000000000stake"
STAKE="100000000stake"
MONIKER="gaia-test-node"
GAIAD="gaiad"


echo "Using home dir: $HOME_DIR"
rm -rf $HOME_DIR/.gaia
$GAIAD init --chain-id $CHAINID $MONIKER --home "$HOME_DIR/.gaia"

echo "Setting up genesis file"
jq ".app_state.gov.params.voting_period = \"20s\" | .app_state.gov.params.expedited_voting_period = \"10s\" | .app_state.staking.params.unbonding_time = \"86400s\"" \
   "${HOME_DIR}/.gaia/config/genesis.json" > \
   "${HOME_DIR}/edited_genesis.json" && mv "${HOME_DIR}/edited_genesis.json" "${HOME_DIR}/.gaia/config/genesis.json"

$GAIAD keys add validator --keyring-backend="test"
$GAIAD keys add user --keyring-backend="test"
$GAIAD genesis add-genesis-account $("${GAIAD}" keys show validator -a --keyring-backend="test") $USER_COINS
$GAIAD genesis add-genesis-account $("${GAIAD}" keys show user -a --keyring-backend="test") $USER_COINS
$GAIAD genesis gentx validator $STAKE --keyring-backend="test" --chain-id $CHAINID
$GAIAD genesis collect-gentxs

# Set proper defaults and change ports
echo "Setting up node configs"
# sed -i '' 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:26657"#g' ~/.gaia/config/config.toml
sleep 1
sed -i -r 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ~/.gaia/config/config.toml
sed -i -r 's/timeout_propose = "3s"/timeout_propose = "1s"/g' ~/.gaia/config/config.toml
sed -i -r 's/index_all_keys = false/index_all_keys = true/g' ~/.gaia/config/config.toml
sed -i -r 's/minimum-gas-prices = ""/minimum-gas-prices = "0stake"/g' ~/.gaia/config/app.toml

# Start the gaia
$GAIAD start --api.enable=true
