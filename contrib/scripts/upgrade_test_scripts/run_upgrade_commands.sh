#!/bin/bash

set -o errexit -o nounset

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

UPGRADE_HEIGHT=$1

if [ -z "$1" ]; then
  echo "Need to add an upgrade height"
  exit 1
fi

NODE_HOME=$(realpath ./build/.gaia)

echo "NODE_HOME = ${NODE_HOME}"

BINARY=$NODE_HOME/cosmovisor/genesis/bin/gaiad
echo "BINARY = ${BINARY}"
chmod a+x $BINARY

$BINARY version

USER_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"
CHAINID=cosmoshub-4

if test -f "$BINARY"; then

  echo "wait 10 seconds for blockchain to start"
  sleep 10

  $BINARY config chain-id $CHAINID --home $NODE_HOME
  $BINARY config output json --home $NODE_HOME
  $BINARY config keyring-backend test --home $NODE_HOME
  $BINARY config --home $NODE_HOME

  key=$($BINARY keys show val --home $NODE_HOME)
  if [ -z "$key" ]; then
    echo $USER_MNEMONIC | $BINARY --home $NODE_HOME keys add val --recover --keyring-backend=test
  fi

  echo "\n"
  echo "Submitting proposal... \n"
  json_content=$(cat <<EOF
  {
    "messages": [
     {
        "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
        "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
        "plan": {
          "name": "$UPGRADE_VERSION",
          "height": "$UPGRADE_HEIGHT",
          "info": "upgrade",
          "upgraded_client_state": null
        }
      }
    ],
    "metadata": "",
    "deposit": "10000000uatom",
    "title": "$UPGRADE_VERSION",
    "summary": "upgrade"
  }
EOF
  )
  echo "$json_content" > "$NODE_HOME/sw_upgrade_proposal.json"
  $BINARY tx gov submit-proposal "$NODE_HOME/sw_upgrade_proposal.json" \
    --gas auto \
    --gas-adjustment 1.3 \
    --fees 330000uatom \
    --from val \
    --keyring-backend test \
    --chain-id $CHAINID \
    --home $NODE_HOME \
    --node tcp://localhost:26657 \
    --yes
  echo "Done \n"

  sleep 6

  echo "Casting vote... \n"

  $BINARY tx gov vote 1 yes \
    --from val \
    --keyring-backend test \
    --chain-id $CHAINID \
    --home $NODE_HOME \
    --fees 330000uatom \
    --node tcp://localhost:26657 \
    --yes

  echo "Done \n"

  $BINARY q gov proposals \
  --home $NODE_HOME \
  --node tcp://localhost:26657

else
  echo "Please build old gaia binary and move to ./build/gaiadold"
fi
