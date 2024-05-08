# `distribution` subspace

The `distribution` module is responsible for distributing staking rewards between validators, delegators, and the Community Pool. It has the following parameters:

<table>
    <tr>
        <th>Key</th>
        <th>Value</th>
    </tr>
    <tr v-for="(v,k) in $themeConfig.currentParameters.distribution">
        <td><a :href="'#'+k"><code>{{ k }}</code></a></td>
        <td><code>{{ v }}</code></td>
    </tr>
</table>

The `distribution` module enables a simple distribution mechanism that passively distributes rewards between validators and delegators. Collected rewards are pooled globally and divided out passively to validators and delegators. Each validator has the opportunity to charge commission to the delegators on the rewards collected on behalf of the delegators. Fees are collected directly into a global reward pool and validator proposer-reward pool.

**There is [a known bug](#known-bug) associated with this module.** 

## Governance notes on parameters

### `communitytax`
**The proportion of staking rewards diverted to the community pool.**

Staking on the Cosmos Hub entitles participants to inflationary (aka "block") rewards and transaction fees. A portion of these staking rewards is diverted to the community pool, which can be spent with a successful community-spend governance proposal. `communitytax` is the parameter that determines the proportion of staking rewards diverted to the community pool, which is currently `0.020000000000000000` (2%) of all staking rewards.

* on-chain value: `{{ $themeConfig.currentParameters.distribution.communitytax }}`
* `cosmoshub-4` default: `0.020000000000000000`
* `cosmoshub-3` default: `0.020000000000000000`

#### Decreasing the value of `communitytax`
Decreasing the value of the `communitytax` parameter will decrease the rate that the community pool is funded and will increase the staking rewards captured by staking participants. This will make it more likely for the community pool to be exhausted and could potentially increase the motivation for participants to stake.

#### Increasing the value of `communitytax`
Increasing the value of the `communitytax` parameter will increase the rate that the community pool is funded and will decrease the staking rewards captured by staking participants. This will make it more less for the community pool to be exhausted and could potentially decrease the motivation for participants to stake.

### `baseproposerreward`
**The fixed base reward bonus for the validator proposing a block, as a proportion of transaction fees.**

All validators in the active set share the rewards for producing a block equally, except for the proposer of a valid block: that validator receives a bonus of `0.010000000000000000` (1%) more in transaction fees. The proposer must include a minimum of 2/3 of precommit signatures from the other validators in the active set in order for the block to be valid and to receive the `baseproposerreward` bonus. All in Bits has published more in-depth information [here](../../validators/validator-faq.html#how-are-fees-distributed).

* on-chain value:  `{{ $themeConfig.currentParameters.distribution.baseproposerreward }}`
* `cosmoshub-4` default: `0.010000000000000000`
* `cosmoshub-3` default: `0.010000000000000000`

#### Decreasing the value of `baseproposerreward`
Decreasing the value of the `baseproposerreward` parameter will decrease the advantage that the proposer has over other validators. This may decrease an operator's motivation to ensure that its validator is reliably online and includes at least 2/3 precommit signatures of the other validators in its proposed block.

#### Increasing the value of `baseproposerreward`
Increasing the value of the `baseproposerreward` parameter will increase the advantage that the proposer has over other validators. This may increase an operator's motivation to ensure that its validator is reliably online and includes at least 2/3 precommit signatures of the other validators in its proposed block.

#### Notes
The Cosmos Hub transaction fee volume is proportionally very low in value compared to the inflationary block rewards, and until that changes, this parameter will likely have very little impact on validator behaviours. As fee volumes increase, the `baseproposerreward` bonus may incentivize delegations to the validator(s) with the greatest stake-backing. There are some detailed discussions about the proposer bonus [here](https://github.com/cosmos/cosmos-sdk/issues/3529).

###  `bonusproposerreward`
**The maximum additional reward bonus for the validator proposing a block, as a proportion of transaction fees.**

All validators in the active set share the rewards for producing a block equally, except for the proposer of a valid block. If that validator includes more than a minimum of 2/3 of precommit signatures from the other validators in the active set, they are eligible to receive the `bonusproposerreward` of up to 4% (`0.040000000000000000`), beyond the 1% `baseproposerreward`. The bonus proposer reward amount that a validator receives depends upon how many precommit signatures are included in the proposed block (additional to the requisite 2/3). All in Bits has published more in-depth information [here](../../validators/validator-faq.html#how-are-fees-distributed).

* on-chain value: `{{ $themeConfig.currentParameters.distribution.bonusproposerreward }}`
* `cosmoshub-4` default: `0.040000000000000000`
* `cosmoshub-3` default: `0.040000000000000000`

#### Decreasing the value of `bonusproposerreward`
Decreasing the value of the `bonusproposerreward` parameter will decrease the advantage that the proposer has over other validators. This may decrease an operator's motivation to ensure that its validator is reliably online and includes more than 2/3 precommit signatures from the other validators in its proposed block.

#### Increasing the value of `bonusproposerreward`
Increasing the value of the `bonusproposerreward` parameter will increase the advantage that the proposer has over other validators. This may increase an operator's motivation to ensure that its validator is reliably online and includes more than 2/3 precommit signatures from the other validators in its proposed block. 

#### Notes
The Cosmos Hub transaction fee volume is proportionally very low in value compared to the inflationary block rewards, and until that changes, this parameter will likely have very little impact on validator behaviours. As fee volumes increase, the `bonusproposerreward` bonus may incentivize delegations to the validator(s) with the greatest stake-backing. There are some detailed discussions about the proposer bonus [here](https://github.com/cosmos/cosmos-sdk/issues/3529).

#### Example
**Note** that "reserve pool" refers to the community pool. In this example from the [All in Bits website](../../validators/validator-faq.html#how-are-fees-distributed), there are 10 validators with equal stake. Each of them applies a 1% commission rate and has 20% of self-delegated Atoms. Now comes a successful block that collects a total of 1025.51020408 Atoms in fees.

First, a 2% tax is applied. The corresponding Atoms go to the reserve pool (aka community pool). Reserve pool's funds can be allocated through governance to fund bounties and upgrades.

2% * 1025.51020408 = 20.51020408 Atoms go to the reserve pool.
1005 Atoms now remain. Let's assume that the proposer included 100% of the signatures in its block. It thus obtains the full bonus of 5%.

We have to solve this simple equation to find the reward R for each validator:

9*R + R + R*5% = 1005 ⇔ R = 1005/10.05 = 100

For the proposer validator:

The pool obtains R + R * 5%: 105 Atoms

Commission: 105 * 80% * 1% = 0.84 Atoms

Validator's reward: 105 * 20% + Commission = 21.84 Atoms

Delegators' rewards: 105 * 80% - Commission = 83.16 Atoms (each delegator will be able to claim its portion of these rewards in proportion to their stake)

For each non-proposer validator:

The pool obtains R: 100 Atoms

Commission: 100 * 80% * 1% = 0.8 Atoms

Validator's reward: 100 * 20% + Commission = 20.8 Atoms

Delegators' rewards: 100 * 80% - Commission = 79.2 Atoms (each delegator will be able to claim their portion of these rewards in proportion to their stake)

### `withdrawaddrenabled`
**Determines whether or not delegators may set a separate address for receiving staking rewards.**

Delegators can designate a separate withdrawal address (account) that receives staking rewards when `withdrawaddrenabled` is set to `true`. When `withdrawaddrenabled` is set to `false`, the delegator can no longer designate a separate address for withdrawals.

* on-chain value: `{{ $themeConfig.currentParameters.distribution.withdrawaddrenabled }}`
* `cosmoshub-4` default: `true`
* `cosmoshub-3` default: `true`

#### Changing the `withdrawaddrenabled` parameter
Changing the `withdrawaddrenabled` to false will prevent delegators from changing or setting a separate withdrawal address (account) that receives the staking rewards. This may disrupt the functionality of applications and the expectations of staking participants.

#### Notes
This parameter was set to `false` before transfers were enabled in order to prevent stakers from diverting their rewards to other addresses ie. to avoid a loophole that would enable ATOM transfer via diverting staking rewards to a designated address. This parameter likely is only useful if [`sendenabled`](./Bank.md#1-sendenabled) is set to `false`.

## Known Bug
There is a known bug associated with this module that has reportedly caused a chain to halt. In [this reported case](https://github.com/cosmos/cosmos-sdk/issues/5808), the chain's parameter values were changed to be:
```
community_tax: "0.020000000000000000"
base_proposer_reward: "0.999000000000000000"
bonus_proposer_reward: "0.040000000000000000"
```

Though the system will not allow eg. `baseproposerreward` to be a value greater than 1.0, it will allow the [`communitytax`](#1-communitytax), [`baseproposerreward`](#2-baseproposerreward), and [`bonusproposerreward`](#3-bonusproposerreward) parameters values to total an amount greater than 1.00, which will apparently cause the chain to panic and halt. You can [read more about the reported issue here](https://github.com/cosmos/cosmos-sdk/issues/5808).
