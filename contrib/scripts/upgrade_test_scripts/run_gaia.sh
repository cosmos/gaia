#!/bin/sh

set -o errexit -o nounset

# find the highest upgrade version number($UPGRADE_VERSION_NUMBER) within the 'app/upgrades' dir.
# the highest upgrade version is used to propose upgrade and create /cosmovisor/upgrades/$UPGRADE_VERSION/bin dir.
UPGRADES_DIR=$(realpath ./app/upgrades)
UPGRADE_VERSION_NUMBER=0

for dir in "$UPGRADES_DIR"/*; do
  if [ -d "$dir" ]; then
    DIR_NAME=$(basename "$dir")
    VERSION_NUMBER="${DIR_NAME#v}"
    if [ "$VERSION_NUMBER" -gt "$UPGRADE_VERSION_NUMBER" ]; then
      UPGRADE_VERSION_NUMBER=$VERSION_NUMBER
    fi
  fi
done

if [ -n "$UPGRADE_VERSION_NUMBER" ]; then
  echo "Upgrade to version: $UPGRADE_VERSION_NUMBER"
else
  echo "No upgrade version found in app/upgrades."
fi

UPGRADE_VERSION=v$UPGRADE_VERSION_NUMBER
NODE_HOME=$(realpath ./build/.gaia)
echo "NODE_HOME = ${NODE_HOME}"
BINARY=$NODE_HOME/cosmovisor/genesis/bin/gaiad
echo "BINARY = ${BINARY}"
CHAINID=cosmoshub-4

USER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"

if ! test -f "./build/gaiadold"; then
  echo "old gaiad binary does not exist"
  exit
fi

rm -rf ./build/.gaia

mkdir -p "$NODE_HOME"/cosmovisor/genesis/bin
cp ./build/gaiadold "$NODE_HOME"/cosmovisor/genesis/bin/gaiad
$BINARY init upgrader --chain-id $CHAINID --home "$NODE_HOME"

if ! test -f "./build/gaiadnew"; then
  echo "new gaiad binary does not exist"
  exit
fi

mkdir -p "$NODE_HOME"/cosmovisor/upgrades/$UPGRADE_VERSION/bin
cp ./build/gaiadnew "$NODE_HOME"/cosmovisor/upgrades/$UPGRADE_VERSION/bin/gaiad

GOPATH=$(go env GOPATH)

export DAEMON_NAME=gaiad
export DAEMON_HOME=$NODE_HOME
COSMOVISOR=$GOPATH/bin/cosmovisor

$BINARY config set client chain-id $CHAINID --home $NODE_HOME
$BINARY config set client keyring-backend test --home $NODE_HOME
tmp=$(mktemp)

# add bank part of genesis
jq --argjson foo "$(jq -c '.' contrib/denom.json)" '.app_state.bank.denom_metadata = $foo' $NODE_HOME/config/genesis.json >"$tmp" && mv "$tmp" $NODE_HOME/config/genesis.json
jq ".app_state.gov.params.expedited_voting_period = \"10s\"" "$NODE_HOME/config/genesis.json" > "$tmp" && mv "$tmp" $NODE_HOME/config/genesis.json


# replace default stake token with uatom
sed -i -e '/total_liquid_staked_tokens/!s/stake/uatom/g' $NODE_HOME/config/genesis.json

# min deposition amount (this one isn't working)
sed -i -e 's/"amount": "10000000",/"amount": "1",/g' $NODE_HOME/config/genesis.json
#   min voting power that a proposal requires in order to be a valid proposal
sed -i -e 's/"quorum": "0.334000000000000000",/"quorum": "0.000000000000000001",/g' $NODE_HOME/config/genesis.json
# the minimum proportion of "yes" votes requires for the proposal to pass
sed -i -e 's/"threshold": "0.500000000000000000",/"threshold": "0.000000000000000001",/g' $NODE_HOME/config/genesis.json
# voting period to 30s
sed -i -e 's/"voting_period": "172800s"/"voting_period": "30s"/g' $NODE_HOME/config/genesis.json

echo $USER_MNEMONIC | $BINARY --home $NODE_HOME keys add val --recover --keyring-backend=test
$BINARY genesis add-genesis-account val 10000000000000000000000000uatom --home $NODE_HOME --keyring-backend test
$BINARY genesis gentx val 1000000000uatom --home $NODE_HOME --chain-id $CHAINID
$BINARY genesis collect-gentxs --home $NODE_HOME

sed -i.bak'' 's/minimum-gas-prices = ""/minimum-gas-prices = "0uatom"/' $NODE_HOME/config/app.toml

perl -i~ -0777 -pe 's/# Enable defines if the API server should be enabled.
enable = false/# Enable defines if the API server should be enabled.
enable = true/g' $NODE_HOME/config/app.toml

pwd
ls $NODE_HOME

$COSMOVISOR run start --home $NODE_HOME --x-crisis-skip-assert-invariants >log.out 2>&1 &
