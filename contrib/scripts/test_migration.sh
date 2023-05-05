#!/bin/bash

NODEADDR=$1
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


curr_params=$(curl -s $NODEADDR:1317/gaia/globalfee/v1beta1/params)


DIFF=$(diff  <(echo ${default_globalfee_params} | jq --sort-keys .) <(echo ${curr_params} | jq --sort-keys .))

if [ "$DIFF" != "" ] 
then
    printf "expected default global fee params:\n${DIFF}"
fi
