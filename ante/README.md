# Gaia Fee Ante-handler

## Gaia Fees

In cosmoshub, there are two types of fees:
- global fee
- min_gas_price
  
### Global Fees
Global fees are set up through gov proposal,  global fees are [`sdk.DecCoins`](https://github.com/cosmos/cosmos-sdk/blob/a1777a87b65fad74732cfe1a4c27683dcffffbfe/types/dec_coin.go#L158) type and globally valid. [Valid global fees](https://github.com/cosmos/gaia/blob/82c4353ab1b04cf656a8c95d226c30c7845f157b/x/globalfee/types/params.go#L54-L99) have to be and sorted by denom, have have nonnegtive amount, with a valid and unique denomination(denom) (i.e no duplicates), Global fees allow zero coins! zero coins can help define the desired fee denoms even the chain might not charge the fees. Each transaction (except bypass transaction types) has to meet the one of the fees in global fees in terms of fee amount and fee denom.

### query global fees
You can query globalfee by 
```shell
gaiad q globalfee minimum-gas-prices
// or
gaiad q params subspace globalfee MinimumGasPricesParam
```
### empty global fees  and default global fees
When global fee is not setup, the query will return empty globalfee `minimum_gas_prices: []`. Gaiad will use `0uatom` as global fee in this case. This is due to the Cosmoshub as default only accepts uatom as fee denom.

### setup global fees
Global fee can be setup by the new gov modules.
```shell
gov submit-legacy-proposal param-change proposal.json
````
A proposal.json  example:
please note in the "value" field, coins must sorted by denom.
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

### Min_Gas_Price
Min_gas_price is [`sdk.DecCoins`](https://github.com/cosmos/cosmos-sdk/blob/a1777a87b65fad74732cfe1a4c27683dcffffbfe/types/dec_coin.go#L158) type. Min_gas_price is set up in `config/app.toml` by each node operator. Min_gas_price is a node's further requirement of fees. A valid min_gas_price is ....(one can setup zerocoins, but it will be removed from the min_gas_price), this is different from global fees validation.
- If the min_gas_price set a denom that is not global fees's denom set. This min_gas_price denom will not be considered when paying fees.
- If the min_gas_price is a denom in global fees's denom set, and the  min_gas_price is lower than global fees, the fee still need to meet the global fees.
- If the min_gas_price is a denom in global fees's denom set, and the  min_gas_price is higher than global fees, the fee need to meet the min_gas_price.

### Bypass Fees Message Types
However, the above mentioned global fees and min_as_price does not apply to bypass message types. transactions of  bypass message types are free of fee charge. However, if the bypass type transactions still carry nonzero fees, the denom has to be a subset of denoms that global fees defined.

A node can set up its own bypass massage types in `config/app.toml`.
exmple of `bypass-min-fee-msg-types` in `app.toml`.
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

**Please note:**

Even each node can setup its own `min_gas_price` and `bypass-min-fee-msg-types`, when the transactions entering validators' mempools, the transactions carried fees have to satisfy validators' `min_gas_price` and `bypass-min-fee-msg-types`'s requirement in order for the validators to process the transactons.


### Fee checks (BypassMinFeeDecorator)
...

### Examples
Here are a few examples to clarify the relations between global fees, min_gas_price and paid fees.
- case 1: globalfee=[], min_gas_price=0.0001uatom, gas=2000000

  This is the same case as globalfee=0uatom, min_gas_price=0.0001uatom, gas=2000000.
  - paidfee = "2000000 * 0.0001uatom", pass
  - paidfee = "2000000 * 0.0001uatom, 0stake", pass
  - paidfee = "2000000 * 0.0001uatom, 1stake", fail 
  - paidfee = "2000000 * 0.0001/2uatom", fail
  - paidfee = "", pass

- case 2: globalfee=[], min_gas_price="", gas=2000000
  - paidfee = "", pass
  - paidfee = "0uatom", pass
  - paidfee = "1uatom", pass
  - paidfee = "0uatom, 0stake", pass
  - paidfee = "0uatom, 1stake", fail
  
- case 3: globalfee=0.002uatom, min_gas_price=0.001uatom, gas=2000000 (global fee is lower than min_as_price)
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 0.0002uatom, 1stake", fail
  - paidfee = "2000000 * 0.0002uatom, 0stake", pass
  - paidfee = "2000000 * 0.0002stake", fail
  - paidfee = "", fail
  - paidfee = 0uatom, fail
  
- case 4:  globalfee=0.001uatom, min_gas_price=0.002uatom, gas=2000000 (global fee is higher than min_as_price)
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 0.0002uatom, 1stake", fail
  - paidfee = "2000000 * 0.0002uatom, 0stake", pass
  - paidfee = "2000000 * 0.0002stake", fail
  - paidfee = "", fail
  - paidfee = 0uatom, fail
  
- case 5: globalfee=[0uatom, 1stake], min_gas_price="", gas=200000.
 - fees="2000000 * 0uatom,2000000 * 0.5stake", fail, (this is due to [fee parsing, zero coins are removed](https://github.com/cosmos/cosmos-sdk/blob/e716e4103e934344aa7be6dc9b5c453bdec5f225/client/tx/factory.go#L144), fees is actually 0.5stake in this case)
 - paidfees="", pass
 - paidfees="2000000 * 1uatom, 0.5stake", pass
 - paidfees="0uatom, 0stake" pass, (due to the parsing of paidfees, it makes paidfees="")
 - paidfees="2000000 * 1stake", pass

- case 6: globalfee=[0.001uatom, 1stake], min_gas_price=0.002uatom, gas=200000.
  - paidfee = "2000000 * 0.0002uatom", pass
  - paidfee = "2000000 * 0.0001uatom", fail
  - paidfee = "2000000 * 1stake", pass
  - paidfee = "2000000 * 1/2stake", fail
  - paidfee = "2000000 * 0.0001uatom, 2000000*1stake", pass
  - paidfee = "2000000 * 0.0002atom, 2000000*1/2stake", pass
  - paidfee = "2000000 * 0.0001uatom, 2000000*1/2stake", fail
  
### Tests
The fee antehandler tests and bypass transactions are tested in e2e test.


### Reference

- [Fee caculations: fee and gas](https://docs.cosmos.network/main/basics/gas-fees.html)
