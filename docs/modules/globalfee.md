# Gaia Fees and Fees Checks

## Fee Parameters
The CosmosHub allows managing fees using 3 parameters:

1. setting global fees (`MinimumGasPricesParam`)
Global fees are defined at the network level by setting `MinimumGasPricesParam`, via [Gov Proposals](https://hub.cosmos.network/main/governance/proposals/)

2. `minimum-gas-prices`
This parameter is part of the node configuration, it can be set in the `config/app.toml` configuration file.

3. `bypass-min-fee-msg-types`
This parameter is part of the node configuration, it can be set in the `config/app.toml` configuration file.
This represent a list of message types that will be excluded from paying any fees for inclusion in a block.

Both global fees (`MinimumGasPricesParam`) and `minimum-gas-prices` represent a list of coins, each denoted by an amount and domination as defined by [sdk.DecCoins](https://github.com/cosmos/cosmos-sdk/blob/82ce891aa67f635f3b324b7a52386d5405c5abd0/types/dec_coin.go#L158) 


## Concepts

## Global Fees

Global fees consist of a list of `sdk.DecCoins` e.g., `[1uatom, 2stake]`. 
Every transaction must pay per unit of gas **at least** one of the amounts stated in this list in the corresponding denomination (denom). By this notion, global fees allow a network to impose a minimum transaction fee.

The paid fees must be paid in at least one denom from the global fees list and the corresponding amount per unit of gas must be greater than or equal to the corresponding amount in the global fees list.

A global fee list must meet the following properties:
- fees have to be alphabetically sorted by denom; 
- fees must have non-negative amount, with a valid and unique denom (i.e. no duplicate denoms are allowed).


There are **two exceptions** from the global fee rules that allow zero fee transactions:

1. transactions that contain only [message types that can bypass the minimum fee](#bypass-fees-message-types) may have zero fees; we refer to this as _bypass transactions_;

2. if one of the entries in the global fees list has a zero amount, e.g., `0uatom`, and the corresponding denom, e.g., `uatom`, is not present in `minimum-gas-prices`.

Some message types can be excluded from paying any fees and therefore are allowed to **bypass the global fee mechanism**.
Node operators can choose to define an exclusion list for each node via the [bypass-fee-message-types](###Bypass Fees Message Types) configuration parameter.

Additionally, node operators may set additional minimum gas prices which can be larger than the minimum gas prices defined on chain.


### minimum-gas-prices

The `minimum-gas-prices` config parameter allows node operators to impose additional requirements for minimum fees.

Amounts in `min-gas-prices` cannot be configured to be less than the global fee amount.
Also, denoms not present in global fees are ignored, meaning that a node cannot be configured to accept denoms not listed in global fees.


## Bypass Fees Message Types

Bypass messages are messages that are exempt from paying fees. The above global fee and `minimum-gas-prices` checks do not apply to bypass message types if these two conditions are met:
- the transaction list contains only bypass message types (bypass messages and non-bypass messages cannot be combined in a single transaction list)
- the total gas of used is less than or equal to `len(messages) * MaxBypassMinFeeMsgGasUsage` (Please note: the current `MaxBypassMinFeeMsgGasUsage` is set to 200,000).
- the transaction fee denom has to be a subset of denoms defined by the global fee

Node operators can configure `bypass-min-fee-msg-types` in `config/app.toml`.

Nodes inited by Gaiad `v7.0.2` or later use `["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement","/ibc.applications.transfer.v1.MsgTransfer"]` as defaults. Nodes with `bypass-min-fee-msg-types = []` or missing this field in `app.toml` also use default bypass message types.

Nodes created using Gaiad `v7.0.1` or earlier do not have `bypass-min-fee-msg-types` configured in `config/app.toml` - they are alsousing default values. The `bypass-min-fee-msg-types` config option can be added to `config/app.toml` before the `[telemetry]` field.

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
bypass-min-fee-msg-types = ["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement","/ibc.applications.transfer.v1.MsgTransfer"]
```

Since each node can be configured with `minimum-gas-prices` and `bypass-min-fee-msg-types` all transactions must satisfy the configured gas fee requirements in order to be processed.


## Fee AnteHandler Behaviour

The denoms in the global fees list and the minimum-gas-prices param are merged and de-duplicated while keeping the higher amounts. Denoms that are only in the `minimum-gas-prices` param are discarded. 

If the paid fee is a subset of the combined fees set and the paid fee amount is greater than or equal to the required fees amount, the transaction can pass the fee check, otherwise an error will occur.


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

Here are a few examples to clarify the relationship between global fees, minimum-gas-prices and paid fees.

*Please note that transactions can include zero coins as paid fees. For example, when adding zero coins as fees in a transaction through the CLI, they will be removed from the fees during the fee [parsing](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/client/tx/factory.go#L144)/[santitizing](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/types/dec_coin.go#L172) before reaching the fee handler. This means `paidfee = "1uatom, 0stake"` and `paidfee = "1uatom"` are equivalent, and similarly, `paidfee = "0uatom"` is equivalent to `paidfee = ""`. In the following examples, zero coins are removed from paidfees for simplicity.*

**Case 1:**

globalfee=[], minimum-gas-prices=0.0001uatom, gas=2000000
  This is the same case as globalfee=0uatom, minimum-gas-prices=0.0001uatom, gas=2000000.
  - paidfee = "2000000 * 0.0001uatom", `pass`
  - paidfee = "2000000 * 0.0001uatom, 1stake", `fail`
  - paidfee = "2000000 * 0.0001/2uatom", `fail`
  - paidfee = "", `fail`

**Case 2:**

globalfee=[], minimum-gas-prices="", gas=2000000 (When globalfee empty, the [default globalfee of 0uatom](https://github.com/cosmos/gaia/blob/d6d2933ede1aa1a13040f5aee2f0f7b795c168d0/x/globalfee/ante/fee.go#L135) will be used.)
  - paidfee = "", `pass`
  - paidfee = "2000000 * 0.0001uatom", `pass`
  - paidfee = "2000000 * 0.0001stake", `fail`
  
**Case 3**

 globalfee=0.0002uatom, minimum-gas-prices=0.0001uatom, gas=2000000 (global fee is lower than min_as_price)
  - paidfee = "2000000 * 0.0002uatom", `pass`
  - paidfee = "2000000 * 0.0001uatom", `fail`
  - paidfee = "2000000 * 0.0002uatom, 1stake", `fail`
  - paidfee = "2000000 * 0.0002stake", `fail`
  - paidfee = "", `fail`
  
**Case 4**

  globalfee=0.0001uatom, minimum-gas-prices=0.0002uatom, gas=2000000 (global fee is higher than min_as_price)
  - paidfee = "2000000 * 0.0002uatom", `pass`
  - paidfee = "2000000 * 0.0001uatom", `fail`
  - paidfee = "2000000 * 0.0002uatom, 1stake", `fail`
  - paidfee = "2000000 * 0.0002uatom", `pass`
  - paidfee = "2000000 * 0.0002stake", `fail`
  - paidfee = "", `fail`
  - paidfee = 0uatom, `fail`
  
**Case 5**

 globalfee=[0uatom, 1stake], minimum-gas-prices="", gas=200000.
  - paidfee ="2000000 * 0.5stake", `fail`
  - paidfee ="", `pass`
  - paidfee ="2000000 * 1uatom, 0.5stake", `pass`
  - paidfee ="2000000 * 1stake", `pass`

**Case 6**

 globalfee=[0.001uatom, 1stake], minimum-gas-prices=0.002uatom, gas=200000.
  - paidfee = "2000000 * 0.0002uatom", `pass`
  - paidfee = "2000000 * 0.0001uatom", `fail`
  - paidfee = "2000000 * 1stake", `pass`
  - paidfee = "2000000 * 1/2stake", `fail`
  - paidfee = "2000000 * 0.0001uatom, 2000000 * 1stake", `pass`
  - paidfee = "2000000 * 0.0002atom, 2000000 * 1/2stake", `pass`
  - paidfee = "2000000 * 0.0001uatom, 2000000 * 1/2stake", `fail`
  
**Case 7**

globalfee=[0.0001uatom], minimum-gas-prices=0.0002uatom,1stake, gas=200000.

   `bypass-min-fee-msg-types = ["/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"]`
  - msg withdraw-all-rewards with paidfee="", `pass`
  - msg withdraw-all-rewards with paidfee="200000 * 0.0001/2uatom", `pass`
  - msg withdraw-all-rewards with paidfee="200000 * 1stake", `fail`

## References

- [Fee caculations: fee and gas](https://docs.cosmos.network/main/basics/gas-fees.html)
