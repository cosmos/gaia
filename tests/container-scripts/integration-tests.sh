#!/bin/bash
NODES=$1
TEST_TYPE=$2
set -eu

FILE=/contracts
if test -f "$FILE"; then
echo "Contracts already deployed, running tests"
else 
echo "Testnet is not started yet, please wait before running tests"
exit 0
fi 

set +e
killall -9 test-runner
set -e

RUST_BACKTRACE=full CHAIN_BINARY=althea TEST_TYPE=$TEST_TYPE RUST_LOG=INFO test-runner
