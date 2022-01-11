#!/usr/bin/env bash

## Programmatic list for creating Gaia Hub chains for testing IBC.
## Instead of blindly running this code, read it line by line and understand the dependencies and tasks.
## Prerequisites: Log into Docker Hub
set -eou pipefail

## After updating the gaia version below, double-check the following (see readme.md also):
##   - the new version made it to docker hub, and is available for download, e.g. `docker pull informaldev/ibc-1:v4.0.0`
##   - the image versions and the relayer release in `docker-compose.yml` are consistent with the new version
GAIA_BRANCH="v5.0.5" # Requires a version with the `--keyring-backend` option. v2.1 and above.

# Check if gaiad is installed and if the versions match
if ! [ -x "$(command -v gaiad)" ]; then
  echo 'Error: gaiad is not installed.' >&2
  exit 1
fi

CURRENT_GAIA="$(gaiad version)"
echo "Current Gaia Version: $CURRENT_GAIA"

if [ "$GAIA_BRANCH" != "$CURRENT_GAIA" ]; then
  echo "Error: gaiad installed is different than target gaiad ($CURRENT_GAIA != $GAIA_BRANCH)"
  exit 1
else
  echo "Gaiad installed matches desired version ($CURRENT_GAIA = $GAIA_BRANCH)"
fi

BASE_DIR="$(dirname $0)"
ONE_CHAIN="$BASE_DIR/../scripts/one-chain"

echo "*** Building config folders"

CHAIN_HOME="./chains/gaia/$GAIA_BRANCH"

# Clean home dir if exists
rm -Rf "$CHAIN_HOME"

# Create home dir
mkdir -p "$CHAIN_HOME"

ls -allh "$CHAIN_HOME"

# Check gaia version
echo "-------------------------------------------------------------------------------------------------------------------"
echo "Gaiad version"
echo "-------------------------------------------------------------------------------------------------------------------"
gaiad version --long  --log_level error

MONIKER=node_ibc_0
CHAIN_ID=ibc-0
CHAIN_IP=172.25.0.10
RPC_PORT=26657
GRPC_PORT=9090
CHAIN_SAMOLEANS=100000000000
"$ONE_CHAIN" gaiad "$CHAIN_ID" "$CHAIN_HOME" "$RPC_PORT" 26656 6060 "$GRPC_PORT" "$CHAIN_SAMOLEANS"

MONIKER=node_ibc_1
CHAIN_ID=ibc-1
CHAIN_IP=172.25.0.11
RPC_PORT=26657
GRPC_PORT=9090
CHAIN_SAMOLEANS=100000000000
"$ONE_CHAIN" gaiad "$CHAIN_ID" "$CHAIN_HOME" "$RPC_PORT" 26656 6060 "$GRPC_PORT" "$CHAIN_SAMOLEANS"

echo "*** Requirements"
which docker

echo "*** Create Docker image and upload to Docker Hub"
docker build --build-arg CHAIN=gaia --build-arg RELEASE=$GAIA_BRANCH --build-arg NAME=ibc-0 -f --no-cache -t informaldev/ibc-0:$GAIA_BRANCH -f "$BASE_DIR/gaia.Dockerfile" .
docker build --build-arg CHAIN=gaia --build-arg RELEASE=$GAIA_BRANCH --build-arg NAME=ibc-1 -f --no-cache -t informaldev/ibc-1:$GAIA_BRANCH -f "$BASE_DIR/gaia.Dockerfile" .

read -p "Press ANY KEY to push image to Docker Hub, or CTRL-C to cancel. " dontcare
docker push informaldev/ibc-0:$GAIA_BRANCH
docker push informaldev/ibc-1:$GAIA_BRANCH
