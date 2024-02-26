# x/metaprotocols module

The `x/metaprotocols` module adds support for encoding and decoding additional fields attached to transactions.

`extension_options` and `non_critical_extension_options` are optional fields that can be used to attach data to valid transactions. The fields are validated by the blockchain, but they are not used in any way. The fields pass validation if they are provided as empty lists (`[ ]`) or they use a list of `ExtensionData` types.

The application does not use the attached data but it does ensure that the correct type is provided and that it can be successfully unmarshalled. The attached data will be part of a block.

Here is an example of a correctly formed `non_critical_extension_options` field:

```json
{
  "@type": "/gaia.metaprotocols.ExtensionData", // must be this exact string
  "protocol_id": "some-protocol",
  "protocol_version": "1",
  "data": "<base64 encoded bytes>"
}
```

Here is an example of a correctly populated `non_critical_extension_options` on a `bank.MsgSend` transaction:

```json
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.bank.v1beta1.MsgSend",
        "from_address": "cosmos1ehpqg9sj09037uhe56sqktk30asn47asthyr22",
        "to_address": "cosmos1ehpqg9sj09037uhe56sqktk30asn47asthyr22",
        "amount": [
          {
            "denom": "uatom",
            "amount": "100"
          }
        ]
      }
    ],
    "memo": "memo_smaller_than_512_bytes",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": [
      {
        "@type": "/gaia.metaprotocols.ExtensionData",
        "protocol_id": "some-protocol",
        "protocol_version": "1",
        "data": "<base64 encoded bytes>"
      }
    ]
  },
  "auth_info": {
    "signer_infos": [],
    "fee": {
      "amount": [],
      "gas_limit": "200000",
      "payer": "",
      "granter": ""
    },
    "tip": null
  },
  "signatures": []
}
```
