# Gaia Fees and Fees Check

## Gaia Fees

The CosmosHub has two types of fees, both of which defined as [sdk.DecCoins](https://github.com/cosmos/cosmos-sdk/blob/82ce891aa67f635f3b324b7a52386d5405c5abd0/types/dec_coin.go#L158) 
(i.e., a list of coins, each denoted by an amount and a denomination):

- global fees (defined at the network level, via [Gov Proposals](https://hub.cosmos.network/main/governance/proposals/));
- `minimum-gas-prices` (specified by validator nodes, in the `config/app.toml` configuration file).
  
### Global Fees

#### Global Fees Concept

Global fees consist of a list of `sdk.DecCoins` e.g., `[1uatom, 2stake]`. 
Every transaction must pay per unit of gas **at least** one of the amounts stated in this list in the corresponding denomination (denom). 
There are two exceptions to this rule:

- first, transactions that contain only [message types that can bypass the minimum fee](#bypass-fees-message-types) may have zero fees; we refer to this as _bypass transactions_;
- second, if one of the entries in the global fees list has a zero amount, e.g., `0uatom`, and the corresponding denom, e.g., `uatom`, is not present in `minimum-gas-prices`.

Global fees are set up through a governance proposal which must be voted on by validators.

For the [global fees](https://github.com/cosmos/gaia/blob/82c4353ab1b04cf656a8c95d226c30c7845f157b/x/globalfee/types/params.go#L54-L99) to be valid:

- fees have to be alphabetically sorted by denom; 
- fees must have non-negative amount, with a valid and unique denom (i.e. no duplicate denoms are allowed).

Global fees allow zero value coins that are used to define accepted denoms without imposing a minimum requirement on the amount. 
For example, `[0uatom]` means that transactions with fees in other denoms than `uatom` will be rejected.

Every transaction (except bypass transactions) have to meet the following global fee requirements:

- All the denoms of the paid fees have to be a subset of the denoms from the global fees list.
- The paid fees contain at least one denom from the global fees list and the corresponding amount per unit of gas is greater than or equal to the corresponding amount in the global fees list.

#### Query Global Fees

CLI queries to retrieve the global fee value:

```shell
gaiad q params subspace globalfee MinimumGasPricesParam
```

#### Empty Global Fees and Default Global Fees

When the global fee is not set, the query will return an empty global fees list: `minimum_gas_prices: []`. However, the Cosmos Hub will use `0uatom` as global fee in this case.

#### Setting Up Global Fees via Gov Proposals

An example of setting up a global fee by a gov proposals is shown below.
  
```shell
gov submit-proposal param-change proposal.json
````

A `proposal.json` example:

```
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

### minimum-gas-prices

The `minimum-gas-prices` config parameter is a node's further requirement for minimum fees. Unlike the global fees list, the zero coins are removed when [parsing minimum-gas-prices](https://github.com/cosmos/cosmos-sdk/blob/3a097012b59413641ac92f18f226c5d6b674ae42/baseapp/options.go#L27).

- If the minimum-gas-prices set a denom that is not global fees's denom set. This minimum-gas-prices denom will not be considered when paying fees.
- If the minimum-gas-prices have a denom in global fees's denom set, and the  minimum-gas-prices are lower than global fees, the fee still need to meet the global fees.
- If the minimum-gas-prices have a denom in global fees's denom set, and the  minimum-gas-prices are higher than global fees, the fee need to meet the minimum-gas-prices.

## Fee AnteHandler

The denoms in the global fees list and the minimum-gas-prices param are merge and de-duplicated while keeping the higher amounts. Denoms that are only in the minimum-gas-prices param are discarded. 

If the paid fee is a subset of the combined fees set and the paid fee amount is greater than or equal to the required fees amount, the transaction can pass the fee check, otherwise an error will occur.

### Bypass Fees Message Types

Bypass messages are messages that are exempt from paying fees. The above global fee and min_as_prices fee checks do not apply to bypass message types under two conditions:
- all transaction messages should be bypass message types.
- the total gas of the messages should be less than or equal to `len(messages)*MaxBypassMinFeeMsgGasUsage` (Please note: the current `MaxBypassMinFeeMsgGasUsage` is set to 200,000).

However, if the bypass type transactions satisfy the above condition and still carry nonzero fees, the denom has to be a subset of denoms that global fees defined.

Each node can configure its own desired `bypass-min-fee-msg-types` in `config/app.toml`. Node inited by Gaiad `v7.0.2` or later will get default bypass messages `["/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgAcknowledgement","/ibc.applications.transfer.v1.MsgTransfer"]` in `app.toml`. Node with `bypass-min-fee-msg-types = []` or missing this field in `app.toml` will also use default bypass message types. Node inited by Gaiad `v7.0.1` or earlier might not have `bypass-min-fee-msg-types`, users can insert it before the field `[telemetry]` in `app.toml`.

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

Even though each node can set its own `minimum-gas-prices` and `bypass-min-fee-msg-types`, when the transactions enters a validator's mempool, the transactions carried fees have to satisfy the validator's `minimum-gas-prices` and `bypass-min-fee-msg-types`'s requirement in order for the validators to process the transactions.

## Examples

Here are a few examples to clarify the relationship between global fees, minimum-gas-prices and paid fees.

*Please note that transactions can include zero coins as paid fees. For example, when adding zero coins as fees in a transaction through the CLI, they will be removed from the fees during the fee [parsing](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/client/tx/factory.go#L144)/[santitizing](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/types/dec_coin.go#L172) before reaching the fee handler. This means `paidfee = "1uatom, 0stake"` and `paidfee = "1uatom"` are equivalent, and similarly, `paidfee = "0uatom"` is equivalent to `paidfee = ""`. In the following examples, zero coins are removed from paidfees for simplicity.*

- **Case 1**: globalfee=[], minimum-gas-prices=0.0001uatom, gas=2000000
  This is the same case as globalfee=0uatom, minimum-gas-prices=0.0001uatom, gas=2000000.
  - paidfee = "2000000 * 0.0001uatom", pass
  - paidfee = "2000000 * 0.0001uatom, 1stake", fail
  - paidfee = "2000000 * 0.0001/2uatom", fail
  - paidfee = "", fail

- **Case 2**: globalfee=[], minimum-gas-prices="", gas=2000000 (When globalfee empty, the [default globalfee of 0uatom](https://github.com/cosmos/gaia/blob/d6d2933ede1aa1a13040f5aee2f0f7b795c168d0/x/globalfee/ante/fee.go#L135) will be used.)
  - paidfee = "", pass
  - paidfee = "2000000 * 0.0001uatom", pass
  - paidfee = "2000000 * 0.0001stake", fail
  
- **Case 3**: globalfee=0.0002uatom, minimum-gas-prices=0.0001uatom, gas=2000000 (global fee is lower than min_as_price)
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 0.0002uatom, 1stake", fail
  - paidfee = "2000000 * 0.0002stake", fail
  - paidfee = "", fail
  
- **Case 4**:  globalfee=0.0001uatom, minimum-gas-prices=0.0002uatom, gas=2000000 (global fee is higher than min_as_price)
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 0.0002uatom, 1stake", fail
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0002stake", fail
  - paidfee = "", fail
  - paidfee = 0uatom, fail
  
- **Case 5**: globalfee=[0uatom, 1stake], minimum-gas-prices="", gas=200000.
  - paidfees="2000000 * 0.5stake", fail
  - paidfees="", pass
  - paidfees="2000000 * 1uatom, 0.5stake", pass
  - paidfees="2000000 * 1stake", pass

- **Case 6**: globalfee=[0.001uatom, 1stake], minimum-gas-prices=0.002uatom, gas=200000.
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 1stake", pass
  - paidfee = "2000000 * 1/2stake", fail
  - paidfee = "2000000 * 0.0001uatom, 2000000 * 1stake", pass
  - paidfee = "2000000 * 0.0002atom, 2000000 * 1/2stake", pass
  - paidfee = "2000000 * 0.0001uatom, 2000000 * 1/2stake", fail
  
- **Case 7**:globalfee=[0.0001uatom], minimum-gas-prices=0.0002uatom,1stake, gas=200000.

   `bypass-min-fee-msg-types = ["/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"]`
  - msg withdraw-all-rewards with paidfee="", pass
  - msg withdraw-all-rewards with paidfee="200000 * 0.0001/2uatom", pass
  - msg withdraw-all-rewards with paidfee="200000 * 1stake", fail

## Reference

- [Fee caculations: fee and gas](https://docs.cosmos.network/main/basics/gas-fees.html)
