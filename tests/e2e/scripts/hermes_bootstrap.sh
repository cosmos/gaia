#!/bin/bash

set -ex

# initialize Hermes relayer configuration
mkdir -p /root/.hermes/
touch /root/.hermes/config.toml

echo $GAIA_B_E2E_RLY_MNEMONIC > /root/.hermes/GAIA_B_E2E_RLY_MNEMONIC.txt
echo $GAIA_A_E2E_RLY_MNEMONIC > /root/.hermes/GAIA_A_E2E_RLY_MNEMONIC.txt

# setup Hermes relayer configuration
tee /root/.hermes/config.toml <<EOF
[global]
log_level = 'info'

[mode]

[mode.clients]
enabled = true
refresh = true
misbehaviour = true

[mode.connections]
enabled = false

[mode.channels]
enabled = true

[mode.packets]
enabled = true
clear_interval = 100
clear_on_start = true
tx_confirmation = true

[rest]
enabled = true
host = '0.0.0.0'
port = 3031

[telemetry]
enabled = true
host = '127.0.0.1'
port = 3001

[[chains]]
id = '$GAIA_A_E2E_CHAIN_ID'
rpc_addr = 'http://$GAIA_A_E2E_VAL_HOST:26657'
grpc_addr = 'http://$GAIA_A_E2E_VAL_HOST:9090'
websocket_addr = 'ws://$GAIA_A_E2E_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'cosmos'
key_name = 'rly01-gaia-a'
store_prefix = 'ibc'
max_gas = 6000000
gas_price = { price = 0.00001, denom = 'uatom' }
gas_multiplier = 1.2
clock_drift = '1m' # to accomdate docker containers
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }

[[chains]]
id = '$GAIA_B_E2E_CHAIN_ID'
rpc_addr = 'http://$GAIA_B_E2E_VAL_HOST:26657'
grpc_addr = 'http://$GAIA_B_E2E_VAL_HOST:9090'
websocket_addr = 'ws://$GAIA_B_E2E_VAL_HOST:26657/websocket'
rpc_timeout = '10s'
account_prefix = 'cosmos'
key_name = 'rly01-gaia-b'
store_prefix = 'ibc'
max_gas =  6000000
gas_price = { price = 0.00001, denom = 'uatom' }
gas_multiplier = 1.2
clock_drift = '1m' # to accomdate docker containers
trusting_period = '14days'
trust_threshold = { numerator = '1', denominator = '3' }
EOF

# import keys
hermes keys add  --key-name rly01-gaia-b  --chain $GAIA_B_E2E_CHAIN_ID --mnemonic-file /root/.hermes/GAIA_B_E2E_RLY_MNEMONIC.txt
sleep 5
hermes keys add  --key-name rly01-gaia-a  --chain $GAIA_A_E2E_CHAIN_ID --mnemonic-file /root/.hermes/GAIA_A_E2E_RLY_MNEMONIC.txt
sleep 5
# start Hermes relayer
hermes start
