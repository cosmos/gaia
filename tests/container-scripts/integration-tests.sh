#!/bin/bash
NODES=$1
TEST_TYPE=$2
set -eu

set +e
killall -9 test-runner
set -e

pushd /althea/integration_tests/test_runner
RUST_BACKTRACE=full TEST_TYPE=$TEST_TYPE RUST_LOG=INFO PATH=$PATH:$HOME/.cargo/bin cargo run --release --bin test-runner
