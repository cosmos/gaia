#!/bin/bash

HERMES_BINARY=hermes
# CONFIG_DIR=./rly-config.toml
CONFIG_DIR=./rly-config.toml

# Sleep is needed otherwise the relayer crashes when trying to init
### Restore Keys
$HERMES_BINARY --config $CONFIG_DIR keys add --key-name hermes-rly0 --chain test-0 --mnemonic-file ./rly0-mnemonic.txt

echo "sleeping"
sleep 5

$HERMES_BINARY --config $CONFIG_DIR keys add --key-name hermes-rly1 --chain test-1 --mnemonic-file ./rly1-mnemonic.txt

echo "sleeping"
sleep 5

### Configure the clients and connection
echo "Initiating connection handshake..."
$HERMES_BINARY --config $CONFIG_DIR create connection --a-chain test-0 --b-chain test-1

echo "sleeping"
sleep 5

echo "Starting hermes relayer..."
$HERMES_BINARY --config $CONFIG_DIR start
