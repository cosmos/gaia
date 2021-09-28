# `gentx`

To generate your `gentx` run the following command with the launch genesis file at `~/.gaiad/config/genesis.json`:

```
$ gaiad gentx \
  --amount <amount_of_delegation> \
  --min-self-delegation <min_self_delegation> \
  --commission-rate <commission_rate> \
  --commission-max-rate <commission_max_rate> \
  --commission-max-change-rate <commission_max_change_rate> \
  --pubkey <consensus_pubkey> \
  --name <key_name>
```

This will produce a file in the `~/.gaiad/config/gentx/` folder that has a name with the format `gentx-<node_id>.json`. The content of the file should have a structure as follows:

```json
{
  "type": "auth/StdTx",
  "value": {
    "msg": [
      {
        "type": "cosmos-sdk/MsgCreateValidator",
        "value": {
          "description": {
            "moniker": "<moniker>",
            "identity": "",
            "website": "",
            "details": ""
          },
          "commission": {
            "rate": "<commission_rate>",
            "max_rate": "<commission_max_rate>",
            "max_change_rate": "<commission_max_change_rate>"
          },
          "min_self_delegation": "1",
          "delegator_address": "cosmos1msz843gguwhqx804cdc97n22c4lllfkk39qlnc",
          "validator_address": "cosmosvaloper1msz843gguwhqx804cdc97n22c4lllfkk5352lt",
          "pubkey": "<consensus_pubkey>",
          "value": {
            "denom": "uatom",
            "amount": "1000000"
          }
        }
      }
    ],
    "fee": {
      "amount": null,
      "gas": "200000"
    },
    "signatures": [
      {
        "pub_key": {
          "type": "tendermint/PubKeySecp256k1",
          "value": "AlT62zuYGlZGUG3Yv0RtIFoPTzVY4N+WEFmBvz1syjws"
        },
        "signature": "ZgoOHWB90GIh++kZKWDv8mZok2nQnVcEyEWM6paafFs2ieu4GfAwdjnxsx608LD6+i63kRPRFJv8E81bSSL92A=="
      }
    ],
    "memo": "<node_id>@<ip>:26656"
  }
}
```

To submit your `gentx` for inclusion in genesis, open a pull request against this repository and place the contents in a file `/gentx/<moniker>.json`.

__**NOTE**__: If you would like to override the memo field use the `--ip` and `--node-id` flags for the `gaiad gentx` command above.
