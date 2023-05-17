#!/bin/bash

NODEADDR=$1

# Define default global fee module's params
# according to gaia/x/globalfee/types/params.go
default_globalfee_params='
{
    "params": {
        "minimum_gas_prices": [],
        "bypass_min_fee_msg_types": [
            "/ibc.core.channel.v1.MsgRecvPacket",
            "/ibc.core.channel.v1.MsgAcknowledgement",
            "/ibc.core.client.v1.MsgUpdateClient",
            "/ibc.core.channel.v1.MsgTimeout",
            "/ibc.core.channel.v1.MsgTimeoutOnClose"
        ],
        "max_total_bypass_min_fee_msg_gas_usage": "1000000"
    }
}'

# Get current global fee default params
curr_params=$(curl -s $NODEADDR:1317/gaia/globalfee/v1beta1/params)

# Check if retrieved params are equal to expected default params
DIFF=$(diff  <(echo ${default_globalfee_params} | jq --sort-keys .) <(echo ${curr_params} | jq --sort-keys .))

if [ "$DIFF" != "" ] 
then
    printf "expected default global fee params:\n${DIFF}"
    exit 1
fi
