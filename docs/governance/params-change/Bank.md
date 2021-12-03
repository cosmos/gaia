# `bank` subspace

The `bank` module is responsible for token transfer functionalities. It has the following parameters:

<table>
    <tr>
        <th>Key</th>
        <th>Value</th>
    </tr>
    <tr v-for="(v,k) in $themeConfig.currentParameters.bank">
        <td><a :href="'#'+k"><code>{{ k }}</code></a></td>
        <td><code>{{ v }}</code></td>
    </tr>
</table>

## Governance notes on parameters
### `SendEnabled`
**Token transfer functionality.**

The Cosmos Hub (cosmoshub-1) launched without transfer functionality enabled. Users were able to stake and earn rewards, but were unable to transfer ATOMs between accounts until the cosmoshub-2 chain launched. Transfer functionality may be disabled and enabled via governance proposal.

* on-chain value: `SendEnabled`: `{{ $themeConfig.currentParameters.bank.SendEnabled }}`
* on-chain value: `DefaultSendEnabled`: `{{ $themeConfig.currentParameters.bank.DefaultSendEnabled }}`
* `cosmoshub-4` added `DefaultSendEnabled`: `true`
* `cosmoshub-3` default: `true`

#### Enabling `sendenabled`
Setting the `sendenabled` parameter to `true` will enable ATOMs to be transferred between accounts. This capability was first enabled when the cosmoshub-2 chain launched.

#### Disabling `sendenabled`
Setting the `sendenabled` parameter to `false` will prevent ATOMs from being transferred between accounts. ATOMs may still be staked and earn rewards. This is how the cosmoshub-1 chain launched.


#### Notes
The cosmoshub-1 chain launched with `sendenabled` set to `false` and with [`withdrawaddrenabled`](./Distribution.md#4-withdrawaddrenabled) set to `false`. Staking was enabled on cosmoshub-1, so setting `withdrawaddrenabled` to false was necessary to prevent a loophole that would enable ATOM transfer via diverting staking rewards to a designated address.
