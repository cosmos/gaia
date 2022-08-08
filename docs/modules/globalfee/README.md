# Gaia Fee Ante-handler

## Gaia Fees

The CosmosHub has two types of fees, both of which, are the [sdk.DecCoins]((https://github.com/cosmos/cosmos-sdk/blob/a1777a87b65fad74732cfe1a4c27683dcffffbfe/types/dec_coin.go#L158)) type:
- global fees (are defined at the network level, via [Gov Proposals](https://hub.cosmos.network/main/governance/proposals/))
- min_gas_prices (are specified validator node, in the `config/app.toml` configuration file)
  
### Global Fees
Global fees are set up through governance proposal which must be voted on by validators. 

For a  [global fee](https://github.com/cosmos/gaia/blob/82c4353ab1b04cf656a8c95d226c30c7845f157b/x/globalfee/types/params.go#L54-L99) to be valid:
- fees have to be alphabetically sorted by denomination (denom)
- fees have to have non-negative amount, with a valid and unique denom (i.e no duplicates). 

Global fees allow denoms with zero coins or value.

Zero value coins are used to define fee denoms, when the chain does not charge fees. Each transaction (except bypass transactions) have to meet the following global fee requirements:
- All denoms of the paid fees (except zero coins) have to be a subset of the global fee's denom set.
- All paidfees' contain at least one denom that is present and greater than/or equal to the amount of the same denom in globalfee.

### Query global fees
CLI queries to retrieve the global fee value:
```shell
gaiad q globalfee minimum-gas-prices
# or
gaiad q params subspace globalfee MinimumGasPricesParam
```
### Empty global fees and default global fees
When the global fee is not setup, the query will return an empty globalfee list: `minimum_gas_prices: []`. Gaiad will use `0uatom` as global fee in this case. This is due to the CosmosHub accepting uatom as fee denom by default.

### Setting up global fees via Gov Proposals
An example of setting up a global fee by a gov proposals is shown below.
  
```shell
gov submit-legacy-proposal param-change proposal.json
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
Please note: in the above "value" field, coins must sorted alphabetically by denom.

### Parameter: min_gas_prices
min_gas_prices are a node's further requirement of fees. Zero coins are removed from min_gas_prices when [parsing min_gas_prices](https://github.com/cosmos/cosmos-sdk/blob/3a097012b59413641ac92f18f226c5d6b674ae42/baseapp/options.go#L27), this is different from global fees.
- If the `min_gas_prices` set a denom that is not global fees's denom set. This min_gas_prices denom will not be considered when paying fees.
- If the `min_gas_prices` have a denom in global fees's denom set, and the  min_gas_prices are lower than global fees, the fee still need to meet the global fees.
- If the `min_gas_prices` have a denom in global fees's denom set, and the  min_gas_prices are higher than global fees, the fee need to meet the min_gas_prices.

### Fee Checks
Global fees, min_gas_prices and the paid fees all allow zero coins setup. After parsing the fee coins, zero coins and paid fees will be removed from the min_gas_prices and paid fees. 

Only global fees might contain zero coins, which is used to define the allowed denoms of paid fees.

The [Fee AnteHandle](https://github.com/cosmos/gaia/blob/yaru/fix-all-fees/ante/fee.go) will take global fees and min_gas_prices and merge them into one [combined `sdk.Deccoins`](https://github.com/cosmos/gaia/blob/f2be720353a969b6362feff369218eb9056a60b9/ante/fee.go#L79) according to the denoms and amounts of global fees and min_gas_prices.

If the paid fee is a subset of the combined fees set and the paid fee amount is greater than or equal to the required fees amount, the transaction can pass the fee check, otherwise an error will occur.

### Bypass Fees Message Types
The above `global fee` and `min_as_prices` fee checks do not apply to bypass message types. Transactions of  bypass message types are free of fee charge. However, if the bypass type transactions still carry nonzero fees, the denom has to be a subset of denoms that global fees defined.

A node can set up its own bypass message types by adding the configuration parameter `bypass-min-fee-msg-types` in `config/app.toml` file.

An example:
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

[comment]: <> (Even though each node can set its own `min_gas_prices` and `bypass-min-fee-msg-types`, when the transactions enters a validator's mempool, the transactions carried fees have to satisfy the validator's `min_gas_prices` and `bypass-min-fee-msg-types`'s requirement in order for the validators to process the transacton.)

### Examples
Here are a few examples to clarify the relationship between global fees, min_gas_prices and paid fees.
- Case 1: globalfee=[], min_gas_prices=0.0001uatom, gas=2000000

  This is the same case as globalfee=0uatom, min_gas_prices=0.0001uatom, gas=2000000.
  - paidfee = "2000000 * 0.0001uatom", pass
  - paidfee = "2000000 * 0.0001uatom, 0stake", pass
  - paidfee = "2000000 * 0.0001uatom, 1stake", fail 
  - paidfee = "2000000 * 0.0001/2uatom", fail
  - paidfee = "", pass


- Case 2: globalfee=[], min_gas_prices="", gas=2000000
  - paidfee = "", pass
  - paidfee = "0uatom", pass
  - paidfee = "1uatom", pass
  - paidfee = "0uatom, 0stake", pass
  - paidfee = "0uatom, 1stake", fail
  

- Case 3: globalfee=0.0002uatom, min_gas_prices=0.0001uatom, gas=2000000 (global fee is lower than min_as_price)
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 0.0002uatom, 1stake", fail
  - paidfee = "2000000 * 0.0002uatom, 0stake", pass
  - paidfee = "2000000 * 0.0002stake", fail
  - paidfee = "", fail
  - paidfee = 0uatom, fail
  

- Case 4:  globalfee=0.0001uatom, min_gas_prices=0.0002uatom, gas=2000000 (global fee is higher than min_as_price)
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 0.0002uatom, 1stake", fail
  - paidfee = "2000000 * 0.0002uatom, 0stake", pass
  - paidfee = "2000000 * 0.0002stake", fail
  - paidfee = "", fail
  - paidfee = 0uatom, fail
  

- Case 5: globalfee=[0uatom, 1stake], min_gas_prices="", gas=200000.
 - fees="2000000 * 0uatom, 2000000 * 0.5stake", fail, (this is due to [fee parsing, zero coins are removed](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/client/tx/factory.go#L144), it equals to paidfees = 0.5stake in this case)
 - paidfees="", pass
 - paidfees="2000000 * 1uatom, 0.5stake", pass
 - paidfees="0uatom, 0stake" pass, (due to the parsing of paidfees, it makes paidfees="")
 - paidfees="2000000 * 1stake", pass


- Case 6: globalfee=[0.001uatom, 1stake], min_gas_prices=0.002uatom, gas=200000.
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 1stake", pass
  - paidfee = "2000000 * 1/2stake", fail
  - paidfee = "2000000 * 0.0001uatom, 2000000*1stake", pass
  - paidfee = "2000000 * 0.0002atom, 2000000*1/2stake", pass
  - paidfee = "2000000 * 0.0001uatom, 2000000*1/2stake", fail
  

- Case 7:  globalfee=[0.0001uatom], min_gas_prices=0.0002uatom,1stake, gas=200000.
  `bypass-min-fee-msg-types = ["/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward"]`
   - msg withdraw-all-rewards with paidfee=0uatom, pass
  - msg withdraw-all-rewards with paidfee=200000 * 0.0001/2uatom, pass
  - msg withdraw-all-rewards with paidfee=0stake,0photon, pass
  - msg withdraw-all-rewards with paidfee=200000 * 1stake, 0photon, fail

### Reference
- [Fee caculations: fee and gas](https://docs.cosmos.network/main/basics/gas-fees.html)
