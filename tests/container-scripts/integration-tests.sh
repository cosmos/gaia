#!/bin/bash
NODES=$1
TEST_TYPE=$2
set -eu

set +e
killall -9 test-runner
set -e

RUST_BACKTRACE=full CHAIN_BINARY=althea ADDRESS_PREFIX=althea STAKING_TOKEN=ualtg TEST_TYPE=$TEST_TYPE RUST_LOG=INFO test-runner
