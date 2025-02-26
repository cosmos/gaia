#!/bin/bash

set -o errexit -o nounset

# find the highest upgrade version number($UPGRADE_VERSION_NUMBER) within the 'app/upgrades' dir.
# the highest upgrade version is used to propose upgrade and create /cosmovisor/upgrades/$UPGRADE_VERSION/bin dir.
UPGRADES_DIR=$(realpath ./app/upgrades)
UPGRADE_VERSION="v0"

version_gt() {
  IFS='_' read -ra LEFT_PARTS <<< "${1#v}"
  IFS='_' read -ra RIGHT_PARTS <<< "${2#v}"

  for ((i=0; i < ${#LEFT_PARTS[@]} || i < ${#RIGHT_PARTS[@]}; i++)); do
    LEFT_NUM=${LEFT_PARTS[i]:-0}  # Default to 0 if missing
    RIGHT_NUM=${RIGHT_PARTS[i]:-0} # Default to 0 if missing

    if (( LEFT_NUM > RIGHT_NUM )); then
      return 0  # Left is greater
    elif (( LEFT_NUM < RIGHT_NUM )); then
      return 1  # Right is greater
    fi
  done

  return 1  # Equal versions, so not greater
}

for dir in "$UPGRADES_DIR"/*; do
  if [ -d "$dir" ]; then
    DIR_NAME=$(basename "$dir")

    if version_gt "$DIR_NAME" "$UPGRADE_VERSION"; then
      UPGRADE_VERSION="$DIR_NAME"
    fi
  fi
done

# Convert "_" to "." in the final output
UPGRADE_VERSION="${UPGRADE_VERSION//_/.}"

echo "Latest upgrade version: $UPGRADE_VERSION"
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
chmod a+x $BINARY
$BINARY init upgrader --chain-id $CHAINID --home "$NODE_HOME"

if ! test -f "./build/gaiadnew"; then
  echo "new gaiad binary does not exist"
  exit
fi

mkdir -p "$NODE_HOME"/cosmovisor/upgrades/$UPGRADE_VERSION/bin
cp ./build/gaiadnew "$NODE_HOME"/cosmovisor/upgrades/$UPGRADE_VERSION/bin/gaiad
chmod a+x "$NODE_HOME"/cosmovisor/upgrades/$UPGRADE_VERSION/bin/gaiad

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
