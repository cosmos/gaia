#!/bin/bash
# Number of validators to start
NODES=$1
# what test to execute
TEST_TYPE=$2
set -eux

# Stop any currently running peggy and eth processes
pkill peggyd || true # allowed to fail
pkill geth || true # allowed to fail

# Wipe filesystem changes
for i in $(seq 1 $NODES);
do
    rm -rf "/validator$i"
done


cd /althea/
export PATH=$PATH:/usr/local/go/bin
make install
tests/container-scripts/setup-validators.sh $NODES
tests/container-scripts/run-testnet.sh $NODES

# deploy the ethereum contracts
DEPLOY_CONTRACTS=1 RUST_BACKTRACE=full TEST_TYPE=$TEST_TYPE RUST_LOG=INFO test-runner

# This keeps the script open to prevent Docker from stopping the container
# immediately if the nodes are killed by a different process
read -p "Press Return to Close..."