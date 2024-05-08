# `crisis` subspace

The `crisis` module is responsible for halting the Cosmos Hub if an invariant is broken. The crisis module has the following parameters:

<table>
    <tr>
        <th>Key</th>
        <th>Value</th>
    </tr>
    <tr v-for="(v,k) in $themeConfig.currentParameters.crisis">
        <td><a :href="'#'+k"><code>{{ k }}</code></a></td>
        <td><code>{{ v }}</code></td>
    </tr>
</table>

The `crisis` module is responsible for halting the blockchain under the circumstance that a blockchain invariant is broken. Invariants can be registered with the application during the application initialization process.

## Governance notes on parameters

### `ConstantFee`
**The amount required to send a message to halt the Cosmos Hub chain if an invariant is broken, in micro-ATOM.**

A Cosmos account (address) can send a transaction message that will halt the Cosmos Hub chain if an invariant is broken. An example of this would be if all of the account balances in total did not equal the total supply. This kind of transaction could consume excessive amounts of gas to compute, beyond the maximum allowable block gas limit. `ConstantFee` makes it possible to bypass the gas limit in order to process this transaction, while setting a cost to disincentivize using the function to attack the network. The cost of the transaction is `1333000000` `uatom` (1,333 ATOM) and will effectively not be paid if the chain halts due to a broken invariant (which similar to being refunded). If the invariant is not broken, then `ConstantFee` will be paid. All in Bits has published more information about the [crisis module here](https://docs.cosmos.network/master/modules/crisis/).

* on-chain value: `{{ $themeConfig.currentParameters.crisis.ConstantFee }}`
* `cosmoshub-4` default: `1333000000` `uatom`
* `cosmoshub-3` default: `1333000000` `uatom`

#### Decreasing the value of `ConstantFee`
Decreasing the value of the `ConstantFee` parameter will reduce the cost of checking an invariant. This will likely make it easier to halt the chain if an invariant is actually broken, but it will lower the cost for an attacker to use this function to slow block production.

#### Increasing the value of `ConstantFee`
Increasing the value of the `ConstantFee` parameter will increase the cost of checking an invariant. This will likely make it more difficult to halt the chain if an invariant is actually broken, but it will increase the cost for an attacker to use this function to slow block production.

#### Notes
Only [registered invariants](https://github.com/cosmos/cosmos-sdk/blob/master/x/supply/keeper/invariants.go) may be checked with this transaction message. Validators are reportedly performant enough to handle large computations like invariant checks, and the likely outcome of multiple invariant checks would be longer block times. In the code, there is a comment that indicates that the designers were targeting $5000 USD as the required amount of ATOMs to run an invariant check.