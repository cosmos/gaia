# Gaia Fee and Fees Checks

## Fee Parameters
The CosmosHub allows managing fees using 4 parameters. At the network level, there are three parameters from globalfee modules (`MinimumGasPricesParam`, `BypassMinFeeMsgTypes`, and `MaxTotalBypassMinFeeMsgGasUsage`) that can be set by gov proposal. Additionally, a fourth parameter which enables individual nodes to impose supplementary fee amount.

1. global fees (`MinimumGasPricesParam`).\
global fees `MinimumGasPricesParam` is established at the network level through globalfee params set via Governance Proposal, it sets a fee requirements that the entire network must adhere to.

   *Please note: in this context, "globalfee" or "Globalfee" are used to refer to the globalfee module, while "global fees" is referring to the `MinimumGasPricesParam` in the globalfee module's params.*


2. `minimum-gas-prices` in `app.toml`\
   By adjusting the `minimum-gas-prices` parameter in `app.toml`, nodes can enforce a fee that is higher than the globally defined `MinimumGasPricesParam`. However, it's importantht to note that this configuration solely determines whether transactions are eligible to enter this specific node's mempool.

    *Please note: in this context, `minimum-gas-prices` are used to refer to the local fee requirement that nodes can set in their `app.toml`, while `MinimumGasPricesParam` is a parameter in the globalfee module, which is the fee requirement at network level.*


3. `BypassMinFeeMsgTypes` and `MaxTotalBypassMinFeeMsgGasUsage`.\
 These two parameters are also part of the globalfee params from gaiad v11.0.0. They can be changed through Gov Proposals. `BypassMinFeeMsgTypes` represents a list of message types that will be excluded from paying any fees for inclusion in a block, `MaxTotalBypassMinFeeMsgGasUsage` is the limit placed on gas usage for `BypassMinFeeMsgTypes`.

## Globalfee module

The globalfee module has three parameters that can be set by governance proposal type `param-change`:
- `MinimumGasPricesParam`
- `BypassMinFeeMsgTypes` 
- `MaxTotalBypassMinFeeMsgGasUsage`

### Globalfee Params: `MinimumGasPricesParam`

Network level, global fees consist of a list of [`sdk.DecCoins`](https://github.com/cosmos/cosmos-sdk/blob/82ce891aa67f635f3b324b7a52386d5405c5abd0/types/dec_coin.go#L158).
Every transaction must pay per unit of gas, **at least**, in one of the denominations (denoms) amounts in the list. This allows the globalfee module to impose a minimum transaction fee for all transactions for a network.

Requirements for the fees include:
- fees have to be alphabetically sorted by denom
- fees must have non-negative amount, with a valid and unique denom (i.e. no duplicate denoms are allowed)

There are **two exceptions** from the global fees rules that allow zero fee transactions:

1. Transactions that contain only message types that can bypass the minimum fee requirement and for which the total gas usage of these bypass messages does not exceed `maxTotalBypassMinFeeMsgGasUsage` may have zero fees. We refer to this as _bypass transactions_.

2. One of the entries in the global fees list has a zero amount, e.g., `0uatom`, and the corresponding denom, e.g., `uatom`, is not present in `minimum-gas-prices` in `app.toml`, or node operators may set additional `minimum-gas-prices` in `app.toml` also zero coins.

### Globalfee Params: `BypassMinFeeMsgTypes` and `MaxTotalBypassMinFeeMsgGasUsage`

Bypass minimum fee messages are messages that are exempt from paying fees. The above global fees and the below local `minimum-gas-prices` checks do not apply for transactions that satisfy the following conditions:

- Transaction contains only bypass message types defined in `BypassMinFeeMsgTypes`.
- The total gas used is less than or equal to `MaxTotalBypassMinFeeMsgGasUsage`.
- In case of non-zero transaction fees, the denom has to be a subset of denoms defined in the global fees list.

Starting from gaiad `v11.0.0`, `BypassMinFeeMsgTypes` and `MaxTotalBypassMinFeeMsgGasUsage` are part of global fee params and can be proposed at network level. The default values are: `bypass-min-fee-msg-types=[
"/ibc.core.channel.v1.MsgRecvPacket",
"/ibc.core.channel.v1.MsgAcknowledgement",
"/ibc.core.client.v1.MsgUpdateClient",
"/ibc.core.channel.v1.MsgTimeout",
"/ibc.core.channel.v1.MsgTimeoutOnClose"
]` and default `maxTotalBypassMinFeeMsgGasUsage=1,000,000`

From gaiad v11.0.0, nodes that have the `bypass-min-fee-msg-types` field in their `app.toml` configuration are **not utilized**. Therefore, node operators have the option to either leave the field in their configurations or remove it. Node inited by gaiad v11.0.0 or later does not have `bypass-min-fee-msg-types` field in the `app.toml`.

Before gaiad `v11.0.0`, `bypass-min-fee-msg-types` can be set by each node in `app.toml`, and [the bypass messages gas usage on average should not exceed `maxBypassMinFeeMsgGasUsage`=200,000](https://github.com/cosmos/gaia/blob/682770f2410ab0d33ac7f0c7203519d7a99fa2b6/x/globalfee/ante/fee.go#L69).

- Nodes created using Gaiad `v7.0.2` - `v10.0.x` use `["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement","/ibc.applications.transfer.v1.MsgTransfer"]` as defaults. 
- Nodes created using Gaiad `v11.0.x` or later use `["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement","/ibc.applications.transfer.v1.MsgTransfer", "/ibc.core.channel.v1.MsgTimeout", "/ibc.core.channel.v1.MsgTimeoutOnClose"]` as defaults. 
- Node Nodes with `bypass-min-fee-msg-types = []` or missing this field in `app.toml` also use default bypass message types.
- Nodes created using gaiad `v7.0.1` and `v7.0.0` do not have `bypass-min-fee-msg-types` configured in `config/app.toml` - they are also using same default values as in `v7.0.2`. The `bypass-min-fee-msg-types` config option can be added to `config/app.toml` before the `[telemetry]` field.

An example of `bypass-min-fee-msg-types` in `app.toml`  **before** gaiad v11.0.0:

```

###############################################################################
###                        Custom Gaia Configuration                        ###
###############################################################################
# bypass-min-fee-msg-types defines custom message types the operator may set that
# will bypass minimum fee checks during CheckTx.
#
# Example:
# ["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement", ...]
bypass-min-fee-msg-types = ["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement","/ibc.applications.transfer.v1.MsgTransfer", "/ibc.core.channel.v1.MsgTimeout", "/ibc.core.channel.v1.MsgTimeoutOnClose"]
```


## `Minimum-gas-prices` (local fee requirement)

The `minimum-gas-prices` parameter enables node operators to set its minimum fee requirements, and it can be set in the `config/app.toml` file.  Please note: if `minimum-gas-prices` is set to include zero coins, the zero coins are sanitized when [`SetMinGasPrices`](https://github.com/cosmos/gaia/blob/76dea00bd6d11bfef043f6062f41e858225820ab/cmd/gaiad/cmd/root.go#L221).
When setting `minimum-gas-prices`, it's important to keep the following rules in mind:

- The denoms in `min-gas-prices` that are not present in the global fees list are ignored. 
- The amounts in `min-gas-prices` that are lower than global fees `MinimumGasPricesParam` are ignored.
- The amounts in `min-gas-prices` are considered as fee requirement only if they are greater than the amounts for the corresponding denoms in the global fees list.  

## Fee AnteHandler Behaviour

The denoms in the global fees list and the `minimum-gas-prices` param are merged and de-duplicated while keeping the higher amounts. Denoms that are only in the `minimum-gas-prices` param are discarded. 

If the denoms of the transaction fees are a subset of the merged fees and at least one of the amounts of the transaction fees is greater than or equal to the corresponding required fees amount, the transaction can pass the fee check, otherwise an error will occur.

## Queries

CLI queries can be used to retrieve the globalfee params:

```shell
gaiad q globalfee params

{
  "minimum_gas_prices": [
    {
      "denom": "uatom",
      "amount": "0.002000000000000000"
    },
  ],
  "bypass_min_fee_msg_types": [
    "/ibc.core.channel.v1.MsgRecvPacket",
    "/ibc.core.channel.v1.MsgAcknowledgement",
    "/ibc.core.client.v1.MsgUpdateClient",
    "/ibc.core.channel.v1.MsgTimeout",
    "/ibc.core.channel.v1.MsgTimeoutOnClose"
  ],
  "max_total_bypass_min_fee_msg_gas_usage": "2000000"
}
```

If the global fees `MinimumGasPricesParam` is not set, the query returns an empty global fees list: `minimum_gas_prices: []`. In this case the Cosmos Hub will use `0uatom` as global fee in this case (the default fee denom).

## Setting Up Globalfee Params via Gov Proposals

An example of setting up a global fee by a gov proposals is shown below.
  
```shell
gov submit-proposal param-change proposal.json
````

A `proposal.json` example to change the `MinimumGasPricesParam` in globalfee params:

```
{
  "title": "Global fee Param Change",
  "description": "Update global fee",
  "changes": [
    {
      "subspace": "globalfee",
      "key": "MinimumGasPricesParam",
      "value": [{"denom":"stake", "amount":"0.002"}, {"denom":"uatom", "amount": "0.001"}]
    }
  ],
  "deposit": "1000stake"
}
```
**Note:** in the above "value" field, coins must sorted alphabetically by denom.

A `proposal.json` example to change the `bypassMinFeeMsgTypes` in globalfee params:

```
{
  "title": "Globalfee Param Change",
  "description": "Update globalfee Params",
  "changes": [
    {
      "subspace": "Globalfee",
      "key": "BypassMinFeeMsgTypes",
      "value": ["/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward", "/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient"]
    }
  ],
  "deposit": "1000000uatom"
}
```
A `proposal.json` example to change the `maxTotalBypassMinFeeMsgGasUsage` in globalfee params:
```
{
  "title": "Globalfee Param Change",
  "description": "Update globalfee Params",
  "changes": [
    {
      "subspace": "globalfee",
      "key": "MaxTotalBypassMinFeeMsgGasUsage",
      "value": 5000
    }
  ],
  "deposit": "1000000uatom"
}
```


## Examples

Here are a few examples to clarify the relationship between global fees, minimum-gas-prices and transaction fees.

**Note:** Transactions can include zero-coin fees. However, these fees are removed from the transaction fees during the fee [parsing](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/client/tx/factory.go#L144) / [sanitizing](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/types/dec_coin.go#L172) before reaching the fee AnteHandler. 
This means `paidfee = "1uatom, 0stake"` and `paidfee = "1uatom"` are equivalent, and similarly, `paidfee = "0uatom"` is equivalent to `paidfee = ""`. 
In the following examples, zero-coin fees are removed from the transaction fees, globalfee refers to `MinimumGasPricesParam` in globalfee params, minimum-gas-prices refers to the local  `minimum-gas-prices` setup in `app.toml`.

### Case 1

**Setting:** globalfee=[], minimum-gas-prices=0.1uatom, gas=2000000. 

Note that this is the same case as globalfee=0uatom, minimum-gas-prices=0.1uatom, gas=2000000.

  - paidfee = "2000000 * 0.1uatom", `pass`
  - paidfee = "2000000 * 0.1uatom, 1stake", `fail` (unexpected denom)
  - paidfee = "", `fail` (insufficient funds)

### Case 2

**Setting:** globalfee=[], minimum-gas-prices="", gas=2000000.

Note that this is the same case as globalfee=0uatom, minimum-gas-prices="", gas=2000000.

  - paidfee = "", `pass`
  - paidfee = "2000000 * 0.1uatom", `pass`
  - paidfee = "2000000 * 0.1stake", `fail` (unexpected denom)
  
### Case 3

**Setting:** globalfee=[0.2uatom], minimum-gas-prices=0.1uatom, gas=2000000 (global fee is higher than min_as_price).

Note that this is the same case as globalfee=0.2uatom, minimum-gas-prices="", gas=2000000.

  - paidfee = "2000000 * 0.2uatom", `pass`
  - paidfee = "2000000 * 0.1uatom", `fail` (insufficient funds)
  - paidfee = "2000000 * 0.2uatom, 1stake", `fail` (unexpected denom)
  - paidfee = "2000000 * 0.2stake", `fail` (unexpected denom)
  - paidfee = "", `fail` (insufficient funds)
  
### Case 4

**Setting:** globalfee=[0.1uatom], minimum-gas-prices=0.2uatom, gas=2000000 (global fee is lower than min_as_price).

Note that the required amount in globalfee is overwritten by the amount in minimum-gas-prices. 

  - paidfee = "2000000 * 0.2uatom", `pass`
  - paidfee = "2000000 * 0.1uatom", `fail` (insufficient funds)
  - paidfee = "2000000 * 0.2uatom, 1stake", `fail` (unexpected denom)
  - paidfee = "2000000 * 0.2stake", `fail` (unexpected denom)
  - paidfee = "", `fail` (insufficient funds)
  - paidfee = 0uatom, `fail` (insufficient funds)
  
### Case 5

**Setting:** globalfee=[0uatom, 1stake], minimum-gas-prices="", gas=200000.

  - paidfee ="2000000 * 0.5stake", `fail` (insufficient funds)
  - paidfee ="", `pass`
  - paidfee ="2000000 * 1uatom, 0.5stake", `pass`
  - paidfee ="2000000 * 1stake", `pass`

### Case 6

**Setting:** globalfee=[0.1uatom, 1stake], minimum-gas-prices=0.2uatom, gas=200000.

Note that the required amount of `uatom` in globalfee is overwritten by the amount in minimum-gas-prices. 

  - paidfee = "2000000 * 0.2uatom", `pass`
  - paidfee = "2000000 * 0.1uatom", `fail` (insufficient funds)
  - paidfee = "2000000 * 1stake", `pass`
  - paidfee = "2000000 * 0.5stake", `fail` (insufficient funds)
  - paidfee = "2000000 * 0.1uatom, 2000000 * 1stake", `pass`
  - paidfee = "2000000 * 0.2atom, 2000000 * 0.5stake", `pass`
  - paidfee = "2000000 * 0.1uatom, 2000000 * 0.5stake", `fail` (insufficient funds)
  
### Case 7

**Setting:** globalfee=[0.1uatom], minimum-gas-prices=[0.2uatom, 1stake], gas=600,000,\
max-total-bypass-min-fee-msg-gas-usage=1,000,000,\
bypass-min-fee-msg-types = [\
"/ibc.core.channel.v1.MsgRecvPacket",\
"/ibc.core.channel.v1.MsgAcknowledgement",\
"/ibc.core.client.v1.MsgUpdateClient",\
"/ibc.core.channel.v1.MsgTimeout",\
"/ibc.core.channel.v1.MsgTimeoutOnClose"\
]

Note that the required amount of `uatom` in globalfee is overwritten by the amount in minimum-gas-prices. 
Also, the `1stake` in minimum-gas-prices is ignored.

  - msgs=["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient"] with paidfee="", `pass`
  - msgs=["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient"] with with paidfee="600000 * 0.05uatom", `pass`
  - msgs= ["/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"] with paidfee="", `fail`
  - msgs=["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient", "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward] with paidfee="", `fail` (transaction contains non-bypass messages)
  - msgs=["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient", "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward] with paidfee="600000 * 0.2uatom", `pass`
  - msgs=["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient"] with paidfee="600000 * 1stake", `fail` (unexpected denom)

### Case 8

**Setting:** globalfee=[1uatom], minimum-gas-prices="0uatom", gas=1,100,000 or 200,\
max-total-bypass-min-fee-msg-gas-usage=1,000,000,\
bypass-min-fee-msg-types = [\
"/ibc.core.channel.v1.MsgRecvPacket",\
"/ibc.core.channel.v1.MsgAcknowledgement",\
"/ibc.core.client.v1.MsgUpdateClient",\
"/ibc.core.channel.v1.MsgTimeout",\
"/ibc.core.channel.v1.MsgTimeoutOnClose"\
]
 - msgs=["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient"] with paidfee="" and gas=1,100,000, `fail` (gas limit exceeded for bypass transactions)
 - msgs=["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient"] with paidfee="200 * 1uatom" and gas=200, `fail` (insufficient funds)
 - msgs=["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.client.v1.MsgUpdateClient"] with paidfee="1,100,000 * 1uatom", `pass` 

## References

- [Gas and Fees in Cosmos SDK](https://docs.cosmos.network/v0.45/basics/gas-fees.html)
