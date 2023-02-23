# Gaia Fees and Fees Checks

## Fee Parameters
The CosmosHub allows managing fees using 3 parameters:

1. setting global fees (`MinimumGasPricesParam`)
Global fees are defined at the network level by setting `MinimumGasPricesParam`, via [Gov Proposals](https://hub.cosmos.network/main/governance/proposals/)

2. `minimum-gas-prices`
This parameter is part of the node configuration, it can be set in the `config/app.toml` configuration file.

3. `bypass-min-fee-msg-types`
This parameter is part of the node configuration, it can be set in the `config/app.toml` configuration file.
This represents a list of message types that will be excluded from paying any fees for inclusion in a block.

Both global fees (`MinimumGasPricesParam`) and `minimum-gas-prices` represent a list of coins, each denoted by an amount and domination as defined by [sdk.DecCoins](https://github.com/cosmos/cosmos-sdk/blob/82ce891aa67f635f3b324b7a52386d5405c5abd0/types/dec_coin.go#L158) 


## Concepts

## Global Fees

Global fees consist of a list of `sdk.DecCoins` e.g., `[1uatom, 2stake]`. 
Every transaction must pay per unit of gas **at least** one of the amounts stated in this list in the corresponding denomination (denom). By this notion, global fees allow a network to impose a minimum transaction fee.

The paid fees must be paid in at least one denom from the global fees list and the corresponding amount per unit of gas must be greater than or equal to the corresponding amount in the global fees list.

A global fees list must meet the following properties:
- fees have to be alphabetically sorted by denom; 
- fees must have non-negative amount, with a valid and unique denom (i.e. no duplicate denoms are allowed).


There are **two exceptions** from the global fees rules that allow zero fee transactions:

1. Transactions that contain only [message types that can bypass the minimum fee](#bypass-fees-message-types) may have zero fees. We refer to this as _bypass transactions_. Node operators can choose to define these message types (for each node) via the `bypass-fee-message-types` configuration parameter.

2. One of the entries in the global fees list has a zero amount, e.g., `0uatom`, and the corresponding denom, e.g., `uatom`, is not present in `minimum-gas-prices`.

Additionally, node operators may set additional minimum gas prices which can be larger than the _global_ minimum gas prices defined on chain.


### minimum-gas-prices

The `minimum-gas-prices` config parameter allows node operators to impose additional requirements for minimum fees. The following rules apply:

- The denoms in `min-gas-prices` that are not present in the global fees list are ignored. 
- The amounts in `min-gas-prices` are considered only if they are greater than the amounts for the corresponding denoms in the global fees list. 

## Bypass Fees Message Types

Bypass messages are messages that are exempt from paying fees. The above global fees and `minimum-gas-prices` checks do not apply for transactions that satisfy the following conditions: 

- Contains only bypass message types, i.e., bypass transactions.
- The total gas used is less than or equal to `MaxTotalBypassMinFeeMsgGasUsage`. Note: the current `MaxTotalBypassMinFeeMsgGasUsage` is set to `1,000,000`.
- In case of non-zero transaction fees, the denom has to be a subset of denoms defined in the global fees list.

Node operators can configure `bypass-min-fee-msg-types` in `config/app.toml`.

- Nodes created using Gaiad `v7.0.2` or later use `["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement","/ibc.applications.transfer.v1.MsgTransfer"]` as defaults. 
- Nodes created using Gaiad `v9.0.1` or later use `["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement","/ibc.applications.transfer.v1.MsgTransfer", "/ibc.core.channel.v1.MsgTimeout", "/ibc.core.channel.v1.MsgTimeoutOnClose"]` as defaults. 
- Node Nodes with `bypass-min-fee-msg-types = []` or missing this field in `app.toml` also use default bypass message types.
- Nodes created using Gaiad `v7.0.1` and `v7.0.0` do not have `bypass-min-fee-msg-types` configured in `config/app.toml` - they are also using same default values as in `v7.0.2`. The `bypass-min-fee-msg-types` config option can be added to `config/app.toml` before the `[telemetry]` field.

An example of `bypass-min-fee-msg-types` in `app.toml`:

```shell

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


## Fee AnteHandler Behaviour

The denoms in the global fees list and the `minimum-gas-prices` param are merged and de-duplicated while keeping the higher amounts. Denoms that are only in the `minimum-gas-prices` param are discarded. 

If the denoms of the transaction fees are a subset of the merged fees and at least one of the amounts of the transaction fees is greater than or equal to the corresponding required fees amount, the transaction can pass the fee check, otherwise an error will occur.

## Queries

CLI queries can be used to retrieve the global fee value:

```shell
gaiad q globalfee minimum-gas-prices
# or
gaiad q params subspace globalfee MinimumGasPricesParam
```

If the global fee is not set, the query returns an empty global fees list: `minimum_gas_prices: []`. In this case the Cosmos Hub will use `0uatom` as global fee in this case (the default fee denom).

## Setting Up Global Fees via Gov Proposals

An example of setting up a global fee by a gov proposals is shown below.
  
```shell
gov submit-proposal param-change proposal.json
````

A `proposal.json` example:

```json
{
  "title": "Global fees Param Change",
  "description": "Update global fees",
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


## Examples

Here are a few examples to clarify the relationship between global fees, minimum-gas-prices and transaction fees.

**Note:** Transactions can include zero-coin fees. However, these fees are removed from the transaction fees during the fee [parsing](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/client/tx/factory.go#L144) / [santitizing](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/types/dec_coin.go#L172) before reaching the fee AnteHandler. 
This means `paidfee = "1uatom, 0stake"` and `paidfee = "1uatom"` are equivalent, and similarly, `paidfee = "0uatom"` is equivalent to `paidfee = ""`. 
In the following examples, zero-coin fees are removed from the transaction fees.

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

**Setting:** globalfee=[0.1uatom], minimum-gas-prices=[0.2uatom, 1stake], gas=200000, bypass-min-fee-msg-types = ["/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"]

Note that the required amount of `uatom` in globalfee is overwritten by the amount in minimum-gas-prices.
Also, the `1stake` in minimum-gas-prices is ignored.

  - msg withdraw-all-rewards with paidfee="", `pass`
  - msg withdraw-all-rewards with paidfee="200000 * 0.05uatom", `pass`
  - msg withdraw-all-rewards with paidfee="200000 * 1stake", `fail` (unexpected denom)

### Case 8

**Setting:** globalfee=[1uatom], minimum-gas-prices="", gas=300000, bypass-min-fee-msg-types = ["/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"]

  - msg withdraw-all-rewards with paidfee="", `fail` (gas limit exceeded for bypass transactions)
  - msg withdraw-all-rewards with paidfee="300000 * 0.5uatom", `fail` (gas limit exceeded for bypass transactions, insufficient funds)
  - msg withdraw-all-rewards with paidfee="300000 * 1uatom", `pass` 

## References

- [Gas and Fees in Cosmos SDK](https://docs.cosmos.network/main/basics/gas-fees.html)
