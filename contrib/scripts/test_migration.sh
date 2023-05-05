#!bin/bash

V10_DEFAULT_PARAMS='{
    params": {
    "minimum_gas_prices": [
    ],
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